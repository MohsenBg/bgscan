package app

import (
	tea "charm.land/bubbletea/v2"

	"github.com/MohsenBg/bgscan/internal/logger"
	"github.com/MohsenBg/bgscan/internal/ui/main/splash"
	"github.com/MohsenBg/bgscan/internal/ui/main/startup"
	"github.com/MohsenBg/bgscan/internal/ui/main/workspace"
	"github.com/MohsenBg/bgscan/internal/ui/shared/layout"
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
	state     *ui.AppState
	splash    ui.Component
	startup   ui.Component
	workspace ui.Component
	stage     AppStage
	dialogStack
}

// New creates the root application model.
func New(logs logger.Set) Application {
	l := layout.New()
	theme.Init()

	settings := ui.LoadSettings(logs.UI)
	if lvl, err := logger.ParseLevel(settings.LogLevel); err == nil {
		logs.SetLevel(lvl)
	}

	state := &ui.AppState{
		Layout:   l,
		Log:      logs,
		Settings: settings,
	}

	return &model{
		state:       state,
		splash:      splash.New(state),
		startup:     startup.New(state),
		workspace:   workspace.New(state),
		stage:       StageSplash,
		dialogStack: newDialogStack(),
	}
}

// SetProgram stores the program for commands that need it.
func (m *model) SetProgram(program ui.Program) {
	m.state.Program = program
}

// Init starts the splash screen.
func (m *model) Init() tea.Cmd {
	return tea.Sequence(tea.ClearScreen, m.splash.Init())
}
