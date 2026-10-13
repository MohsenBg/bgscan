package ui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/MohsenBg/bgscan/internal/core/config"
	"github.com/MohsenBg/bgscan/internal/logger"
	"github.com/MohsenBg/bgscan/internal/ui/shared/keys"
	"github.com/MohsenBg/bgscan/internal/ui/shared/layout"
	"github.com/MohsenBg/bgscan/internal/ui/theme"
)

type Program interface {
	Send(msg tea.Msg)
}

type Session struct {
	env      *Env
	Settings UISettings
	Config   *config.ScannerConfig
	Store    *config.Store
	Program  Program
	core     logger.Logger
}

// Env bundles what leaf components need to render and log.
type Env struct {
	Layout layout.Layout
	Theme  theme.Theme
	Keys   *keys.Keys
	Log    logger.Logger
}

func NewSession(settings UISettings, ui logger.Logger, core logger.Logger) *Session {
	env := newEnv(settings, ui)
	return &Session{env: env, Settings: settings, core: core}
}

func newEnv(setting UISettings, log logger.Logger) *Env {
	k := keys.Default()
	l := layout.New()
	th, _ := theme.Get(setting.ThemeName)

	e := &Env{Keys: &k, Log: log, Theme: th}
	if l != nil {
		e.Layout = *l
	}
	return e
}

func (s *Session) Env() *Env {
	if s == nil {
		return newEnv(DefaultUISettings(), *logger.Discard())
	}
	if s.env == nil {
		s.env = newEnv(s.Settings, *logger.Discard())
	}
	return s.env
}

// CoreLog returns the core.log file logger for scanner, config,
// result and xray call sites. Never hand this to leaf components.
func (s *Session) CoreLog() logger.Logger {
	if s == nil {
		return *logger.Discard()
	}
	return s.core
}

// SetTheme replaces the active palette, persists the name and returns
// a command broadcasting ThemeChanged to huh/bubbles-caching leaves.
func (s *Session) SetTheme(name string) (tea.Cmd, error) {
	th, err := theme.Get(name)
	if err != nil {
		return nil, err
	}
	s.Env().Theme = th
	s.Settings.ThemeName = name
	SaveSettings(s.Settings)
	return theme.ThemeChangedCmd(), nil
}
