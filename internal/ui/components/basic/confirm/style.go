package confirm

import (
	"charm.land/lipgloss/v2"
)

func (m *Model) messageStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).
		Bold(true).
		Foreground(m.theme.Text).
		Padding(1, 2).
		Align(lipgloss.Center)
}

// buttonStyle renders an unfocused confirm button ("No"/"Yes").
func (m *Model) buttonStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(m.theme.Muted).
		Padding(0, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.Border)
}

// selectButtonStyle renders the focused confirm button.
func (m *Model) selectButtonStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.Primary).
		Padding(0, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.BorderActive)
}

// MessageStyle keeps backward compatibility for external callers.
func MessageStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).
		Bold(true).
		Padding(1, 2).
		Align(lipgloss.Center)
}

// ButtonStyle renders an unfocused confirm button ("No"/"Yes").
func ButtonStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Padding(0, 2).
		Border(lipgloss.RoundedBorder())
}

// SelectButtonStyle renders the focused confirm button.
func SelectButtonStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Padding(0, 2).
		Border(lipgloss.RoundedBorder())
}
