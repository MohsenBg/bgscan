// style.go
package input

import (
	"charm.land/lipgloss/v2"

	"github.com/MohsenBg/bgscan/internal/ui/theme"
)

// ContainerStyle returns a base style with a fixed width.
func ContainerStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().Width(width)
}

// MessageStyle formats the primary prompt or instruction text.
func MessageStyle(th *theme.Theme) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(th.Text).
		Bold(true)
}

// ErrorStyle formats validation error messages.
func ErrorStyle(th *theme.Theme) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(th.Error).
		Bold(true).
		MarginTop(1)
}

// KeyHintStyle formats the submit/cancel key hints below the input.
func KeyHintStyle(th *theme.Theme) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(th.Muted).
		MarginTop(1)
}
