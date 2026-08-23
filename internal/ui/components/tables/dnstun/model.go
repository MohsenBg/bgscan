package dnstun

import (
	"bgscan/internal/core/dns"
	"bgscan/internal/ui/components/basic/crud"
	"bgscan/internal/ui/shared/env"
	"bgscan/internal/ui/shared/layout"
	"bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

// Model coordinates outbound configuration additions, list table management,
// and multi-step dialog sequencing paths within the UI stack.
type Model struct {
	id        ui.ComponentID
	name      string
	layout    *layout.Layout
	crudTable *crud.Model[dns.DNSTunConfigFile]
}

// New creates a new outbound template list component view layer.
func New(l *layout.Layout, title string, onSelect func(*dns.DNSTunConfigFile) tea.Cmd) *Model {
	m := &Model{
		id:     ui.NewComponentID(),
		name:   "outbounds",
		layout: l,
	}

	canAdd := true
	m.crudTable = crud.New("dns tunling", l, newProvider(l, onSelect), 100, canAdd)

	return m
}

func (m *Model) Init() tea.Cmd      { return m.crudTable.Init() }
func (m *Model) ID() ui.ComponentID { return m.id }
func (m *Model) Name() string       { return m.name }
func (m *Model) OnClose() tea.Cmd   { return m.crudTable.OnClose() }
func (m *Model) Mode() env.Mode     { return m.crudTable.Mode() }
