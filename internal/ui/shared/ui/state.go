package ui

import (
	tea "charm.land/bubbletea/v2"

	"github.com/MohsenBg/bgscan/internal/core/config"
	"github.com/MohsenBg/bgscan/internal/logger"
	"github.com/MohsenBg/bgscan/internal/ui/shared/layout"
	"github.com/MohsenBg/bgscan/internal/ui/theme"
)

// Program is the subset of *tea.Program that AppState needs.
// Using an interface keeps AppState decoupled/testable.
type Program interface {
	Send(msg tea.Msg)
}

// AppState holds shared application state available to UI components.
type AppState struct {
	Layout   *layout.Layout
	Config   *config.ScannerConfig
	Store    *config.Store
	Program  Program
	Log      logger.Set
	Settings UISettings
}

// ThemeChangedMsg is broadcast when the user picks a new theme in UI
// settings. Components holding a cached *theme.Theme should handle it by
// calling their SetTheme.
type ThemeChangedMsg struct {
	Theme *theme.Theme
	Name  string
}

// Theme returns the active palette, resolving auto-detect when Settings.Theme
// is nil (e.g. tests constructing AppState by hand).
func (s *AppState) Theme() *theme.Theme {
	if s != nil && s.Settings.Theme != nil {
		return s.Settings.Theme
	}
	th, _ := theme.Get("")
	return th
}
