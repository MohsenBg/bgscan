package header

import (
	"bgscan/internal/ui/theme"

	"charm.land/lipgloss/v2"
)

func bannerStyle(width, height int) lipgloss.Style {
	return lipgloss.NewStyle().
		Align(lipgloss.Center, lipgloss.Bottom).
		Width(width).Height(height).
		Foreground(theme.Current().Success).
		Bold(true)
}
