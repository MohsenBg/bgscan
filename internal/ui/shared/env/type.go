package env

// Mode represents the current UI state of the application.
type Mode int

const (
	NormalMode Mode = iota
	InputMode
	ManagedMode
)

// String returns the string representation of the Mode.
// useful for logging and debugging the TUI state.
func (m Mode) String() string {
	switch m {
	case NormalMode:
		return "NormalMode"
	case InputMode:
		return "InputMode"
	case ManagedMode:
		return "ManagedMode"
	default:
		return "UnknownMode"
	}
}
