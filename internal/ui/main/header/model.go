package header

import (
	"bgscan/internal/ui/shared/env"
	"bgscan/internal/ui/shared/layout"
	"bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

type Model struct {
	id     ui.ComponentID
	name   string
	layout *layout.Layout
}

func New(l *layout.Layout) *Model {
	return &Model{layout: l, name: "Header", id: ui.NewComponentID()}
}

func (m *Model) ID() ui.ComponentID { return m.id }
func (m *Model) Mode() env.Mode     { return env.NormalMode }
func (m *Model) Name() string       { return m.name }
func (m *Model) Init() tea.Cmd      { return nil }
func (m *Model) OnClose() tea.Cmd   { return nil }
