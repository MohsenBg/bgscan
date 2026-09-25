package app

import (
	"charm.land/lipgloss/v2"
)

func (m *model) warningStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(m.state.Theme().Yellow).
		Bold(true).
		Padding(1, 2)
}

func (m *model) centerStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Align(lipgloss.Center, lipgloss.Center)
}
