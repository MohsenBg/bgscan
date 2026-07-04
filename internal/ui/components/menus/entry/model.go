package entry

import (
	"bgscan/internal/ui/components/basic/menu"
	"bgscan/internal/ui/components/menus/logs"
	"bgscan/internal/ui/components/menus/settings"
	"bgscan/internal/ui/components/menus/targetsource"
	"bgscan/internal/ui/components/tables/iplist"
	"bgscan/internal/ui/components/tables/outbounds"
	"bgscan/internal/ui/components/tables/resultlist"
	"bgscan/internal/ui/shared/env"
	"bgscan/internal/ui/shared/layout"
	"bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

// Model is the main app model with a menu stack
type Model struct {
	id     ui.ComponentID
	name   string
	menu   ui.Component
	Layout *layout.Layout
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

// New creates a new entry model with main menu and subviews
func New(layout *layout.Layout) *Model {
	entry := &Model{
		id:     ui.NewComponentID(),
		name:   "Main Menu",
		Layout: layout,
		menu:   newMainMenu(layout),
	}
	return entry
}

// Init satisfies Bubble Tea interface
func (m *Model) Init() tea.Cmd {
	return nil
}

// helpers
func newMainMenu(layout *layout.Layout) *menu.Model {
	items := []menu.MenuItem{
		menu.NewMenuItem(
			"▶", "Run Scan", "s",
			func() tea.Msg {
				return ui.OpenComponentMsg{
					Component: targetsource.New(layout),
				}
			},
		),
		menu.NewMenuItem("☰", "IP Files", "i", func() tea.Msg {
			return ui.OpenComponentMsg{
				Component: iplist.New(layout, "IP Files", nil),
			}
		}),
		menu.NewMenuItem("≡", "Result Files", "r", func() tea.Msg {
			var maxRenderIP uint32 = 10_000
			return ui.OpenComponentMsg{
				Component: resultlist.New(layout, "Result Files", maxRenderIP, nil),
			}
		}),
		menu.NewMenuItem("▸", "Xray Outbound", "x", func() tea.Msg {
			return ui.OpenComponentMsg{
				Component: outbounds.New(
					layout,
					"Xray Outbound",
					nil,
				),
			}
		}),
		menu.NewMenuItem("⚙", "Settings", "c", func() tea.Msg {
			return ui.OpenComponentMsg{
				Component: settings.New(layout),
			}
		}),
		menu.NewMenuItem("▤", "Logs", "l", func() tea.Msg {
			return ui.OpenComponentMsg{
				Component: logs.New(layout),
			}
		}),
	}
	return menu.New(items, "Main Menu", layout)
}
