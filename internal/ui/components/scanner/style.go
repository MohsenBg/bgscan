package scanner

import (
	"github.com/MohsenBg/bgscan/internal/ui/theme"

	"charm.land/lipgloss/v2"
)

// scannedStyle emphasizes processed count.
func scannedStyle() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(theme.Current().Yellow)
}

func leftStyle() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(theme.Current().Info)
}

func foundStyle() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(theme.Current().Success)
}

func elapsedStyle() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(theme.Current().Purple)
}

func elapsedEndStyle() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(theme.Current().Orange)
}

func separatorStyle() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(theme.Current().Primary)
}
