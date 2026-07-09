package inspector

import (
	"bgscan/internal/ui/theme"

	"charm.land/lipgloss/v2"
)

//
// ────────────────────────────────────────────────────────────────────────────────
//  Menu Styles
// ────────────────────────────────────────────────────────────────────────────────
//  This file defines the visual styles used by the menu component.
//  All styles rely on the active UI theme for consistent coloring.
//

// filedNameStyle returns the style used for normal (non‑selected) menu item titles.
//
// Visual characteristics:
//   - Standard foreground color from the theme
//   - No padding to keep compact alignment with icons and shortcuts
func filedNameStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Text).
		Padding(0, 1)
}

// selectedFiledNameStyle returns the style for the currently focused
// or selected menu item title.
//
// Visual characteristics:
//   - Primary theme color for strong emphasis
//   - Bold text
//   - Slight horizontal padding to visually separate the highlight
func selectedFiledNameStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Primary).
		Padding(0, 0).
		Bold(true)
}

// vlaueStyle returns the style used for displaying keyboard shortcuts
// associated with menu actions (e.g. "Ctrl+C", "Enter").
//
// The shortcut text uses the default text color to maintain readability
// without competing visually with the selected item highlight.
func vlaueStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Text)
}

// titleStyle returns the style used for the menu title or header.
//
// Visual characteristics:
//   - Center-aligned text
//   - Informational theme color
//   - Bold font weight for prominence
func titleStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Align(lipgloss.Center).
		Foreground(theme.Current().Info).
		Bold(true)
}

// PaddingCell returns a helper style used to create vertical spacing
// between menu rows or sections.
//
// Behavior:
//   - Adds one line of top padding
//   - Does not affect horizontal spacing
func PaddingCell() lipgloss.Style {
	return lipgloss.NewStyle().
		Padding(1, 0, 0, 0)
}
