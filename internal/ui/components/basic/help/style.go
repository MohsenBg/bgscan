package help

import (
	"charm.land/lipgloss/v2"
)

func (m *Model) keyStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(m.deps.Theme.Primary).
		Bold(true)
}

func (m *Model) labelStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(m.deps.Theme.Muted)
}

func (m *Model) sepStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(m.deps.Theme.Border)
}

func (m *Model) longHelpStyle() lipgloss.Style {
	return lipgloss.NewStyle().Padding(framePaddingVertical, framePaddingHorizontal).Align(lipgloss.Center)
}
