package notice

import (
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// LEVEL represents the severity level of a notice.
type LEVEL int

const (
	NOTICE_ERROR LEVEL = iota
	NOTICE_INFO
	NOTICE_SUCCESS
)

// Model is a notice dialog showing an informational, success, or error
// message inside a scrollable viewport.
type Model struct {
	id   ui.ComponentID
	name string

	env *ui.Env

	noticeType LEVEL
	message    string
	title      string

	viewport viewport.Model

	titleHeight    int
	viewportHeight int
	footerHeight   int

	// inDialog is true while the notice is hosted inside a dialog
	// frame: the title and the continue hint then live in the frame's
	// border lines instead of the body.
	inDialog bool
}

// New displays a titled message, scrolling when it exceeds the viewport
// height.
func New(env *ui.Env, title, message string, level LEVEL) *Model {
	v := viewport.New()

	m := &Model{
		id:         ui.NewComponentID(),
		name:       "Notice",
		env:        env,
		noticeType: level,
		message:    message,
		title:      title,
		viewport:   v,
	}

	wrapped := lipgloss.NewStyle().
		Width(m.Width()).
		Render(m.message)

	m.viewport.SetContent(wrapped)
	m.UpdateSize()

	return m
}

func (m *Model) Init() tea.Cmd {
	return m.viewport.Init()
}

func (m *Model) Name() string {
	return m.name
}

func (m *Model) ID() ui.ComponentID {
	return m.id
}

func (m *Model) OnClose() tea.Cmd {
	return nil
}

// Width is clamped to avoid excessively wide dialogs.
func (m *Model) Width() int {
	return min(50, m.env.Layout.Body().Width)
}

// Height is constrained so the notice never fills the screen.
func (m *Model) Height() int {
	return min(50, m.env.Layout.Body().Height)
}

// UpdateSize recomputes viewport dimensions from the current layout;
// call it after a resize.
func (m *Model) UpdateSize() {
	m.titleHeight = 0
	m.footerHeight = 0

	if !m.inDialog {
		m.titleHeight = lipgloss.Height(m.headerView(m.Width()))
		m.footerHeight = lipgloss.Height(m.footerView(m.Width()))
	}

	m.viewportHeight = max(
		m.Height()-m.titleHeight-m.footerHeight,
		1,
	)

	wrappedMsgHeight := lipgloss.Height(
		lipgloss.NewStyle().
			Width(m.Width()).
			Render(m.message),
	)

	m.viewportHeight = min(wrappedMsgHeight, m.viewportHeight)

	m.viewport.SetWidth(m.Width())
	m.viewport.SetHeight(m.viewportHeight)
}

func (m *Model) CloseCmd() tea.Cmd {
	return func() tea.Msg {
		return ui.CloseComponentMsg{ID: m.ID()}
	}
}

func (m *Model) Mode() ui.KeyMode {
	return ui.KeyModeNormal
}

// SetInDialog moves the notice's title and continue hint into the dialog
// frame's border lines.
func (m *Model) SetInDialog(inDialog bool) {
	m.inDialog = inDialog
	m.UpdateSize()
}
