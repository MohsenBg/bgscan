package inspector

import (
	"github.com/MohsenBg/bgscan/internal/ui/theme"

	"charm.land/lipgloss/v2"
)

func fieldNameStyle(th *theme.Theme) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(th.Text).
		Padding(0, 1)
}

func selectedFieldNameStyle(th *theme.Theme) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(th.Primary).
		Padding(0, 0).
		Bold(true)
}

func valueStyle(th *theme.Theme) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(th.Text)
}

func (m *Model) titleStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Align(lipgloss.Center).
		Foreground(m.theme.Info).
		Bold(true)
}

// PaddingCell adds one line of top padding between rows.
func PaddingCell() lipgloss.Style {
	return lipgloss.NewStyle().
		Padding(1, 0, 0, 0)
}
