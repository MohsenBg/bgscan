package dnstun

import (
	"github.com/MohsenBg/bgscan/internal/core/dns"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/crud"
	"github.com/MohsenBg/bgscan/internal/ui/components/menus/dnstunmenu"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case crud.MsgActionTrigger:
		if msg.ActionType == "add" {
			return m, m.selectDNSTunMethod()
		}

	case dnstunmenu.MsgSelectProtocol:
		if m.dnsTunMenu != nil {
			switch msg.Protocol {
			case dns.DNSTunProtocolSlipstream:
				return m, tea.Sequence(
					m.closeDNSTunMenu(),
					m.openSlipstreamForm(nil),
				)
			case dns.DNSTunProtocolDNSTT:
				return m, tea.Sequence(
					m.closeDNSTunMenu(),
					m.openDNSTTForm(nil),
				)
			case dns.DNSTunProtocolVayDNS:
				return m, tea.Sequence(
					m.closeDNSTunMenu(),
					m.openVayDNSForm(nil),
				)
			default:
				return m, m.closeDNSTunMenu()
			}
		}
	}

	updated, cmd := m.crudTable.Update(msg)

	if table, ok := updated.(*crud.Model[dns.DNSTunConfigFile]); ok {
		m.crudTable = table
	}
	return m, cmd
}
