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
	"bgscan/internal/ui/shared/layout"
	"bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

type Model struct {
	id     ui.ComponentID
	name   string
	Layout *layout.Layout
	menu   ui.Component
}

func New(layout *layout.Layout) *Model {
	items := []menu.MenuItem{
		menu.NewMenuItem("▤", "General Settings", "g", func() tea.Msg {
			return ui.OpenComponentMsg{
				Component: general.New(layout, "General Settings"),
			}
		}),
		menu.NewMenuItem("◈", "ICMP Settings", "i", func() tea.Msg {
			return ui.OpenComponentMsg{
				Component: icmp.New(layout, "ICMP Settings"),
			}
		}),
		menu.NewMenuItem("⇄", "TCP Settings", "t", func() tea.Msg {
			return ui.OpenComponentMsg{
				Component: tcp.New(layout, "TCP Settings"),
			}
		}),
		menu.NewMenuItem("◎", "HTTP Settings", "h", func() tea.Msg {
			return ui.OpenComponentMsg{
				Component: http.New(layout, "HTTP Settings"),
			}
		}),
		menu.NewMenuItem("◇", "XRay Settings", "x", func() tea.Msg {
			return ui.OpenComponentMsg{
				Component: xray.New(layout, "XRay Settings"),
			}
		}),
		menu.NewMenuItem("⌘", "DNS Settings", "d", func() tea.Msg {
			return ui.OpenComponentMsg{
				Component: dns.New(layout, "DNS Settings"),
			}
		}),
	}
	return &Model{
		menu:   menu.New(items, "Settings", layout),
		id:     ui.NewComponentID(),
		name:   "settings",
		Layout: layout,
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
