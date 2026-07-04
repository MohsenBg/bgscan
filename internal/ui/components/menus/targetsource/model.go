package targetsource

import (
	"bgscan/internal/core/iplist"
	"bgscan/internal/core/result"
	"bgscan/internal/ui/components/basic/menu"
	"bgscan/internal/ui/components/menus/scantype"
	iplistTable "bgscan/internal/ui/components/tables/iplist"
	resultlistTable "bgscan/internal/ui/components/tables/resultlist"
	"bgscan/internal/ui/shared/env"
	"bgscan/internal/ui/shared/layout"
	"bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

// ═══ Model ═══
type Model struct {
	id     ui.ComponentID
	name   string
	layout *layout.Layout
	menu   ui.Component
}

// ═══ Constructor ═══
func New(layout *layout.Layout) *Model {
	m := &Model{
		layout: layout,
		id:     ui.NewComponentID(),
		name:   "Target Source",
	}
	items := []menu.MenuItem{
		menu.NewMenuItem("▤", "IP List", "i", func() tea.Msg {
			return m.OpenIPList(func(i *iplist.IPFileInfo) tea.Cmd {
				return ui.OpenComponentCmd(scantype.New(layout, i.Path))
			})
		}),
		menu.NewMenuItem("▤", "Result List", "r", func() tea.Msg {
			return m.OpenResultIPList(func(r *result.ResultFile) tea.Cmd {
				return ui.OpenComponentCmd(scantype.New(layout, r.Path))
			})
		}),
	}

	m.menu = menu.New(items, "Select Target Source", layout)
	return m
}

// OpenIPList opens the IP file picker overlay.
// onSelect is called by the iplist component once the user picks a file.
func (m *Model) OpenIPList(onSelect func(*iplist.IPFileInfo) tea.Cmd) tea.Msg {
	return ui.OpenComponentMsg{Component: iplistTable.New(m.layout, "Select IP File", onSelect)}
}

// OpenResultIPList opens the ResultIP file picker overlay.
// onSelect is called by the resultlist component once the user picks a file.
func (m *Model) OpenResultIPList(onSelect func(*result.ResultFile) tea.Cmd) tea.Msg {
	var maxRenderIP uint32 = 10_000
	return ui.OpenComponentMsg{Component: resultlistTable.New(m.layout, "Select IP Result File", maxRenderIP, onSelect)}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) ID() ui.ComponentID {
	return m.id
}

func (m *Model) Name() string {
	return m.name
}

func (m *Model) OnClose() tea.Cmd {
	return nil
}

func (m *Model) Mode() env.Mode {
	return env.NormalMode
}
