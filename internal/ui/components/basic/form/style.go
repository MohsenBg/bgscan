package form

import (
	"bgscan/internal/ui/theme"

	"charm.land/lipgloss/v2"
)

func keyHintStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Muted).
		Padding(1, 0)
}
