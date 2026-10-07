package menu

import (
	"image/color"

	"charm.land/lipgloss/v2"

	"github.com/MohsenBg/bgscan/internal/ui/theme"
)

// resolveTheme guards against a nil theme.
func resolveTheme(th *theme.Theme) *theme.Theme {
	if th == nil {
		fallback, err := theme.Get(theme.List()[0])
		if err == nil {
			return fallback
		}

		return &theme.Theme{}
	}

	return th
}

// titleStyle is the centered menu heading.
func titleStyle(th *theme.Theme) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(resolveTheme(th).Info).
		Bold(true)
}

// helpStyle is the muted key hint under the rows.
func helpStyle(th *theme.Theme) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(resolveTheme(th).Muted)
}

// rowIconStyle renders the icon column padded to a fixed width.
func rowIconStyle(fg color.Color, bold bool) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(fg).
		Bold(bold).
		Width(iconColumnWidth)
}

// rowTitleStyle renders a row title.
func rowTitleStyle(fg color.Color, bold bool) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(fg).
		Bold(bold)
}

// keycapStyle renders a shortcut chip.
func keycapStyle(fg, bg color.Color) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(fg).
		Background(bg).
		Padding(0, 1)
}
