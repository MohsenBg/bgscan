// Package app implements the root BubbleTea application model and view.
package app

import (
	"bgscan/internal/core/config"
	"bgscan/internal/ui/main/body"
	"bgscan/internal/ui/main/footer"
	"bgscan/internal/ui/main/header"
	"bgscan/internal/ui/shared/dialog"
	"bgscan/internal/ui/shared/layout"
	"bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

// dialogPosition stores overlay placement metadata for a component.
type dialogPosition struct {
	XPos    dialog.DialogPosition
	YPos    dialog.DialogPosition
	XOffset int
	YOffset int
}

// model is the root BubbleTea model.
type model struct {
	state            *ui.AppState
	dialog           []ui.Component
	dialogPlacements map[ui.ComponentID]*dialogPosition
	header           ui.Component
	body             ui.Component
	footer           ui.Component
}

// New initializes the root application model.
func New(cfg *config.ScannerConfig, store *config.Store) tea.Model {
	l := layout.New()
	state := &ui.AppState{
		Layout: l,
		Config: cfg,
		Store:  store,
	}

	return &model{
		state:            state,
		dialog:           make([]ui.Component, 0, 5),
		dialogPlacements: make(map[ui.ComponentID]*dialogPosition),
		header:           header.New(l),
		body:             body.New(state),
		footer:           footer.New(l),
	}
}

// Init initializes the base components.
func (m *model) Init() tea.Cmd {
	return tea.Batch(
		m.header.Init(),
		m.body.Init(),
		m.footer.Init(),
	)
}

// getDialogPlacement returns placement for an overlay, creating a centered
// default if needed.
func (m *model) getDialogPlacement(id ui.ComponentID) *dialogPosition {
	if p, ok := m.dialogPlacements[id]; ok {
		return p
	}

	p := &dialogPosition{
		XPos:    dialog.Center,
		YPos:    dialog.Center,
		XOffset: 0,
		YOffset: 0,
	}

	m.dialogPlacements[id] = p
	return p
}
