// Package dnstunmenu selects a DNS tunnel protocol.
package dnstunmenu

import (
	"github.com/MohsenBg/bgscan/internal/core/dns"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/menu"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

// MsgSelectProtocol is fired when a protocol is selected.
type MsgSelectProtocol struct {
	Protocol dns.DNSTunProtocol
}

// Model is the DNS tunnel protocol selection menu.
type Model struct {
	id      ui.ComponentID
	name    string
	menu    ui.Component
	session *ui.Session
}

// New creates the DNS tunnel protocol menu.
func New(session *ui.Session) *Model {
	m := &Model{
		id:      ui.NewComponentID(),
		name:    "DNS Tunnel Menu",
		session: session,
	}
	m.menu = newMenu(session)
	return m
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

func (m *Model) Init() tea.Cmd {
	return m.menu.Init()
}

// Title is the dialog-frame title for the DNS tunnel protocol menu.
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
