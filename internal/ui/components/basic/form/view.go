package form

import (
	"charm.land/lipgloss/v2"
)

func (m *Model) View() string {
	title := m.renderTitle()

	if m.inspector == nil {
		return title
	}

	body := lipgloss.NewStyle().
		Padding(0, 2).
		Render(m.inspector.View())

	return lipgloss.JoinVertical(lipgloss.Center, title, body)
}

func (m *Model) renderTitle() string {
	return lipgloss.NewStyle().
		Foreground(m.env.Theme.Primary).
		Bold(true).
		Padding(0, 0, 1, 0).
		Align(lipgloss.Center).
		Width(m.width).
		Render(m.name)
}
