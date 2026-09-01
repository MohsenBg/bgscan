package ui

import (
	tea "charm.land/bubbletea/v2"

	"github.com/MohsenBg/bgscan/internal/core/config"
	"github.com/MohsenBg/bgscan/internal/ui/shared/layout"
)

// Program is the subset of *tea.Program that AppState needs.
// Using an interface keeps AppState decoupled/testable.
type Program interface {
	Send(msg tea.Msg)
}

// AppState holds globally shared application state available to UI components.
type AppState struct {
	Layout  *layout.Layout
	Config  *config.ScannerConfig
	Store   *config.Store
	Program Program
}

func NewAppState(l *layout.Layout, cfg *config.ScannerConfig, store *config.Store) *AppState {
	return &AppState{
		Layout: l,
		Config: cfg,
		Store:  store,
	}
}
