package confirm

import (
	"bgscan/internal/ui/theme"

	"charm.land/lipgloss/v2"
)

func MessageStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).
		Bold(true).
		Foreground(theme.Current().Text).
		Padding(1, 2).
		Align(lipgloss.Center)
}

// ButtonStyle renders an unfocused confirm button ("No"/"Yes").
func ButtonStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Muted).
		Padding(0, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Current().Border)
}

// SelectButtonStyle renders the focused confirm button.
func SelectButtonStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Current().Primary).
		Padding(0, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Current().BorderActive)
}
