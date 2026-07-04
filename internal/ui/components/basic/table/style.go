package table

import (
	"bgscan/internal/ui/theme"

	"charm.land/bubbles/v2/table"
	"charm.land/lipgloss/v2"
)

func tableStyles() table.Styles {
	s := table.DefaultStyles()

	// Header styling
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(theme.Current().Info).
		BorderBottom(true).
		Padding(0, 1)

	s.Cell = s.Cell.Padding(0, 1)

	s.Selected = s.Selected.
		Foreground(theme.Current().Text).
		Background(theme.Current().Purple).
		Height(1).
		Bold(true)

	return s
}

func titleStyles(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center).
		Foreground(theme.Current().Info).
		Bold(true).
		Padding(1, 0)
}

func tableViewStyle(_ int) lipgloss.Style {
	return lipgloss.NewStyle().
		Align(lipgloss.Left).
		Foreground(theme.Current().Secondary).
		Padding(0, 1, 0, 1)
}

// helpViewStyle now includes a top border to visually separate it from the table
func helpViewStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width-2).
		Align(lipgloss.Center).
		Padding(0, 1).
		Foreground(theme.Current().Secondary).
		MarginTop(1)
}
