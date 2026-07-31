// Package ui defines shared interfaces and state used across
// application UI components.
package ui

import (
	"bgscan/internal/core/config"
	"bgscan/internal/ui/shared/layout"
)

// AppState holds globally shared application state available to UI components.
type AppState struct {
	Layout *layout.Layout
	Config *config.ScannerConfig
	Store  *config.Store
}
