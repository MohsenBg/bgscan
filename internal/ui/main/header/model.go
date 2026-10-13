package header

import (
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

type Model struct {
	id   ui.ComponentID
	name string
	env  *ui.Env
}

func New(env *ui.Env) *Model {
	return &Model{env: env, name: "Header", id: ui.NewComponentID()}
}

func (m *Model) ID() ui.ComponentID { return m.id }
func (m *Model) Mode() ui.KeyMode   { return ui.KeyModeNormal }
func (m *Model) Name() string       { return m.name }
func (m *Model) Init() tea.Cmd      { return nil }
func (m *Model) OnClose() tea.Cmd   { return nil }
