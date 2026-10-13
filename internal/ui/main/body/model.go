package body

import (
	mainMenu "github.com/MohsenBg/bgscan/internal/ui/components/menus/entry"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

type Model struct {
	id         ui.ComponentID
	name       string
	session    *ui.Session
	components []ui.Component
}

func New(session *ui.Session) *Model {
	return &Model{
		id:         ui.NewComponentID(),
		name:       "body",
		session:    session,
		components: make([]ui.Component, 0, 4),
	}
}

func (m *Model) Init() tea.Cmd {
	m.components = append(m.components, mainMenu.New(m.session))
	return m.components[0].Init()
}

func (m *Model) ID() ui.ComponentID { return m.id }
func (m *Model) Mode() ui.KeyMode   { return ui.KeyModeNormal }
func (m *Model) Name() string       { return m.name }
func (m *Model) OnClose() tea.Cmd   { return nil }
