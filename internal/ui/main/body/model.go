// Package body hosts the active main content area, including the entry menu
// stack and pushed subcomponents.
package body

import (
	mainMenu "bgscan/internal/ui/components/menus/entry"
	"bgscan/internal/ui/shared/env"
	"bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

// Model is the body region of the application. It maintains a stack of
// active components so users can navigate deeper and pop back.
type Model struct {
	id         ui.ComponentID
	name       string
	state      *ui.AppState
	components []ui.Component
}

// New creates the body model.
func New(state *ui.AppState) *Model {
	return &Model{
		id:         ui.NewComponentID(),
		name:       "body",
		state:      state,
		components: make([]ui.Component, 0, 4),
	}
}

// Init pushes the entry menu onto the component stack.
func (m *Model) Init() tea.Cmd {
	m.components = append(m.components, mainMenu.New(m.state))
	return m.components[0].Init()
}

// ID returns the component identifier.
func (m *Model) ID() ui.ComponentID {
	return m.id
}

// Name returns the component name.
func (m *Model) Name() string {
	return m.name
}

// OnClose is a no-op for the body stack.
func (m *Model) OnClose() tea.Cmd {
	return nil
}

// Mode returns the body interaction mode.
func (m *Model) Mode() env.Mode {
	return env.NormalMode
}
