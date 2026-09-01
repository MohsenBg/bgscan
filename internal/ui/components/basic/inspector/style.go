package inspector

import (
	"github.com/MohsenBg/bgscan/internal/ui/theme"

	"charm.land/lipgloss/v2"
)

func fieldNameStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Text).
		Padding(0, 1)
}

func selectedFieldNameStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Primary).
		Padding(0, 0).
		Bold(true)
}

func valueStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Text)
}

func titleStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Align(lipgloss.Center).
		Foreground(theme.Current().Info).
		Bold(true)
}

// PaddingCell adds one line of top padding between rows.
func PaddingCell() lipgloss.Style {
	return lipgloss.NewStyle().
		Padding(1, 0, 0, 0)
}
