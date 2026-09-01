package dnstun

import (
	"github.com/MohsenBg/bgscan/internal/core/dns"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/crud"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/notice"
	"github.com/MohsenBg/bgscan/internal/ui/components/form/dnstt"
	"github.com/MohsenBg/bgscan/internal/ui/components/form/slipstream"
	"github.com/MohsenBg/bgscan/internal/ui/components/form/vaydns"
	"github.com/MohsenBg/bgscan/internal/ui/components/menus/dnstunmenu"
	"github.com/MohsenBg/bgscan/internal/ui/shared/dialog"
	"github.com/MohsenBg/bgscan/internal/ui/shared/env"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

// Model coordinates outbound configuration additions, list table management,
// and multi-step dialog sequencing paths within the UI stack.
type Model struct {
	id         ui.ComponentID
	name       string
	state      *ui.AppState
	crudTable  *crud.Model[dns.DNSTunConfigFile]
	dnsTunMenu ui.Component
}

// New creates a new outbound template list component view layer.
func New(state *ui.AppState, title string, onSelect func(*dns.DNSTunConfigFile) tea.Cmd) *Model {
	m := &Model{
		id:    ui.NewComponentID(),
		name:  "outbounds",
		state: state,
	}

	if onSelect == nil {
		onSelect = m.openDNSTunForm
	}
	canAdd := true
	m.crudTable = crud.New("dns tunling", state.Layout, newProvider(state, onSelect), 100, canAdd)

	return m
}

func (m *Model) Init() tea.Cmd      { return m.crudTable.Init() }
func (m *Model) ID() ui.ComponentID { return m.id }
func (m *Model) Name() string       { return m.name }
func (m *Model) OnClose() tea.Cmd   { return m.crudTable.OnClose() }
func (m *Model) Mode() env.Mode     { return m.crudTable.Mode() }

func (m *Model) selectDNSTunMethod() tea.Cmd {
	menu := dnstunmenu.New(m.state.Layout)
	m.dnsTunMenu = menu
	return func() tea.Msg {
		return dialog.OpenDialog(menu)
	}
}

func (m *Model) closeDNSTunMenu() tea.Cmd {
	return func() tea.Msg {
		if m.dnsTunMenu == nil {
			return nil
		}

		id := m.dnsTunMenu.ID()
		m.dnsTunMenu = nil
		return ui.CloseComponentMsg{ID: id}
	}
}

func (m *Model) openSlipstreamForm(original *dns.DNSTunConfigFile) tea.Cmd {
	return func() tea.Msg {
		frm, err := slipstream.New(m.state.Layout, m.state, original)
		if err != nil {
			return notice.NewNoticeCmd(m.state.Layout, "Error", err.Error(), notice.NOTICE_ERROR)
		}
		return dialog.OpenDialog(frm,
			dialog.WithOnClose(func() tea.Msg { return crud.MsgRefresh{} }))
	}
}

func (m *Model) openDNSTTForm(original *dns.DNSTunConfigFile) tea.Cmd {
	return func() tea.Msg {
		frm, err := dnstt.New(m.state.Layout, m.state, original)
		if err != nil {
			return notice.NewNoticeCmd(m.state.Layout, "Error", err.Error(), notice.NOTICE_ERROR)
		}
		return dialog.OpenDialog(frm, dialog.WithOnClose(func() tea.Msg { return crud.MsgRefresh{} }))
	}
}

func (m *Model) openVayDNSForm(original *dns.DNSTunConfigFile) tea.Cmd {
	return func() tea.Msg {
		frm, err := vaydns.New(m.state.Layout, m.state, original)
		if err != nil {
			return notice.NewNoticeCmd(m.state.Layout, "Error", err.Error(), notice.NOTICE_ERROR)
		}
		return dialog.OpenDialog(frm, dialog.WithOnClose(func() tea.Msg { return crud.MsgRefresh{} }))
	}
}

func (m *Model) openDNSTunForm(original *dns.DNSTunConfigFile) tea.Cmd {
	return func() tea.Msg {
		var (
			frm ui.Component
			err error
		)

		switch original.Protocol {
		case dns.DNSTunProtocolDNSTT:
			frm, err = dnstt.New(m.state.Layout, m.state, original)
		case dns.DNSTunProtocolVayDNS:
			frm, err = vaydns.New(m.state.Layout, m.state, original)
		default:
			frm, err = slipstream.New(m.state.Layout, m.state, original)
		}

		if err != nil {
			return notice.NewNoticeCmd(
				m.state.Layout,
				"Error",
				err.Error(),
				notice.NOTICE_ERROR,
			)
		}

		return dialog.OpenDialog(frm,
			dialog.WithOnClose(func() tea.Msg { return crud.MsgRefresh{} }))
	}
}
