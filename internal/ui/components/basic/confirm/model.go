package confirm

import (
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

// Model represents a confirmation dialog component.
type Model struct {
	id   ui.ComponentID
	name string

	env *ui.Env

	message string
	confirm bool

	confirmFunc tea.Cmd
}

func New(
	env *ui.Env,
	message string,
	onConfirm func() tea.Msg,
	defaultYes bool,
) *Model {
	return &Model{
		id:          ui.NewComponentID(),
		name:        "confirm",
		env:         env,
		message:     message,
		confirm:     defaultYes,
		confirmFunc: onConfirm,
	}
}

func (m *Model) ID() ui.ComponentID {
	return m.id
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Name() string {
	return m.name
}

func (m *Model) OnClose() tea.Cmd {
	return nil
}
