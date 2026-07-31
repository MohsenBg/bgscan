// Package theme provides a centralized color palette and theme management
// system for the terminal UI.
//
// It supports three modes:
//
//   - ModeDark  – forces the dark color palette
//   - ModeLight – forces the light color palette
//   - ModeAuto  – automatically selects a palette based on terminal background
//
// The package exposes a minimal API used by UI components to retrieve
// the active theme and react to theme changes.
//
// Example:
//
//	th := theme.Current()
//
//	title := lipgloss.NewStyle().
//		Foreground(th.Primary).
//		Bold(true)
//
//	fmt.Println(title.Render("BGScan"))
package theme

import (
	"image/color"
	"os"

	"charm.land/lipgloss/v2"
)

type ThemeMode int

const (
	// ModeAuto selects the theme automatically based on terminal detection.
	ModeAuto ThemeMode = iota

	// ModeDark forces the dark color palette.
	ModeDark

	// ModeLight forces the light color palette.
	ModeLight
)

// Theme is the app's active color palette. Components should use it rather
// than hardcoding colors so the UI stays consistent across theme modes.
type Theme struct {
	Primary   color.Color
	Secondary color.Color

	Border       color.Color
	BorderActive color.Color

	Text  color.Color
	Muted color.Color

	Info    color.Color
	Error   color.Color
	Success color.Color

	Orange color.Color
	Yellow color.Color
	Purple color.Color

	ProgressStart color.Color
	ProgressEnd   color.Color
}

// Dark defines the dark terminal color palette.
var Dark = Theme{
	Primary:       lipgloss.Color("#D75FD7"),
	Secondary:     lipgloss.Color("#8A8A8A"),
	Border:        lipgloss.Color("#585858"),
	BorderActive:  lipgloss.Color("#5F5FD7"),
	Text:          lipgloss.Color("#D0D0D0"),
	Muted:         lipgloss.Color("#626262"),
	Error:         lipgloss.Color("#FF0000"),
	Success:       lipgloss.Color("#00D787"),
	Info:          lipgloss.Color("#00AFFF"),
	Orange:        lipgloss.Color("#FF8700"),
	Yellow:        lipgloss.Color("#FFD700"),
	Purple:        lipgloss.Color("#5F5FD7"),
	ProgressStart: lipgloss.Color("#A78BFA"),
	ProgressEnd:   lipgloss.Color("#7DD3FC"),
}

// Light defines the light terminal color palette.
var Light = Theme{
	Primary:       lipgloss.Color("#D75FD7"),
	Secondary:     lipgloss.Color("#A8A8A8"),
	Border:        lipgloss.Color("#808080"),
	BorderActive:  lipgloss.Color("#5F5FD7"),
	Text:          lipgloss.Color("#1C1C1C"),
	Muted:         lipgloss.Color("#949494"),
	Info:          lipgloss.Color("#005FFF"),
	Error:         lipgloss.Color("#D70000"),
	Success:       lipgloss.Color("#008700"),
	Orange:        lipgloss.Color("#FF8700"),
	ProgressStart: lipgloss.Color("#6D28D9"),
	ProgressEnd:   lipgloss.Color("#0369A1"),
}

var (
	current Theme
	mode    = ModeAuto
)

// Current returns the active theme palette.
//
// The returned Theme is a copy of internal state and is safe for concurrent reads.
func Current() Theme {
	return current
}

// Mode returns the currently configured ThemeMode.
func Mode() ThemeMode {
	return mode
}

// SetMode updates the active theme mode and resolves the matching palette.
func SetMode(m ThemeMode) {
	mode = m
	resolve()
}

func resolve() {
	switch mode {

	case ModeDark:
		current = Dark

	case ModeLight:
		current = Light

	case ModeAuto:
		if terminalLooksDark() {
			current = Dark
		} else {
			current = Light
		}

	}
}

// terminalLooksDark detects terminal background darkness from COLORFGBG.
func terminalLooksDark() bool {
	bg := os.Getenv("COLORFGBG")

	if bg == "" {
		return true
	}

	for i := len(bg) - 1; i >= 0; i-- {
		if bg[i] == ';' {
			bg = bg[i+1:]
			break
		}
	}

	switch bg {
	case "0", "1", "2", "3", "4", "5", "6", "7":
		return true
	default:
		return false
	}
}

// Init initializes the active theme from the configured mode.
func Init() {
	resolve()
}
