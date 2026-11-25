// Package notice provides a small message dialog for errors, info,
// and success feedback. NewNoticeCmd opens the dialog with a frame
// style that matches the level. The message scrolls when it is
// taller than the viewport, and the continue key closes the dialog.
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
	// NOTICE_ERROR marks an error notice. It uses the danger dialog frame.
	NOTICE_ERROR LEVEL = iota
	// NOTICE_INFO marks an information notice. It uses the info dialog frame.
	NOTICE_INFO
	// NOTICE_SUCCESS marks a success notice. It uses the success dialog frame.
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

	viewport viewport.Model

	viewportHeight int
}

// new displays a titled message, scrolling when it exceeds the viewport
// height.
func new(env *ui.Env, message string, level LEVEL) *Model {
	v := viewport.New()

	m := &Model{
		id:         ui.NewComponentID(),
		name:       "Notice",
		env:        env,
		noticeType: level,
		message:    message,
		viewport:   v,
	}

	wrapped := lipgloss.NewStyle().
		Width(m.Width()).
		Render(m.message)

	m.viewport.SetContent(wrapped)
	m.UpdateSize()

	return m
}

// Init starts the inner viewport.
func (m *Model) Init() tea.Cmd {
	return m.viewport.Init()
}

// Name returns the component name.
func (m *Model) Name() string {
	return m.name
}

// ID returns the component ID.
func (m *Model) ID() ui.ComponentID {
	return m.id
}

// OnClose returns no command. The notice holds no resources to free.
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
	m.viewportHeight = max(
		m.Height(),
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

// CloseCmd returns a command that closes the component.
func (m *Model) CloseCmd() tea.Cmd {
	return func() tea.Msg {
		return ui.CloseComponentMsg{ID: m.ID()}
	}
}

// Mode returns the key mode the component uses.
func (m *Model) Mode() ui.KeyMode {
	return ui.KeyModeNormal
}
