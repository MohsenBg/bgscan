package menu

import (
	"github.com/MohsenBg/bgscan/internal/ui/theme"

	"charm.land/lipgloss/v2"
)

func resolveTheme(th *theme.Theme) *theme.Theme {
	if th != nil {
		return th
	}
	t, _ := theme.Get("")
	return t
}

func itemTitleStyle(th *theme.Theme) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(resolveTheme(th).Text).
		Padding(0, 0, 0, 0)
}

func selectedItemTitleStyle(th *theme.Theme) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(resolveTheme(th).Primary).
		Padding(0, 1).
		Bold(true)
}

func shortcutStyle(th *theme.Theme) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(resolveTheme(th).Text)
}

func iconStyle(th *theme.Theme) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(resolveTheme(th).Text).
		Width(3).
		Bold(true)
}

func selectedIconStyle(th *theme.Theme) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(resolveTheme(th).Primary).
		Width(3).
		Bold(true)
}

func titleStyle(th *theme.Theme) lipgloss.Style {
	return lipgloss.NewStyle().
		Align(lipgloss.Center).
		Foreground(resolveTheme(th).Info).
		Bold(true)
}

// PaddingCell adds one line of top padding between menu rows.
func PaddingCell() lipgloss.Style {
	return lipgloss.NewStyle().
		Padding(1, 0, 0, 0)
}
