package scanner

import (
	"charm.land/lipgloss/v2"
)

// scannedStyle emphasizes processed count.
func (m *Model) scannedStyle() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(m.state.Theme().Yellow)
}

func (m *Model) leftStyle() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(m.state.Theme().Info)
}

func (m *Model) foundStyle() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(m.state.Theme().Success)
}

func (m *Model) elapsedStyle() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(m.state.Theme().Purple)
}

func (m *Model) elapsedEndStyle() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(m.state.Theme().Orange)
}

func (m *Model) separatorStyle() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(m.state.Theme().Primary)
}
