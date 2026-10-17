package tabs

import (
	"charm.land/lipgloss/v2"
)

func (m *Model[T]) defaultActiveStyle() lipgloss.Style {
	if m.animated {
		return lipgloss.NewStyle().
			Bold(true).
			Foreground(m.env.Theme.Primary)
	}

	return lipgloss.NewStyle().
		Bold(true).
		Foreground(m.env.Theme.Text).
		Background(m.env.Theme.Selected).
		Padding(0, 2).
		MarginRight(1)
}

func (m *Model[T]) defaultInactiveStyle() lipgloss.Style {
	if m.animated {
		return lipgloss.NewStyle().
			Foreground(m.env.Theme.Muted)
	}

	return lipgloss.NewStyle().
		Faint(true).
		Foreground(m.env.Theme.Muted).
		Padding(0, 2).
		MarginRight(1)
}

func (m *Model[T]) activeBorderStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(m.env.Theme.Purple)
}

func (m *Model[T]) inactiveBorderStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Faint(true).
		Foreground(m.env.Theme.Secondary)
}
