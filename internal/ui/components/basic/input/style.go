package input

import (
	"bgscan/internal/ui/theme"

	"charm.land/lipgloss/v2"
)

func ContainerStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().Width(width)
}

func MessageStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Text).
		Bold(true).
		MarginBottom(1)
}

func ErrorStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Error).
		Bold(true).
		MarginTop(1)
}

func KeyHintStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Muted).
		MarginTop(1)
}
