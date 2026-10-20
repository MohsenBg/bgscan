package app

import (
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
	m.session.Program = program
}

// Init starts the splash screen.
func (m *model) Init() tea.Cmd {
	return tea.Sequence(tea.ClearScreen, m.splash.Init())
}
