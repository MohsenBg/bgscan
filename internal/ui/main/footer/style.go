package footer

import (
	"charm.land/lipgloss/v2"
	"github.com/MohsenBg/bgscan/internal/ui/theme"
)

func containerStyle(t *theme.Theme, width, height int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		Foreground(t.Text)
}

func separatorStyle(t *theme.Theme, width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).
		Foreground(t.BorderActive)
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

func appNameStyle(t *theme.Theme) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.Success).
		Bold(true)
}

func versionStyle(t *theme.Theme) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.Muted)
}

func statusTextStyle(t *theme.Theme) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.Info).
		Bold(true)
}

func statsStyle(t *theme.Theme) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.Secondary)
}

func statsValueStyle(t *theme.Theme) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.Text).
		Bold(true)
}
