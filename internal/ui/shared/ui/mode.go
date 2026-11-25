package ui

// KeyMode describes how the TUI currently interprets key presses.
type KeyMode int

const (
	// KeyModeNormal is the mode for navigation and shortcuts.
	KeyModeNormal KeyMode = iota
	// KeyModeInput is the mode for typing into a text field.
	KeyModeInput
	// KeyModeManage is the mode for managing items, such as rename,
	// delete and reorder.
	KeyModeManage
)

// String returns the display name of the mode, or "Unknown" for a value
// outside the defined modes.
func (m KeyMode) String() string {
	switch m {
	case KeyModeNormal:
		return "Normal"
	case KeyModeInput:
		return "Input"
	case KeyModeManage:
		return "Manage"
	default:
		return "Unknown"
	}
}
