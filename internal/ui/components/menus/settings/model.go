// Package settings exposes grouped configuration inspectors for scanner settings.
package settings

import (
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/menu"
	"github.com/MohsenBg/bgscan/internal/ui/components/inspector/dns"
	"github.com/MohsenBg/bgscan/internal/ui/components/inspector/general"
	"github.com/MohsenBg/bgscan/internal/ui/components/inspector/http"
	"github.com/MohsenBg/bgscan/internal/ui/components/inspector/icmp"
	"github.com/MohsenBg/bgscan/internal/ui/components/inspector/tcp"
	"github.com/MohsenBg/bgscan/internal/ui/components/inspector/uisettings"
	"github.com/MohsenBg/bgscan/internal/ui/components/inspector/xray"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

// Model is the settings menu component.
type Model struct {
	id      ui.ComponentID
	name    string
	session *ui.Session
	menu    ui.Component
}

// New creates the settings menu with configuration sections.
func New(session *ui.Session) *Model {
	items := []menu.MenuItem{
		menu.NewMenuItem("▤", "General Settings", "g", func() tea.Msg {
			return ui.OpenComponentMsg{
				Component: general.New(session, "General Settings"),
			}
		}).WithDescription("Probe rates, concurrency limits and the result writer"),
		menu.NewMenuItem("◈", "ICMP Settings", "i", func() tea.Msg {
			return ui.OpenComponentMsg{
				Component: icmp.New(session, "ICMP Settings"),
			}
		}).WithDescription("Ping sweep workers, timeout and retry counts"),
		menu.NewMenuItem("⇄", "TCP Settings", "t", func() tea.Msg {
			return ui.OpenComponentMsg{
				Component: tcp.New(session, "TCP Settings"),
			}
		}).WithDescription("Target port, workers and retry behaviour"),
		menu.NewMenuItem("◎", "HTTP Settings", "h", func() tea.Msg {
			return ui.OpenComponentMsg{
				Component: http.New(session, "HTTP Settings"),
			}
		}).WithDescription("Host, port, TLS versions and request fingerprint"),
		menu.NewMenuItem("◇", "Xray Settings", "x", func() tea.Msg {
			return ui.OpenComponentMsg{
				Component: xray.New(session, "Xray Settings"),
			}
		}).WithDescription("Speed test thresholds, dial limits and connectivity"),
		menu.NewMenuItem("⌘", "DNS Settings", "d", func() tea.Msg {
			return ui.OpenComponentMsg{
				Component: dns.New(session, "DNS Settings"),
			}
		}).WithDescription("Resolver transport, query types and tunneling limits"),
		menu.NewMenuItem("◐", "UI Settings", "u", func() tea.Msg {
			return ui.OpenComponentMsg{
				Component: uisettings.New(session, "UI Settings"),
			}
		}).WithDescription("Theme, table row limit and log level"),
	}
	return &Model{
		menu:    menu.New(items, "Settings", session.Env()),
		id:      ui.NewComponentID(),
		name:    "settings",
		session: session,
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

func (m *Model) Mode() ui.KeyMode {
	return ui.KeyModeNormal
}
