package scanner

// tickMsg drives the scanner's periodic update loop.
// Triggered by a BubbleTea tick command.
type tickMsg struct{}

// immediateTickMsg forces an instant UI refresh without
// waiting for the next tick interval.
type immediateTickMsg struct{}

// TogglePauseMsg is emitted when the user requests to
// pause or resume the active scan.
type TogglePauseMsg struct{}

// scanClosedMsg is sent when Scanner.Close() returns,
// allowing the UI to navigate away without blocking.
type scanClosedMsg struct {
	err error
}

// scanErrorMsg is sent once when the scan goroutine
// encounters an error that hasn't been shown yet.
type scanErrorMsg struct {
	err error
}
