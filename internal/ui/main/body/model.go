package body

import (
	mainMenu "github.com/MohsenBg/bgscan/internal/ui/components/menus/entry"
	"github.com/MohsenBg/bgscan/internal/ui/shared/env"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

type Model struct {
	id         ui.ComponentID
	name       string
	state      *ui.AppState
	components []ui.Component
}

func New(state *ui.AppState) *Model {
	return &Model{
		id:         ui.NewComponentID(),
		name:       "body",
		state:      state,
		components: make([]ui.Component, 0, 4),
	}
}

func (m *Model) Init() tea.Cmd {
	m.components = append(m.components, mainMenu.New(m.state))
	return m.components[0].Init()
}

func (m *Model) ID() ui.ComponentID { return m.id }
func (m *Model) Mode() env.Mode     { return env.NormalMode }
func (m *Model) Name() string       { return m.name }
func (m *Model) OnClose() tea.Cmd   { return nil }
