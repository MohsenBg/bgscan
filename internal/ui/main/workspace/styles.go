package workspace

import (
	"bgscan/internal/ui/theme"

	"charm.land/lipgloss/v2"
)

func containerStyle(termWidth, termHeight int) lipgloss.Style {
	return lipgloss.NewStyle().
		Align(lipgloss.Center, lipgloss.Center).
		Width(termWidth).
		Height(termHeight)
}

func mainStyle(contentWidth, contentHeight int) lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Current().BorderActive).
		Width(contentWidth).
		Height(contentHeight)
}

func WindowStyle(maxWidth int) lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		MaxWidth(maxWidth).
		BorderForeground(theme.Current().BorderActive).
		Padding(0, 1)
}
