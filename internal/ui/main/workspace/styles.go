package workspace

import (
	"charm.land/lipgloss/v2"
)

// screenStyle centers the app within the terminal.
func screenStyle(width, height int) lipgloss.Style {
	return lipgloss.NewStyle().
		Align(lipgloss.Center, lipgloss.Center).
		Width(width).
		Height(height)
}

// contentStyle returns the style for the main layout frame.
func (m *model) contentStyle(width, height int) lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.session.Env().Theme.BorderActive).
		Width(width).
		Height(height)
}
