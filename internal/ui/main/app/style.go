package app

import (
	"github.com/MohsenBg/bgscan/internal/ui/theme"

	"charm.land/lipgloss/v2"
)

func warningStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Yellow).
		Bold(true).
		Padding(1, 2)
}

func centerStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Align(lipgloss.Center, lipgloss.Center)
}

func WindowStyle(maxWidth int) lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		MaxWidth(maxWidth).
		BorderForeground(theme.Current().BorderActive).
		Padding(0, 1)
}
