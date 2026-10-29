package logview

import (
	"context"
	"sync"
	"time"

	"github.com/MohsenBg/bgscan/internal/logger"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/help"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const (
	// max reading line
	maxLine = 2000

	// maxBuffer is how many log lines the view keeps.
	maxBuffer = 10000

	// flushInterval is how long the reader collects lines before handing them to the UI.
	flushInterval = 500 * time.Millisecond
)

// logEntry is one buffered log line with the rows it renders to.
type logEntry struct {
	raw  string
	rows []string
}

// Model is a scrollable log viewer. It subscribes to the application logger,
type Model struct {
	id    ui.ComponentID
	name  string
	title string

	env *ui.Env

	padding           int
	containerMaxWidth int
	containerWidth    int
	showBorder        bool

	viewport viewport.Model
	help     *help.Model

	logger *logger.Logger

	subOnce sync.Once
	subMu   sync.Mutex

	loggerChan chan string
	closed     bool

	entries []logEntry
	lines   []string

	renderWidth int
	droppedRows int
	appliedRows int

	ctx    context.Context
	cancel context.CancelFunc
}

func New(env *ui.Env, log *logger.Logger, title string) *Model {
	ctx, cancel := context.WithCancel(context.Background())

	m := &Model{
		id:     ui.NewComponentID(),
		name:   title,
		title:  title,
		env:    env,
		logger: log,

		viewport: viewport.New(),
		help:     help.New(env, title, helpSet(env.Keys)),

		padding:           10,
		containerMaxWidth: env.Layout.Body().Width - 10,
		showBorder:        true,

		ctx:    ctx,
		cancel: cancel,
	}

	m.viewport.KeyMap = viewportKeyMap(env.Keys.LogView)
	m.setSize()
	m.viewport.GotoBottom()

	return m
}

// SetContainerWidth limits the maximum width of the log container.
func (m *Model) SetContainerWidth(width int) {
	m.containerMaxWidth = width
	m.setSize()
}

// SetShowBorder enables or disables the container border.
func (m *Model) SetShowBorder(border bool) {
	m.showBorder = border
	m.setSize()
}

// Init starts the log subscription and the background reader.
func (m *Model) Init() tea.Cmd {
	return func() tea.Msg {
		m.subOnce.Do(func() {
			ch := m.logger.Subscribe(maxBuffer, maxLine)

			m.subMu.Lock()
			if m.closed {
				m.subMu.Unlock()
				m.logger.Unsubscribe(ch)
				return
			}
			m.loggerChan = ch
			m.subMu.Unlock()

			go m.readLogs(ch)
		})
		return nil
	}
}

// readLogs collects log lines and hands them to the UI as one message per flush interval.
func (m *Model) readLogs(ch <-chan string) {
	pending := make([]string, 0, maxBuffer)

	timer := time.NewTimer(flushInterval)
	if !timer.Stop() {
		<-timer.C
	}
	defer timer.Stop()

	var armed <-chan time.Time

	send := func() {
		batch := pending
		pending = make([]string, 0, maxBuffer)

		if m.env.Program != nil {
			m.env.Program.Send(LogAppendMsg{Lines: batch})
		}
	}

	for {
		select {

		case <-m.ctx.Done():
			return

		case <-armed:
			armed = nil
			if len(pending) > 0 {
				send()
			}

		case line, ok := <-ch:
			if !ok {
				if len(pending) > 0 {
					send()
				}
				return
			}

			pending = append(pending, line)
			if len(pending) > maxBuffer {
				pending = pending[len(pending)-maxBuffer:]
			}

			if armed == nil {
				timer.Reset(flushInterval)
				armed = timer.C
			}
		}
	}
}

// setSize recalculates the viewport, the help bar and the container size.
func (m *Model) setSize() {
	const maxViewportWidth = 80

	m.containerWidth = min(
		m.containerMaxWidth,
		m.env.Layout.Body().Width-10,
	)

	m.viewport.SetWidth(min(maxViewportWidth, m.containerWidth-2))
	m.help.SetShortWidth(max(1, m.viewport.Width()-5))

	helpHeight := lipgloss.Height(
		m.helpStyle(m.viewport.Width()).Render(m.help.View()),
	)

	vHeght := m.env.Layout.Body().Height -
		m.padding -
		lipgloss.Height(m.title) -
		helpHeight

	m.viewport.SetHeight(max(vHeght, 20))

	m.pushRows()
}

// contentWidth is the text width a log line is wrapped to.
func (m *Model) contentWidth() int {
	return max(1, m.viewport.Width()-5)
}

// appendLines adds a batch of raw lines to the buffer. Rendering waits until
// pushRows, so a batch costs one pass instead of one per line.
func (m *Model) appendLines(lines []string) {
	for _, raw := range lines {
		m.entries = append(m.entries, logEntry{raw: raw})
	}
}

// rebuildRows renders the lines that have no rows yet, drops the oldest
func (m *Model) rebuildRows() (lines []string, dropped int) {
	width := m.contentWidth()
	rewrap := width != m.renderWidth
	m.renderWidth = width

	for i := range m.entries {
		if rewrap || m.entries[i].rows == nil {
			m.entries[i].rows = renderRows(&m.env.Theme, m.entries[i].raw, width)
		}
	}

	for len(m.entries) > maxBuffer {
		m.droppedRows += len(m.entries[0].rows)
		m.entries = m.entries[1:]
	}

	m.lines = m.lines[:0]
	for i := range m.entries {
		m.lines = append(m.lines, m.entries[i].rows...)
	}

	dropped = m.droppedRows - m.appliedRows
	m.appliedRows = m.droppedRows

	if rewrap {
		dropped = 0
	}

	return m.lines, dropped
}

// pushRows rebuilds the rows and hands them to the viewport. The scroll
// position is kept unless the view was pinned to the bottom.
func (m *Model) pushRows() {
	lines, dropped := m.rebuildRows()

	atBottom := m.viewport.AtBottom()
	m.viewport.SetContentLines(lines)

	switch {
	case atBottom:
		m.viewport.GotoBottom()
	case dropped > 0:
		m.viewport.SetYOffset(m.viewport.YOffset() - dropped)
	}
}

func (m *Model) ID() ui.ComponentID {
	return m.id
}

func (m *Model) Name() string {
	return m.name
}

// Mode defines the input mode used by this component.
func (m *Model) Mode() ui.KeyMode {
	return ui.KeyModeNormal
}

// OnClose cleans up resources when the component is removed.
func (m *Model) OnClose() tea.Cmd {
	m.cancel()

	m.subMu.Lock()
	ch := m.loggerChan
	m.loggerChan = nil
	m.closed = true
	m.subMu.Unlock()

	if ch != nil {
		m.logger.Unsubscribe(ch)
	}

	return nil
}
