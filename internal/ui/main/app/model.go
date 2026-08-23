package app

import (
	"bgscan/internal/ui/main/splash"
	"bgscan/internal/ui/main/startup"
	"bgscan/internal/ui/main/workspace"
	"bgscan/internal/ui/shared/layout"
	"bgscan/internal/ui/shared/ui"

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

func New() Application {
	l := layout.New()
	state := &ui.AppState{Layout: l}
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
