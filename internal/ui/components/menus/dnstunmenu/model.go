// Package dnstunmenu selects a DNS tunnel protocol.
package dnstunmenu

import (
	"github.com/MohsenBg/bgscan/internal/core/dns"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/menu"
	"github.com/MohsenBg/bgscan/internal/ui/shared/env"
	"github.com/MohsenBg/bgscan/internal/ui/shared/layout"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

// MsgSelectProtocol is fired when a protocol is selected.
type MsgSelectProtocol struct {
	Protocol dns.DNSTunProtocol
}

// Model is the DNS tunnel protocol selection menu.
type Model struct {
	id     ui.ComponentID
	name   string
	menu   ui.Component
	Layout *layout.Layout
}

// New creates the DNS tunnel protocol menu.
func New(layout *layout.Layout) *Model {
	m := &Model{
		id:     ui.NewComponentID(),
		name:   "DNS Tunnel Menu",
		Layout: layout,
	}
	m.menu = newMenu(layout)
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

func (m *Model) Mode() env.Mode {
	return env.NormalMode
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func newMenu(layout *layout.Layout) *menu.Model {
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
	}

	return menu.New(
		items, "Select Protocol", layout,
		menu.WithHeight(14),
		menu.WithWidth(40),
	)
}
