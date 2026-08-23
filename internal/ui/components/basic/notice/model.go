package notice

import (
	"bgscan/internal/ui/shared/env"
	"bgscan/internal/ui/shared/layout"
	"bgscan/internal/ui/shared/ui"

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

	layout *layout.Layout

	noticeType LEVEL
	message    string
	title      string

	viewport viewport.Model

	titleHeight    int
	viewportHeight int
	footerHeight   int
}

// New creates a new Notice component.
//
// The notice displays a titled message with optional scrolling if the
// message exceeds the available viewport height.
func New(layout *layout.Layout, title, message string, level LEVEL) *Model {
	v := viewport.New()

	m := &Model{
		id:         ui.NewComponentID(),
		name:       "Notice",
		layout:     layout,
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

// Init initializes the BubbleTea component.
func (m *Model) Init() tea.Cmd {
	return m.viewport.Init()
}

// Name returns the component name.
func (m *Model) Name() string {
	return m.name
}

// ID returns the unique component identifier.
func (m *Model) ID() ui.ComponentID {
	return m.id
}

// OnClose executes cleanup logic when the component is closed.
func (m *Model) OnClose() tea.Cmd {
	return nil
}

// Width returns the notice width.
//
// The width is clamped to avoid excessively wide dialogs.
func (m *Model) Width() int {
	if m.layout == nil {
		return 50
	}

	return min(50, m.layout.Body.Width)
}

// Height returns the notice height.
//
// The height is constrained to prevent the notice from occupying
// the entire screen.
func (m *Model) Height() int {
	if m.layout == nil {
		return 20
	}

	return min(50, m.layout.Body.Height)
}

// UpdateSize recalculates the internal layout dimensions.
//
// This method should be called when the terminal or layout size changes.
func (m *Model) UpdateSize() {
	m.titleHeight = lipgloss.Height(m.headerView(m.Width()))
	m.footerHeight = lipgloss.Height(m.footerView(m.Width()))

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

// CloseCmd returns a command that closes the notice component.
func (m *Model) CloseCmd() tea.Cmd {
	return func() tea.Msg {
		return ui.CloseComponentMsg{ID: m.ID()}
	}
}

// Mode returns the UI mode used by the notice dialog.
func (m *Model) Mode() env.Mode {
	return env.NormalMode
}
