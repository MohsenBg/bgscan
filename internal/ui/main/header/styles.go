package header

import (
	"charm.land/lipgloss/v2"
)

func (m *Model) bannerStyle(width, height int) lipgloss.Style {
	return lipgloss.NewStyle().
		Align(lipgloss.Center, lipgloss.Bottom).
		Width(width).Height(height).
		Foreground(m.theme.Success).
		Bold(true)
}
