package body

import "charm.land/lipgloss/v2"

func containerStyle(width, height int) lipgloss.Style {
	return lipgloss.NewStyle().
		Align(lipgloss.Center, lipgloss.Center).
		Width(width).
		Height(height)
}
