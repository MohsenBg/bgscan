package body

import (
	mainMenu "github.com/MohsenBg/bgscan/internal/ui/components/menus/entry"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

type Model struct {
	id      ui.ComponentID
	name    string
	session *ui.Session
	screens []screen

	// swap is the running screen transition; nil when the stack is settled.
	swap *swap
}

func New(session *ui.Session) *Model {
	return &Model{
		id:      ui.NewComponentID(),
		name:    "body",
		session: session,
		screens: make([]screen, 0, 4),
	}
}

func (m *Model) Init() tea.Cmd {
	m.screens = append(m.screens, screen{component: mainMenu.New(m.session)})
	return m.screens[0].component.Init()
}

func (m *Model) ID() ui.ComponentID { return m.id }
func (m *Model) Name() string       { return m.name }

func (m *Model) Mode() ui.KeyMode {
	for i := len(m.screens) - 1; i >= 0; i-- {
		if s := m.screens[i]; !s.closing() {
			return s.component.Mode()
		}
	}

	return ui.KeyModeNormal
}

func (m *Model) OnClose() tea.Cmd { return nil }
