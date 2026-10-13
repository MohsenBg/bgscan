package form

import (
	"charm.land/lipgloss/v2"
)

func (m *Model) keyHintStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(m.env.Theme.Muted).
		Padding(1, 0)
}
