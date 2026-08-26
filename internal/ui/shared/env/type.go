package env

// Mode represents the current UI state of the application.
//
// The mode determines how keyboard input is interpreted and
// what actions are available to the user.
type Mode int

const (
	// NormalMode is the default navigation mode.
	// Users can move through menus and trigger actions.
	NormalMode Mode = iota

	// InputMode is active when the user is typing input,
	// such as entering a target IP range or configuration.
	InputMode

	// ManagedMode is active while a network scan is running.
	// During this mode most UI actions are disabled and
	// control is handled by scanning goroutines.
	ManagedMode
)
