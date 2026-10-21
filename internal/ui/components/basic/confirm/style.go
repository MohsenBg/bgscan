package confirm

import (
	"charm.land/lipgloss/v2"
)

func (m *Model) messageStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).
		Bold(true).
		Foreground(m.env.Theme.Text).
		Padding(0, 2, 1, 2).
		Align(lipgloss.Center)
}

// buttonStyle renders an unfocused confirm button ("No"/"Yes").
func (m *Model) buttonStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(m.env.Theme.Muted).
		Padding(0, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.env.Theme.Border)
}

// selectButtonStyle renders the focused confirm button.
func (m *Model) selectButtonStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(m.env.Theme.Primary).
		Padding(0, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.env.Theme.BorderActive)
}
