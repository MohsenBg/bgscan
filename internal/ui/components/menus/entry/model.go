// Package entry is the root main-menu component that opens app sections.
package entry

import (
	"bgscan/internal/ui/components/basic/menu"
	"bgscan/internal/ui/components/menus/logs"
	"bgscan/internal/ui/components/menus/settings"
	"bgscan/internal/ui/components/menus/targetsource"
	"bgscan/internal/ui/components/tables/dnstun"
	"bgscan/internal/ui/components/tables/iplist"
	"bgscan/internal/ui/components/tables/outbounds"
	"bgscan/internal/ui/components/tables/resultlist"
	"bgscan/internal/ui/shared/env"
	"bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

// Model is the main entry menu component.
type Model struct {
	id    ui.ComponentID
	name  string
	menu  ui.Component
	state *ui.AppState
}

// New creates the main entry menu with available sections.
func New(state *ui.AppState) *Model {
	entry := &Model{
		id:    ui.NewComponentID(),
		name:  "Main Menu",
		state: state,
		menu:  newMainMenu(state),
	}
	return entry
}

// Init is a no-op for the entry menu.
func (m *Model) Init() tea.Cmd {
	return nil
}

// ID returns the component identifier.
func (m *Model) ID() ui.ComponentID {
	return m.id
}

// Name returns the component name.
func (m *Model) Name() string {
	return m.name
}

// OnClose is a no-op for the entry menu.
func (m *Model) OnClose() tea.Cmd {
	return nil
}

// Mode returns the entry menu interaction mode.
func (m *Model) Mode() env.Mode {
	return env.NormalMode
}

func newMainMenu(state *ui.AppState) *menu.Model {
	items := []menu.MenuItem{
		menu.NewMenuItem(
			"▶",
			"Run Scan",
			"s",
			func() tea.Msg {
				return ui.OpenComponentMsg{
					Component: targetsource.New(state),
				}
			},
		),
		menu.NewMenuItem(
			"▣",
			"IP Files",
			"i",
			func() tea.Msg {
				return ui.OpenComponentMsg{
					Component: iplist.New(state.Layout, "IP Files", nil),
				}
			},
		),
		menu.NewMenuItem(
			"◆",
			"Result Files",
			"r",
			func() tea.Msg {
				var maxRenderIP uint32 = 10_000
				return ui.OpenComponentMsg{
					Component: resultlist.New(
						state,
						"Result Files",
						maxRenderIP,
						nil,
					),
				}
			},
		),
		menu.NewMenuItem(
			"→",
			"Xray Outbound",
			"x",
			func() tea.Msg {
				return ui.OpenComponentMsg{
					Component: outbounds.New(
						state.Layout,
						"Xray Outbound",
						nil,
					),
				}
			},
		),
		menu.NewMenuItem(
			"≈",
			"DNS Tunneling",
			"d",
			func() tea.Msg {
				return ui.OpenComponentMsg{
					Component: dnstun.New(
						state,
						"DNS Tunneling",
						nil,
					),
				}
			},
		),
		menu.NewMenuItem(
			"⚙",
			"Settings",
			"c",
			func() tea.Msg {
				return ui.OpenComponentMsg{
					Component: settings.New(state),
				}
			},
		),
		menu.NewMenuItem(
			"≡",
			"Logs",
			"l",
			func() tea.Msg {
				return ui.OpenComponentMsg{
					Component: logs.New(state),
				}
			},
		),
	}

	return menu.New(items, "Main Menu", state.Layout)
}
