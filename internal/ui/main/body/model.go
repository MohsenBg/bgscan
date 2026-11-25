// Package body owns the stack of full screens shown in the workspace.
// The main menu is the root screen and cannot be closed. Opening a screen
// pushes it on the stack, the back key pops it, and push and pop play a
// slide transition when animations are enabled.
package body

import (
	mainMenu "github.com/MohsenBg/bgscan/internal/ui/components/menus/entry"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

// Model is the body component. It holds the screen stack and forwards
// input to the top screen.
type Model struct {
	id      ui.ComponentID
	name    string
	session *ui.Session
	screens []screen

	// swap is the running screen transition; nil when the stack is settled.
	swap *swap
}

// New creates an empty body. Init pushes the main menu as the root screen.
func New(session *ui.Session) *Model {
	return &Model{
		id:      ui.NewComponentID(),
		name:    "body",
		session: session,
		screens: make([]screen, 0, 4),
	}
}

// Init pushes the main menu as the root screen and starts it.
func (m *Model) Init() tea.Cmd {
	m.screens = append(m.screens, screen{component: mainMenu.New(m.session)})
	return m.screens[0].component.Init()
}

// ID returns the component ID of the model. Name returns its display
// name.
func (m *Model) ID() ui.ComponentID { return m.id }
func (m *Model) Name() string       { return m.name }

// Mode returns the key mode of the topmost screen that is not closing, or
// KeyModeNormal when every screen is closing.
func (m *Model) Mode() ui.KeyMode {
	for i := len(m.screens) - 1; i >= 0; i-- {
		if s := m.screens[i]; !s.closing() {
			return s.component.Mode()
		}
	}

	return ui.KeyModeNormal
}

// OnClose returns nil. The body has no close work of its own.
func (m *Model) OnClose() tea.Cmd { return nil }
