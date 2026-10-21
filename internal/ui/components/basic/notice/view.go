package notice

import (
	"charm.land/lipgloss/v2"
)

// View renders the notice component.
//
// Layout structure:
//
//	Header
//	Message viewport
func (m *Model) View() string {
	width := m.Width()
	if width <= 0 {
		return ""
	}

	// Wrap message content to component width
	wrapped := lipgloss.NewStyle().
		Width(width).
		Render(m.message)

	m.viewport.SetContent(wrapped)

	content := lipgloss.JoinVertical(
		lipgloss.Top,
		m.viewport.View(),
	)

	return containerStyle(width).Render(content)
}
