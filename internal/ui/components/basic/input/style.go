package input

import (
	"github.com/MohsenBg/bgscan/internal/ui/theme"

	"charm.land/lipgloss/v2"
)

func ContainerStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().Width(width)
}

func MessageStyle(th *theme.Theme) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(th.Text).
		Bold(true).
		MarginBottom(1)
}

func ErrorStyle(th *theme.Theme) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(th.Error).
		Bold(true).
		MarginTop(1)
}

func KeyHintStyle(th *theme.Theme) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(th.Muted).
		MarginTop(1)
}
