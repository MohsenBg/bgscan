package scanner

import (
	"bgscan/internal/ui/theme"

	"github.com/charmbracelet/lipgloss"
)

func scannedStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Yellow)
}

func leftStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Info)
}

func foundStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Success)
}

func elapsedStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Purple)
}

func elapsedEndStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Orange)
}

func separatorStyle() lipgloss.Style { return lipgloss.NewStyle().Foreground(theme.Current().Primary) }
