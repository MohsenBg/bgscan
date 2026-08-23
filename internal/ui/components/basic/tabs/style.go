package tabs

import (
	"bgscan/internal/ui/theme"

	"charm.land/lipgloss/v2"
)

// defaultActiveStyle returns the style used for the selected tab — a bold
// filled "pill" that pops against the muted tabs around it.
func defaultActiveStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Current().Text).
		Background(theme.Current().Purple).
		Padding(0, 2).
		MarginRight(1)
}

// defaultInactiveStyle returns the style used for unselected tabs — muted
// and unobtrusive so the active tab commands attention.
func defaultInactiveStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Faint(true).
		Foreground(theme.Current().Muted).
		Padding(0, 2).
		MarginRight(1)
}

// activeBorderStyle colors the stretch of the full-width border line that
// sits beneath the active tab.
func activeBorderStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Current().Purple)
}

// inactiveBorderStyle colors the rest of the full-width border line.
func inactiveBorderStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Faint(true).
		Foreground(theme.Current().Secondary)
}
