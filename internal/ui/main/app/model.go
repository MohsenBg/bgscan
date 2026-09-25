package app

import (
	"github.com/MohsenBg/bgscan/internal/logger"
	"github.com/MohsenBg/bgscan/internal/ui/main/splash"
	"github.com/MohsenBg/bgscan/internal/ui/main/startup"
	"github.com/MohsenBg/bgscan/internal/ui/main/workspace"
	"github.com/MohsenBg/bgscan/internal/ui/shared/layout"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
	"github.com/MohsenBg/bgscan/internal/ui/theme"

	tea "charm.land/bubbletea/v2"
)

type Application interface {
	tea.Model
	SetProgram(program ui.Program)
}

type AppStage uint8

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
}

func New(logs logger.Set) Application {
	l := layout.New()
	theme.Init()

	settings := ui.LoadSettings(logs.UI)
	if lvl, err := logger.ParseLevel(settings.LogLevel); err == nil {
		logs.SetLevel(lvl)
	}

	state := &ui.AppState{Layout: l, Log: logs, Settings: settings}

	return &model{
		state:     state,
		splash:    splash.New(state),
		startup:   startup.New(state),
		workspace: workspace.New(state),
		stage:     StageSplash,
	}
}

func (m *model) SetProgram(program ui.Program) {
	m.state.Program = program
}

func (m *model) Init() tea.Cmd {
	return tea.Sequence(tea.ClearScreen, m.splash.Init())
}
