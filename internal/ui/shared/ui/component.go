package ui

import (
	"uuid"

	"github.com/MohsenBg/bgscan/internal/ui/shared/env"

	tea "charm.land/bubbletea/v2"
)

// ComponentID uniquely identifies a UI component instance.
type ComponentID uuid.UUID

// NewComponentID generates a new unique ComponentID.
func NewComponentID() ComponentID {
	return ComponentID(uuid.New())
}

// Component is a self-contained UI module managed by the application.
type Component interface {
	// ID returns the unique identifier of the component instance.
	ID() ComponentID

	// Name returns a human readable component name.
	Name() string

	// Init is called when the component is first mounted.
	Init() tea.Cmd

	// Update handles incoming BubbleTea messages and updates
	// the component state.
	Update(tea.Msg) (Component, tea.Cmd)

	// View renders the component UI.
	View() string

	// OnClose is executed when the component is removed
	// from the component stack.
	OnClose() tea.Cmd

	// Mode returns the input mode the component operates in.
	Mode() env.Mode
}

// CloseComponentMsg signals that a component should be closed.
type CloseComponentMsg struct {
	ID ComponentID
}

// OpenComponentMsg requests opening a new component.
type OpenComponentMsg struct {
	Component Component
}

// ResetComponentStacksMsg clears all component stacks.
type ResetComponentStacksMsg struct{}

// OpenComponentCmd returns a command that emits an OpenComponentMsg to mount
// the given component.
func OpenComponentCmd(component Component) tea.Cmd {
	return func() tea.Msg {
		return OpenComponentMsg{
			Component: component,
		}
	}
}
