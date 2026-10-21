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
	env     *Env
	Config  *config.ScannerConfig
	Store   *config.Store
	Program Program
	core    logger.Logger
}

// Env bundles what leaf components need to render and log.
type Env struct {
	Layout   layout.Layout
	Theme    theme.Theme
	Keys     *keys.Keys
	Log      logger.Logger
	Settings UISettings
}

func NewSession(settings UISettings, ui logger.Logger, core logger.Logger) *Session {
	env := newEnv(settings, ui)
	return &Session{env: env, core: core}
}

func newEnv(setting UISettings, log logger.Logger) *Env {
	k := keys.Default()
	l := layout.New()
	th, _ := theme.Get(setting.ThemeName)

	e := &Env{Keys: &k, Settings: setting, Log: log, Theme: th}
	if l != nil {
		e.Layout = *l
	}
	return e
}

func (s *Session) Env() *Env {
	if s == nil || s.env == nil {
		return newEnv(DefaultUISettings(), *logger.Discard())
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
	s.Env().Settings.ThemeName = name
	err = SaveSettings(s.Env().Settings)
	if err != nil {
		return nil, err
	}
	return theme.ThemeChangedCmd(), nil
}
