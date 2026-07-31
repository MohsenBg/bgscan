package picker

import (
	"bgscan/internal/ui/shared/layout"
	"bgscan/internal/ui/theme"

	"charm.land/lipgloss/v2"
)

// pickerHeight returns the vertical space available for the picker overlay.
func pickerHeight(layout *layout.Layout) int {
	return layout.Body.Height - 10
}

func containerStyle(width, height int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		Align(lipgloss.Left, lipgloss.Top).
		Padding(0, 1).
		Margin(1, 0)
}

func TitleStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width-5).
		Align(lipgloss.Center).
		Bold(true).
		Foreground(theme.Current().Info).
		Padding(0, 0, 2, 0).
		BorderForeground(lipgloss.Color("240"))
}

func currentDirStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width-5).
		Align(lipgloss.Left).
		Bold(true).
		Foreground(theme.Current().Yellow).
		Border(lipgloss.NormalBorder(), false, false, true, false)
}

func helpStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width - 5).
		Foreground(theme.Current().Muted).
		Align(lipgloss.Center).
		PaddingTop(1)
}

func helpKeyStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Secondary).
		Bold(true)
}
