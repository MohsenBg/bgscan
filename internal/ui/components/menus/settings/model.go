// Package settings exposes grouped configuration inspectors for scanner settings.
package settings

import (
	"bgscan/internal/ui/components/basic/menu"
	"bgscan/internal/ui/components/inspector/dns"
	"bgscan/internal/ui/components/inspector/general"
	"bgscan/internal/ui/components/inspector/http"
	"bgscan/internal/ui/components/inspector/icmp"
	"bgscan/internal/ui/components/inspector/tcp"
	"bgscan/internal/ui/components/inspector/xray"
	"bgscan/internal/ui/shared/env"
	"bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

// Model is the settings menu component.
type Model struct {
	id    ui.ComponentID
	name  string
	state *ui.AppState
	menu  ui.Component
}

// New creates the settings menu with configuration sections.
func New(state *ui.AppState) *Model {
	items := []menu.MenuItem{
		menu.NewMenuItem("▤", "General Settings", "g", func() tea.Msg {
			return ui.OpenComponentMsg{
				Component: general.New(state, "General Settings"),
			}
		}),
		menu.NewMenuItem("◈", "ICMP Settings", "i", func() tea.Msg {
			return ui.OpenComponentMsg{
				Component: icmp.New(state, "ICMP Settings"),
			}
		}),
		menu.NewMenuItem("⇄", "TCP Settings", "t", func() tea.Msg {
			return ui.OpenComponentMsg{
				Component: tcp.New(state, "TCP Settings"),
			}
		}),
		menu.NewMenuItem("◎", "HTTP Settings", "h", func() tea.Msg {
			return ui.OpenComponentMsg{
				Component: http.New(state, "HTTP Settings"),
			}
		}),
		menu.NewMenuItem("◇", "Xray Settings", "x", func() tea.Msg {
			return ui.OpenComponentMsg{
				Component: xray.New(state, "Xray Settings"),
			}
		}),
		menu.NewMenuItem("⌘", "DNS Settings", "d", func() tea.Msg {
			return ui.OpenComponentMsg{
				Component: dns.New(state, "DNS Settings"),
			}
		}),
	}
	return &Model{
		menu:  menu.New(items, "Settings", state.Layout),
		id:    ui.NewComponentID(),
		name:  "settings",
		state: state,
	}
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
