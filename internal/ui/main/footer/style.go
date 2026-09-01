package footer

import (
	"github.com/MohsenBg/bgscan/internal/ui/theme"

	"charm.land/lipgloss/v2"
)

func containerStyle(width, height int) lipgloss.Style {
	t := theme.Current()
	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		Foreground(t.Text)
}

func separatorStyle(width int) lipgloss.Style {
	t := theme.Current()
	return lipgloss.NewStyle().
		Width(width).
		Foreground(t.Border)
}

func leftSectionStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).
		Padding(0, 1).
		Align(lipgloss.Left)
}

func centerSectionStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).
		Padding(0, 1).
		Align(lipgloss.Center)
}

func rightSectionStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).
		Padding(0, 1).
		Align(lipgloss.Right)
}

func appNameStyle() lipgloss.Style {
	t := theme.Current()
	return lipgloss.NewStyle().
		Foreground(t.Yellow).
		Bold(true)
}

func versionStyle() lipgloss.Style {
	t := theme.Current()
	return lipgloss.NewStyle().
		Foreground(t.Success).
		Faint(true)
}

func statusTextStyle() lipgloss.Style {
	t := theme.Current()
	return lipgloss.NewStyle().
		Foreground(t.Primary).
		Bold(true)
}

func iconStyle() lipgloss.Style {
	t := theme.Current()
	return lipgloss.NewStyle().
		Foreground(t.Orange)
}
