// Package header renders the top banner of the application.
package header

import (
	"bgscan/internal/ui/shared/env"
	"bgscan/internal/ui/shared/layout"
	"bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

// Model is the header component, currently displaying the app banner.
type Model struct {
	id     ui.ComponentID
	name   string
	layout *layout.Layout
}

// New creates the header model.
func New(l *layout.Layout) *Model {
	return &Model{
		layout: l,
		name:   "Header",
		id:     ui.NewComponentID(),
	}
}

// ID returns the component identifier.
func (m *Model) ID() ui.ComponentID {
	return m.id
}

// Name returns the component name.
func (m *Model) Name() string {
	return m.name
}

// Init is a no-op for the header.
func (m *Model) Init() tea.Cmd {
	return nil
}

// OnClose is a no-op for the header.
func (m *Model) OnClose() tea.Cmd {
	return nil
}

// Mode returns the header interaction mode.
func (m *Model) Mode() env.Mode {
	return env.NormalMode
}
