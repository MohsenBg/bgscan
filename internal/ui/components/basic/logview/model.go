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

// Model is a scrollable log viewer. It subscribes to the application logger
// and streams messages into a viewport, buffering them and refreshing
// periodically to avoid excessive UI updates.
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

	logger     *logger.Logger
	maxMessage int

	subOnce sync.Once
	subMu   sync.Mutex

	loggerChan chan string
	closed     bool

	mu sync.Mutex

	entries     []logEntry
	lines       []string
	renderWidth int
	trimmedRows int
	appliedTrim int
	needUpdate  bool

	ctx    context.Context
	cancel context.CancelFunc
}

func New(env *ui.Env, log *logger.Logger, title string) *Model {
	ctx, cancel := context.WithCancel(context.Background())

	m := &Model{
		id:         ui.NewComponentID(),
		name:       title,
		title:      title,
		env:        env,
		logger:     log,
		maxMessage: 200,

		viewport: viewport.New(),
		help:     help.New(env, title, helpSet(env.Keys)),

		padding:           5,
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

// Init starts the log subscription and background reader.
func (m *Model) Init() tea.Cmd {
	m.subOnce.Do(func() {
		ch := m.logger.Subscribe(200, m.maxMessage)

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
	return m.tick()
}

// readLogs listens for new log messages from the logger.
func (m *Model) readLogs(ch chan string) {
	for {
		select {

		case <-m.ctx.Done():
			return

		case logMsg, ok := <-ch:
			if !ok {
				return
			}

			m.mu.Lock()
			m.appendLocked(logMsg)
			m.mu.Unlock()
		}
	}
}

// tickInterval is the cadence of viewport refreshes draining the
// background log buffer into the UI.
const tickInterval = time.Second

// tick schedules periodic UI refreshes.
func (m *Model) tick() tea.Cmd {
	return tea.Tick(
		tickInterval,
		func(time.Time) tea.Msg {
			return LogUpdateTickMsg{}
		},
	)
}

// setSize recalculates the viewport and container dimensions.
func (m *Model) setSize() {
	maxViewportWidth := 80

	m.containerWidth = min(
		m.containerMaxWidth,
		m.env.Layout.Body().Width-10,
	)

	width := min(maxViewportWidth, m.containerWidth-2)

	m.viewport.SetWidth(width)
	m.help.SetShortWidth(max(1, m.viewport.Width()-5))
	helpHeight := lipgloss.Height(
		m.helpStyle(m.viewport.Width()).Render(m.help.View()),
	)

	height := m.env.Layout.Body().Height -
		m.padding -
		lipgloss.Height(m.title) -
		helpHeight
	m.viewport.SetHeight(height)

	m.pushContent()
}

// contentWidth is the text width a log line is wrapped to.
func (m *Model) contentWidth() int {
	return max(1, m.viewport.Width()-5)
}

// appendLocked buffers one raw log line. It runs on the reader goroutine, so
// it only stores the text; rendering stays on the UI goroutine.
func (m *Model) appendLocked(raw string) {
	m.entries = append(m.entries, logEntry{raw: raw})

	for len(m.entries) > m.maxMessage {
		if rows := m.entries[0].rows; len(rows) > 0 {
			m.trimmedRows += len(rows)
		} else {
			m.trimmedRows++
		}
		m.entries = m.entries[1:]
	}

	m.needUpdate = true
}

// refreshLocked renders entries that have no rows yet and rebuilds the flat
// row list. It returns the rows for the viewport and how many rows were
// dropped from the front since the previous refresh.
func (m *Model) refreshLocked() (lines []string, drop int) {
	rebuild := false

	if w := m.contentWidth(); w != m.renderWidth {
		m.renderWidth = w
		rebuild = true
	}

	dirty := rebuild
	for i := range m.entries {
		if m.entries[i].rows == nil {
			m.entries[i].rows = renderRows(&m.env.Theme, m.entries[i].raw, m.renderWidth)
			dirty = true
		}
	}

	if !dirty {
		return m.lines, 0
	}

	m.lines = m.lines[:0]
	for i := range m.entries {
		m.lines = append(m.lines, m.entries[i].rows...)
	}

	if rebuild {
		m.appliedTrim = m.trimmedRows
		return m.lines, 0
	}

	drop = m.trimmedRows - m.appliedTrim
	m.appliedTrim = m.trimmedRows

	return m.lines, drop
}

// pushContent hands the cached rows to the viewport, re-rendering the buffer
// first when the width changed. The scroll position is kept unless the view
// was pinned to the bottom.
func (m *Model) pushContent() {
	atBottom := m.viewport.AtBottom()
	offset := m.viewport.YOffset()

	m.mu.Lock()
	lines, _ := m.refreshLocked()
	m.mu.Unlock()

	m.viewport.SetContentLines(lines)

	if atBottom {
		m.viewport.GotoBottom()
		return
	}

	m.viewport.SetYOffset(offset)
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

// messageCount reports how many messages are buffered.
func (m *Model) messageCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.entries)
}
