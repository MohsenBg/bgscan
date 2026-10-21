package ui

// KeyMode describes how the TUI currently interprets key presses.
type KeyMode int

const (
	KeyModeNormal KeyMode = iota // navigation and shortcuts
	KeyModeInput                 // typing into a text field
	KeyModeManage                // managing items (rename, delete, reorder, ...)
)

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
