package app

import (
	"bgscan/internal/ui/theme"

	"github.com/charmbracelet/lipgloss"
)

func containerStyle(termWidth, termHeight int) lipgloss.Style {
	return lipgloss.NewStyle().
		Align(lipgloss.Center, lipgloss.Center).
		Width(termWidth).
		Height(termHeight)
}

func mainStyle(contentWidth, contentHeight int) lipgloss.Style {
	return lipgloss.NewStyle().Align(lipgloss.Center, lipgloss.Center).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Current().BorderActive).
		Width(contentWidth).
		Height(contentHeight)
}

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
		Border(lipgloss.RoundedBorder()).MaxWidth(maxWidth).
		BorderForeground(theme.Current().BorderActive).Padding(0, 1)
}
