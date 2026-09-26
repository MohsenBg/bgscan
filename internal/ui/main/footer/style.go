package footer

import (
	"charm.land/lipgloss/v2"
)

func (m *Model) containerStyle(width, height int) lipgloss.Style {
	t := m.theme
	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		Foreground(t.Text)
}

func (m *Model) separatorStyle(width int) lipgloss.Style {
	t := m.theme
	return lipgloss.NewStyle().
		Width(width).
		Foreground(t.BorderActive)
}

func leftSectionStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).
		Padding(0, 1).
		Align(lipgloss.Left)
}

func centerSectionStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).
		Padding(0, 1).
		Align(lipgloss.Center)
}

func rightSectionStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).
		Padding(0, 1).
		Align(lipgloss.Right)
}

func (m *Model) appNameStyle() lipgloss.Style {
	t := m.theme
	return lipgloss.NewStyle().
		Foreground(t.Success).
		Bold(true)
}

func (m *Model) versionStyle() lipgloss.Style {
	t := m.theme
	return lipgloss.NewStyle().
		Foreground(t.Muted)
}

func (m *Model) statusTextStyle() lipgloss.Style {
	t := m.theme
	return lipgloss.NewStyle().
		Foreground(t.Info).
		Bold(true)
}

func (m *Model) statsStyle() lipgloss.Style {
	t := m.theme
	return lipgloss.NewStyle().
		Foreground(t.Secondary)
}

func (m *Model) statsValueStyle() lipgloss.Style {
	t := m.theme
	return lipgloss.NewStyle().
		Foreground(t.Text).
		Bold(true)
}
