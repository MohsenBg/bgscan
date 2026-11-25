// Package app holds the root Bubble Tea model of the TUI.
// It moves the program through the splash, startup, and workspace stages,
// routes messages to the active stage, and draws open dialogs as animated
// layers above the stage. Key messages go to the top dialog only, while
// other messages go to every dialog layer.
package app

import (
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/MohsenBg/bgscan/internal/logger"
	"github.com/MohsenBg/bgscan/internal/ui/main/splash"
	"github.com/MohsenBg/bgscan/internal/ui/main/startup"
	"github.com/MohsenBg/bgscan/internal/ui/main/workspace"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
	"github.com/MohsenBg/bgscan/internal/ui/theme"
)

// Application is the root Bubble Tea model.
type Application interface {
	tea.Model

	// SetProgram stores the running program so commands can send messages
	// back into the update loop.
	SetProgram(program ui.Program)
}

// AppStage identifies the current application screen.
type AppStage uint8

// Application stages.
const (
	StageSplash AppStage = iota
	StageStartUP
	StageWorkspace
)

type model struct {
	session              *ui.Session
	splash               ui.Component
	startup              ui.Component
	workspace            ui.Component
	stage                AppStage
	isNonLatinDialogOpen bool
	dialogStack

	// base caches the last workspace render, with the time it was made, so
	// View can reuse it while a dialog is open.
	base   string
	baseAt time.Time
	baseOK bool
}

// New creates the root application model.
func New(uiLogger, coreLogger logger.Logger) Application {
	theme.Init()

	settings := ui.LoadSettings(uiLogger)
	if lvl, err := logger.ParseLevel(settings.LogLevel); err == nil {
		uiLogger.SetLevel(lvl)
		coreLogger.SetLevel(lvl)
	}

	session := ui.NewSession(settings, uiLogger, coreLogger)
	return &model{
		session:     session,
		splash:      splash.New(session),
		startup:     startup.New(session),
		workspace:   workspace.New(session),
		stage:       StageSplash,
		dialogStack: newDialogStack(),
	}
}

// SetProgram stores the program for commands that need it.
func (m *model) SetProgram(program ui.Program) {
	m.session.SetProgram(program)
}

// Init starts the splash screen.
func (m *model) Init() tea.Cmd {
	return tea.Sequence(tea.ClearScreen, m.splash.Init())
}
