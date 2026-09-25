package tabs

import (
	"charm.land/lipgloss/v2"
)

// defaultActiveStyle returns the style used for the selected tab — a bold
// filled "pill" that pops against the muted tabs around it.
func (m *Model[T]) defaultActiveStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.Text).
		Background(m.theme.Selected).
		Padding(0, 2).
		MarginRight(1)
}

// defaultInactiveStyle returns the style used for unselected tabs — muted
// and unobtrusive so the active tab commands attention.
func (m *Model[T]) defaultInactiveStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Faint(true).
		Foreground(m.theme.Muted).
		Padding(0, 2).
		MarginRight(1)
}

// activeBorderStyle colors the stretch of the full-width border line that
// sits beneath the active tab.
func (m *Model[T]) activeBorderStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.Purple)
}

// inactiveBorderStyle colors the rest of the full-width border line.
func (m *Model[T]) inactiveBorderStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Faint(true).
		Foreground(m.theme.Secondary)
}
