package confirm

import (
	"github.com/MohsenBg/bgscan/internal/ui/shared/layout"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

// Model represents a confirmation dialog component.
type Model struct {
	id   ui.ComponentID
	name string

	layout *layout.Layout

	message string
	confirm bool

	confirmFunc tea.Cmd
}

// New creates a new confirmation dialog.
func New(
	layout *layout.Layout,
	message string,
	onConfirm func() tea.Msg,
	defaultYes bool,
) *Model {
	return &Model{
		id:          ui.NewComponentID(),
		name:        "confirm",
		layout:      layout,
		message:     message,
		confirm:     defaultYes,
		confirmFunc: onConfirm,
	}
}

// ID returns the component identifier.
func (m *Model) ID() ui.ComponentID {
	return m.id
}

// Init initializes the component.
func (m *Model) Init() tea.Cmd {
	return nil
}

// Name returns the component name.
func (m *Model) Name() string {
	return m.name
}

// OnClose is called when the component is removed.
func (m *Model) OnClose() tea.Cmd {
	return nil
}
