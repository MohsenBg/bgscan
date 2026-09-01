package startup

import "github.com/MohsenBg/bgscan/internal/core/config"

// categoryStartMsg marks a category as having begun running.
type categoryStartMsg struct {
	categoryID string
}

// categoryEndMsg marks a category as finished, carrying its final status.
type categoryEndMsg struct {
	categoryID string
	status     categoryStatus
}

// configLoadedMsg delivers the loaded config/store once, sent directly via
// AppState.Send from the checklist goroutine rather than through a mutex.
type configLoadedMsg struct {
	cfg   *config.ScannerConfig
	store *config.Store
}

// StartupChecksDoneMsg signals that the user has confirmed and startup is complete.
type StartupChecksDoneMsg struct{}
