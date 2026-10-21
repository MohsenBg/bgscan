package inspector

import (
	"charm.land/lipgloss/v2"
	"github.com/MohsenBg/bgscan/internal/ui/theme"
)

func fieldNameStyle(th *theme.Theme) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(th.Text).Padding(0, 1)
}

func selectedFieldNameStyle(th *theme.Theme) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(th.Primary).Bold(true)
}

func valueStyle(th *theme.Theme) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(th.Text)
}

func titleStyle(th *theme.Theme) lipgloss.Style {
	return lipgloss.NewStyle().Align(lipgloss.Center).Foreground(th.Info).Bold(true)
}

// paddingCell adds one line of top padding between rows.
func paddingCell() lipgloss.Style {
	return lipgloss.NewStyle().Padding(1, 0, 0, 0)
}
