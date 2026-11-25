package confirm

import (
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

// Model is a confirmation dialog component with a message and a No
// and Yes choice.
type Model struct {
	id   ui.ComponentID
	name string

	env *ui.Env

	message string
	confirm bool

	confirmFunc tea.Cmd
}

// New creates a confirmation dialog that shows message and runs
// onConfirm when the user picks Yes. defaultYes selects Yes at first
// instead of No.
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

// ID returns the unique identifier of the component.
func (m *Model) ID() ui.ComponentID {
	return m.id
}

// Init returns no command, as the dialog has nothing to load.
func (m *Model) Init() tea.Cmd {
	return nil
}

// Name returns the component name.
func (m *Model) Name() string {
	return m.name
}

// OnClose returns no command, as the dialog holds nothing to release.
func (m *Model) OnClose() tea.Cmd {
	return nil
}
