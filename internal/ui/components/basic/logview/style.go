package logview

import (
	"bgscan/internal/ui/theme"

	"charm.land/lipgloss/v2"
)

func TitleStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center, lipgloss.Center).
		Bold(true).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(theme.Current().BorderActive)
}

func ContainerStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center, lipgloss.Center)
}

// BorderStyle wraps the log viewport. The border width is clamped to 80
// columns so the box stays readable on small terminals.
func BorderStyle(width int) lipgloss.Style {
	width = min(80, width)
	return lipgloss.NewStyle().Padding(0, 1).
		Width(width).
		Align(lipgloss.Left).
		Border(lipgloss.RoundedBorder(), true, true, true, true).
		BorderForeground(theme.Current().Secondary)
}

func helpStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width-5).
		Foreground(theme.Current().Muted).
		Align(lipgloss.Center, lipgloss.Center).
		Padding(1)
}

func helpKeyStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Secondary).
		Bold(true)
}

// ScrollBarStyle returns the style for the scroll bar track.
func ScrollBarStyle() lipgloss.Style {
	return lipgloss.NewStyle().Align(lipgloss.Center, lipgloss.Center).
		Foreground(theme.Current().Muted)
}

// ScrollBarThumbStyle returns the style for the scroll bar thumb.
func ScrollBarThumbStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Primary)
}
