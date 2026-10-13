package app

import (
	"charm.land/lipgloss/v2"
	"github.com/MohsenBg/bgscan/internal/ui/theme"
)

func warningStyle(theme *theme.Theme) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Yellow).
		Bold(true).
		Padding(1, 2)
}

func centerStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Align(lipgloss.Center, lipgloss.Center)
}
