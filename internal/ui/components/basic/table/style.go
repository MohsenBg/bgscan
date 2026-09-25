package table

import (
	"charm.land/bubbles/v2/table"
	"charm.land/lipgloss/v2"
)

func (m *Model) tableStyles() table.Styles {
	s := table.DefaultStyles()

	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(m.theme.Info).
		BorderBottom(true).
		Padding(0, 1)

	s.Cell = s.Cell.Padding(0, 1)

	s.Selected = s.Selected.
		Foreground(m.theme.Text).
		Background(m.theme.Selected).
		Height(1).
		Bold(true)

	return s
}

func (m *Model) titleStyles(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center).
		Foreground(m.theme.Info).
		Bold(true).
		Padding(1, 0)
}

func (m *Model) filterStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Align(lipgloss.Left).
		Foreground(m.theme.Info).
		Padding(0, 1)
}

func (m *Model) tableViewStyle(_ int) lipgloss.Style {
	return lipgloss.NewStyle().
		Align(lipgloss.Left).
		Foreground(m.theme.Secondary).
		Padding(0, 1, 0, 1)
}

func (m *Model) helpViewStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width-2).
		Align(lipgloss.Center).
		Padding(0, 1).
		Foreground(m.theme.Secondary).
		MarginTop(1)
}
