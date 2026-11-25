// Package dnstunmenu is the menu that lets the user pick a DNS tunnel
// protocol. It wraps a basic menu and reports the choice to its caller
// with a MsgSelectProtocol message.
package dnstunmenu

import (
	"github.com/MohsenBg/bgscan/internal/core/dns"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/menu"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

// MsgSelectProtocol reports the DNS tunnel protocol the user selected.
type MsgSelectProtocol struct {
	Protocol dns.DNSTunProtocol
}

// Model is the DNS tunnel protocol selection menu component.
type Model struct {
	id      ui.ComponentID
	name    string
	menu    ui.Component
	session *ui.Session
}

// New creates the DNS tunnel protocol menu for the given session.
func New(session *ui.Session) *Model {
	m := &Model{
		id:      ui.NewComponentID(),
		name:    "DNS Tunnel Menu",
		session: session,
	}
	m.menu = newMenu(session)
	return m
}

// ID returns the unique identifier of the component.
func (m *Model) ID() ui.ComponentID {
	return m.id
}

// Name returns the display name of the component.
func (m *Model) Name() string {
	return m.name
}

// OnClose runs when the component closes and returns no command.
func (m *Model) OnClose() tea.Cmd {
	return nil
}

// Mode returns the key mode the component uses while it is active.
func (m *Model) Mode() ui.KeyMode {
	return ui.KeyModeNormal
}

// Init starts the inner menu and returns its first command.
func (m *Model) Init() tea.Cmd {
	return m.menu.Init()
}

// Title is the dialog title shown at the top of the menu.
const Title = "Select Protocol"

func newMenu(session *ui.Session) *menu.Model {
	items := []menu.MenuItem{
		menu.NewMenuItem(
			"≈",
			"DNSTT",
			"d",
			func() tea.Msg {
				return MsgSelectProtocol{Protocol: dns.DNSTunProtocolDNSTT}
			},
		),
		menu.NewMenuItem(
			"~",
			"Slipstream",
			"s",
			func() tea.Msg {
				return MsgSelectProtocol{Protocol: dns.DNSTunProtocolSlipstream}
			},
		),
		menu.NewMenuItem(
			"◆",
			"VayDNS",
			"v",
			func() tea.Msg {
				return MsgSelectProtocol{Protocol: dns.DNSTunProtocolVayDNS}
			},
		),
		menu.NewMenuItem(
			"◈",
			"MasterDNS",
			"m",
			func() tea.Msg {
				return MsgSelectProtocol{Protocol: dns.DNSTunProtocolMasterDNS}
			},
		),
		menu.NewMenuItem(
			"◇",
			"StormDNS",
			"t",
			func() tea.Msg {
				return MsgSelectProtocol{Protocol: dns.DNSTunProtocolStormDNS}
			},
		),
		menu.NewMenuItem(
			"♨",
			"TheFeed",
			"f",
			func() tea.Msg {
				return MsgSelectProtocol{Protocol: dns.DNSTunProtocolTheFeed}
			},
		),
	}

	return menu.New(
		items, Title, session.Env(),
		menu.WithHeight(20),
		menu.WithWidth(40),
	)
}
