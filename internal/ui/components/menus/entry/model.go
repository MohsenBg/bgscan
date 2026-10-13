// Package entry is the root main-menu component that opens app sections.
package entry

import (
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/menu"
	"github.com/MohsenBg/bgscan/internal/ui/components/menus/logs"
	"github.com/MohsenBg/bgscan/internal/ui/components/menus/settings"
	"github.com/MohsenBg/bgscan/internal/ui/components/menus/targetsource"
	"github.com/MohsenBg/bgscan/internal/ui/components/tables/dnstun"
	"github.com/MohsenBg/bgscan/internal/ui/components/tables/iplist"
	"github.com/MohsenBg/bgscan/internal/ui/components/tables/outbounds"
	"github.com/MohsenBg/bgscan/internal/ui/components/tables/resultlist"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

// Model is the main entry menu component.
type Model struct {
	id      ui.ComponentID
	name    string
	menu    ui.Component
	session *ui.Session
}

// New creates the main entry menu with available sections.
func New(session *ui.Session) *Model {
	entry := &Model{
		id:      ui.NewComponentID(),
		name:    "Main Menu",
		session: session,
		menu:    newMainMenu(session),
	}
	return entry
}

// Init starts the main menu, which restarts its entrance animation.
func (m *Model) Init() tea.Cmd {
	return m.menu.Init()
}

func (m *Model) ID() ui.ComponentID {
	return m.id
}

func (m *Model) Name() string {
	return m.name
}

// OnClose is a no-op for the entry menu.
func (m *Model) OnClose() tea.Cmd {
	return nil
}

func (m *Model) Mode() ui.KeyMode {
	return ui.KeyModeNormal
}

func newMainMenu(sesstion *ui.Session) *menu.Model {
	items := []menu.MenuItem{
		menu.NewMenuItem(
			"▶",
			"Run Scan",
			"s",
			func() tea.Msg {
				return ui.OpenComponentMsg{
					Component: targetsource.New(sesstion),
				}
			},
		).WithDescription("Start a new scan from an IP list or a saved result"),
		menu.NewMenuItem(
			"▣",
			"IP Files",
			"i",
			func() tea.Msg {
				return ui.OpenComponentMsg{
					Component: iplist.New(sesstion, "IP Files", nil),
				}
			},
		).WithDescription("Manage the IP ranges loaded for scanning"),
		menu.NewMenuItem(
			"◆",
			"Result Files",
			"r",
			func() tea.Msg {
				var maxRenderIP uint32 = 10_000
				return ui.OpenComponentMsg{
					Component: resultlist.New(
						sesstion,
						"Result Files",
						maxRenderIP,
						nil,
					),
				}
			},
		).WithDescription("Browse and export previous scan results"),
		menu.NewMenuItem(
			"→",
			"Xray Outbound",
			"x",
			func() tea.Msg {
				return ui.OpenComponentMsg{
					Component: outbounds.New(
						sesstion,
						"Xray Outbound",
						nil,
					),
				}
			},
		).WithDescription("Import and manage Xray outbound configs"),
		menu.NewMenuItem(
			"≈",
			"DNS Tunneling",
			"d",
			func() tea.Msg {
				return ui.OpenComponentMsg{
					Component: dnstun.New(
						sesstion,
						"DNS Tunneling",
						nil,
					),
				}
			},
		).WithDescription("Configure DNS tunneling clients"),
		menu.NewMenuItem(
			"⚙",
			"Settings",
			"c",
			func() tea.Msg {
				return ui.OpenComponentMsg{
					Component: settings.New(sesstion),
				}
			},
		).WithDescription("Theme, table and scanner preferences"),
		menu.NewMenuItem(
			"≡",
			"Logs",
			"l",
			func() tea.Msg {
				return ui.OpenComponentMsg{
					Component: logs.New(sesstion),
				}
			},
		).WithDescription("Inspect application and scanner logs"),
	}
	return menu.New(items, "Main Menu", sesstion.Env(), menu.WithBackHint(false))
}
