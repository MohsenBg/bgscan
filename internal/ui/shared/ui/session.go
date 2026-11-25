package ui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/MohsenBg/bgscan/internal/core/config"
	"github.com/MohsenBg/bgscan/internal/logger"
	"github.com/MohsenBg/bgscan/internal/ui/shared/keys"
	"github.com/MohsenBg/bgscan/internal/ui/shared/layout"
	"github.com/MohsenBg/bgscan/internal/ui/theme"
)

// Program is the part of the running Bubble Tea program that the
// session needs. It lets code outside the update loop send a message
// into the program.
type Program interface {
	// Send delivers msg to the program's update loop.
	Send(msg tea.Msg)
}

// Session holds the state shared by the whole TUI run. Config and Store
// are set during startup, after NewSession returns, so components must
// not read them before startup fills them in.
type Session struct {
	env    *Env
	Config *config.ScannerConfig
	Store  *config.Store
	core   logger.Logger
}

// Env bundles what leaf components need to render and log.
type Env struct {
	Layout   layout.Layout
	Theme    theme.Theme
	Keys     *keys.Keys
	Log      logger.Logger
	Settings UISettings
	Program  Program
}

// NewSession builds a Session with an Env derived from settings. The ui
// logger goes into the Env for components, and the core logger is kept
// apart for scanner and config code. See CoreLog.
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

// Env returns the session environment. A nil session, or one built
// without an environment, returns a fresh default environment instead
// of nil, so callers can always read fields from the result.
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

// SetProgram records the running program in the Env so components can
// send messages to it. Startup calls it once the program exists.
func (s *Session) SetProgram(program Program) {
	s.Env().Program = program
}
