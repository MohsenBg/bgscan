package footer

import (
	"bgscan/internal/ui/theme"

	"github.com/charmbracelet/lipgloss"
)

var (
	// ----------------------------------------
	// Container
	// ----------------------------------------

	containerStyle = func(width, height int) lipgloss.Style {
		t := theme.Current()

		return lipgloss.NewStyle().
			Width(width).
			Height(height).
			Foreground(t.Text)
	}

	// ----------------------------------------
	// Separator
	// ----------------------------------------

	separatorStyle = func(width int) lipgloss.Style {
		t := theme.Current()

		return lipgloss.NewStyle().
			Width(width).
			Foreground(t.Border)
	}

	// ----------------------------------------
	// Sections
	// ----------------------------------------

	leftSectionStyle = func(width int) lipgloss.Style {
		return lipgloss.NewStyle().
			Width(width).
			Padding(0, 1).
			Align(lipgloss.Left)
	}

	centerSectionStyle = func(width int) lipgloss.Style {
		return lipgloss.NewStyle().
			Width(width).
			Padding(0, 1).
			Align(lipgloss.Center)
	}

	rightSectionStyle = func(width int) lipgloss.Style {
		return lipgloss.NewStyle().
			Width(width).
			Padding(0, 1).
			Align(lipgloss.Right)
	}

	// ----------------------------------------
	// Text styles
	// ----------------------------------------

	appNameStyle = func() lipgloss.Style {
		t := theme.Current()

		return lipgloss.NewStyle().
			Foreground(t.Yellow).
			Bold(true)
	}

	versionStyle = func() lipgloss.Style {
		t := theme.Current()

		return lipgloss.NewStyle().
			Foreground(t.Success).
			Faint(true)
	}

	statusTextStyle = func() lipgloss.Style {
		t := theme.Current()

		return lipgloss.NewStyle().
			Foreground(t.Primary).
			Bold(true)
	}

	iconStyle = func() lipgloss.Style {
		t := theme.Current()

		return lipgloss.NewStyle().
			Foreground(t.Orange)
	}
)
