package logview

import (
	"charm.land/lipgloss/v2"
)

func (m *Model) titleStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center, lipgloss.Center).
		Bold(true).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(m.state.Theme().BorderActive)
}

func ContainerStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center, lipgloss.Center)
}

// borderStyle wraps the log viewport. The border width is clamped to 80
// columns so the box stays readable on small terminals.
func (m *Model) borderStyle(width int) lipgloss.Style {
	width = min(80, width)
	return lipgloss.NewStyle().Padding(0, 1).
		Width(width).
		Align(lipgloss.Left).
		Border(lipgloss.RoundedBorder(), true, true, true, true).
		BorderForeground(m.state.Theme().Secondary)
}

func (m *Model) helpStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width-5).
		Foreground(m.state.Theme().Muted).
		Align(lipgloss.Center, lipgloss.Center).
		Padding(1)
}

func (m *Model) helpKeyStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(m.state.Theme().Secondary).
		Bold(true)
}

// scrollBarStyle returns the style for the scroll bar track.
func (m *Model) scrollBarStyle() lipgloss.Style {
	return lipgloss.NewStyle().Align(lipgloss.Center, lipgloss.Center).
		Foreground(m.state.Theme().Muted)
}

// scrollBarThumbStyle returns the style for the scroll bar thumb.
func (m *Model) scrollBarThumbStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(m.state.Theme().Primary)
}

// TitleStyle keeps backward compatibility.
func TitleStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center, lipgloss.Center).
		Bold(true).
		Border(lipgloss.NormalBorder(), false, false, true, false)
}

// BorderStyle keeps backward compatibility.
func BorderStyle(width int) lipgloss.Style {
	width = min(80, width)
	return lipgloss.NewStyle().Padding(0, 1).
		Width(width).
		Align(lipgloss.Left).
		Border(lipgloss.RoundedBorder(), true, true, true, true)
}

// ScrollBarStyle keeps backward compatibility.
func ScrollBarStyle() lipgloss.Style {
	return lipgloss.NewStyle().Align(lipgloss.Center, lipgloss.Center)
}

// ScrollBarThumbStyle keeps backward compatibility.
func ScrollBarThumbStyle() lipgloss.Style {
	return lipgloss.NewStyle()
}
