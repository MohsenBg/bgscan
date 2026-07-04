package logview

import (
	"bgscan/internal/ui/theme"

	"charm.land/lipgloss/v2"
)

//
// ────────────────────────────────────────────────────────────────────────────────
//  Styles for Log Viewer UI Components
// ────────────────────────────────────────────────────────────────────────────────
//

// TitleStyle returns the style for the log viewer's title bar.
//
// Behavior:
//   - Width is enforced to ensure alignment with the container.
//   - Title is centered horizontally and vertically.
//   - Bottom-only border for visual separation.
//   - Uses active border color from the theme.
//
// Visual characteristics:
//   - Bold title text
//   - Minimal border decoration for clean look
func TitleStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center, lipgloss.Center).
		Bold(true).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(theme.Current().BorderActive)
}

// ContainerStyle returns the base style for the log container.
// This defines the main bounding box for the log output area.
//
// Behavior:
//   - Width locked to the given value
//   - Horizontal/vertical centering alignment for inner content
func ContainerStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center, lipgloss.Center)
}

// BorderStyle returns the decorative border style around the log view.
// The border width is clamped to a maximum of 80 characters for consistent
// visual layout on small terminals.
//
// Behavior:
//   - Double-line border for strong contrast
//   - Active border color for emphasis
//   - Content centered within the bordered box
func BorderStyle(width int) lipgloss.Style {
	width = min(80, width)
	return lipgloss.NewStyle().Padding(0, 1).
		Width(width).
		Align(lipgloss.Left).
		Border(lipgloss.RoundedBorder(), true, true, true, true).
		BorderForeground(theme.Current().Secondary)
}

// helpStyle returns the style used for the footer help text,
// typically used for keyboard hints like "Press q to exit".
//
// Behavior:
//   - Slightly narrower than container (width - 5) to prevent overflow
//   - Muted text color for subtlety
//   - Centered alignment
//   - Padding for breathing room around the help text
func helpStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width-5).
		Foreground(theme.Current().Muted).
		Align(lipgloss.Center, lipgloss.Center).
		Padding(1)
}

// helpKeyStyle returns the style for highlighted key tokens inside help text,
// such as the "q" in "Press q to exit".
//
// Behavior:
//   - Bold and secondary-colored for visual contrast
//   - Designed to be embedded inside helpStyle()
func helpKeyStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Secondary).
		Bold(true)
}
