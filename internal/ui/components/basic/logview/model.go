package logview

import (
	"context"
	"sync"
	"time"

	"github.com/MohsenBg/bgscan/internal/logger"
	"github.com/MohsenBg/bgscan/internal/ui/shared/env"
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

	state *ui.AppState

	padding           int
	containerMaxWidth int
	containerWidth    int
	showBorder        bool

	viewport viewport.Model

	logger     *logger.Logger
	loggerChan chan string
	maxMessage int

	mu         sync.Mutex
	messages   []string
	needUpdate bool

	ctx    context.Context
	cancel context.CancelFunc
}

// New creates a new log viewer component.
func New(state *ui.AppState, log *logger.Logger, title string) *Model {
	ctx, cancel := context.WithCancel(context.Background())

	m := &Model{
		id:         ui.NewComponentID(),
		name:       title,
		title:      title,
		state:      state,
		logger:     log,
		maxMessage: 200,

		viewport: viewport.New(),

		padding:           5,
		containerMaxWidth: state.Layout.Body.Width - 10,
		showBorder:        true,

		ctx:    ctx,
		cancel: cancel,
	}

	m.setSize()

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
	go m.readLogs()
	return m.tick()
}

// readLogs listens for new log messages from the logger.
func (m *Model) readLogs() {
	m.loggerChan = m.logger.Subscribe(200, m.maxMessage)

	for {
		select {

		case <-m.ctx.Done():
			return

		case logMsg, ok := <-m.loggerChan:
			if !ok {
				return
			}

			m.mu.Lock()

			m.messages = append(m.messages, logMsg)

			if len(m.messages) > m.maxMessage {
				m.messages = m.messages[len(m.messages)-m.maxMessage:]
			}

			m.needUpdate = true

			m.mu.Unlock()
		}
	}
}

// tick schedules periodic UI refreshes.
func (m *Model) tick() tea.Cmd {
	return tea.Tick(
		m.state.Config.General.StatusInterval.Duration(),
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
		m.state.Layout.Body.Width-10,
	)

	width := min(maxViewportWidth, m.containerWidth-2)

	m.viewport.SetWidth(width)
	helpHeight := lipgloss.Height(
		helpStyle(m.viewport.Width()).Render(helpView()),
	)

	height := m.state.Layout.Body.Height -
		m.padding -
		lipgloss.Height(m.title) -
		helpHeight
	m.viewport.SetHeight(height)
}

// ID returns the component identifier.
func (m *Model) ID() ui.ComponentID {
	return m.id
}

// Name returns the component name.
func (m *Model) Name() string {
	return m.name
}

// Mode defines the input mode used by this component.
func (m *Model) Mode() env.Mode {
	return env.NormalMode
}

// OnClose cleans up resources when the component is removed.
func (m *Model) OnClose() tea.Cmd {
	m.cancel()

	if m.loggerChan != nil {
		m.logger.Unsubscribe(m.loggerChan)
	}

	return nil
}
