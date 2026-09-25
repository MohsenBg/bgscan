package workspace

import (
	"charm.land/lipgloss/v2"
)

func containerStyle(termWidth, termHeight int) lipgloss.Style {
	return lipgloss.NewStyle().
		Align(lipgloss.Center, lipgloss.Center).
		Width(termWidth).
		Height(termHeight)
}

func (m *model) mainStyle(contentWidth, contentHeight int) lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.state.Theme().BorderActive).
		Width(contentWidth).
		Height(contentHeight)
}

func (m *model) windowStyle(maxWidth int) lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		MaxWidth(maxWidth).
		BorderForeground(m.state.Theme().BorderActive).
		Padding(0, 1)
}
