package env

import (
	"slices"

	tea "charm.land/bubbletea/v2"
)

// Key string constants used when matching Bubble Tea key press messages.
const (
	KeyEnter     = "enter"
	KeyEsc       = "esc"
	KeyBackspace = "backspace"
	KeyCtrlC     = "ctrl+c"
	KeyTab       = "tab"
	KeyShiftTab  = "shift+tab"
	KeyCtrlT     = "ctrl+t"
)

var backKeys = map[Mode][]string{
	NormalMode: {"b", KeyBackspace, KeyEsc},
	InputMode:  {KeyEsc},
	ScanMode:   {},
}

var quitKeys = map[Mode][]string{
	NormalMode: {"q", KeyCtrlC},
	InputMode:  {KeyCtrlC},
	ScanMode:   {KeyCtrlC},
}

// IsBackKey reports whether the key press represents a back action in the given mode.
func IsBackKey(msg tea.KeyPressMsg, mode Mode) bool {
	if keys, ok := backKeys[mode]; ok {
		return slices.Contains(keys, msg.String())
	}
	return false
}

// IsQuitKey reports whether the key press represents a quit action in the given mode.
func IsQuitKey(msg tea.KeyPressMsg, mode Mode) bool {
	if keys, ok := quitKeys[mode]; ok {
		return slices.Contains(keys, msg.String())
	}
	return false
}
