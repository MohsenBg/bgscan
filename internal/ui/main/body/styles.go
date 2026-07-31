package body

import (
	"charm.land/lipgloss/v2"
)

// containerStyle centers content within the given dimensions.
func containerStyle(width, height int) lipgloss.Style {
	return lipgloss.NewStyle().Align(lipgloss.Center, lipgloss.Center).
		Width(width).Height(height)
}
