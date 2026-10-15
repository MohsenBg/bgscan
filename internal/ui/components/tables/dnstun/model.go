package dnstun

import (
	"github.com/MohsenBg/bgscan/internal/core/dns"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/crud"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/form"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/notice"
	"github.com/MohsenBg/bgscan/internal/ui/components/form/dnstt"
	"github.com/MohsenBg/bgscan/internal/ui/components/form/masterdns"
	"github.com/MohsenBg/bgscan/internal/ui/components/form/slipstream"
	"github.com/MohsenBg/bgscan/internal/ui/components/form/stormdns"
	"github.com/MohsenBg/bgscan/internal/ui/components/form/thefeed"
	"github.com/MohsenBg/bgscan/internal/ui/components/form/vaydns"
	"github.com/MohsenBg/bgscan/internal/ui/components/menus/dnstunmenu"
	"github.com/MohsenBg/bgscan/internal/ui/shared/dialog"
	"github.com/MohsenBg/bgscan/internal/ui/shared/keys"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

// Model coordinates outbound configuration additions, list table management,
// and multi-step dialog sequencing paths within the UI stack.
type Model struct {
	id         ui.ComponentID
	name       string
	session    *ui.Session
	crudTable  *crud.Model[dns.DNSTunConfigFile]
	dnsTunMenu ui.Component
}

func New(session *ui.Session, title string, onSelect func(*dns.DNSTunConfigFile) tea.Cmd) *Model {
	m := &Model{
		id:      ui.NewComponentID(),
		name:    "outbounds",
		session: session,
	}

	if onSelect == nil {
		onSelect = m.openDNSTunForm
	}
	m.crudTable = crud.New(session.Env(), newSource(session), newHooks(session, onSelect, m.selectDNSTunMethod), 100)

	return m
}

func (m *Model) Init() tea.Cmd      { return m.crudTable.Init() }
func (m *Model) ID() ui.ComponentID { return m.id }
func (m *Model) Name() string       { return m.name }
func (m *Model) OnClose() tea.Cmd   { return m.crudTable.OnClose() }
func (m *Model) Mode() ui.KeyMode   { return m.crudTable.Mode() }

func (m *Model) selectDNSTunMethod() tea.Cmd {
	protoMenu := dnstunmenu.New(m.session)
	m.dnsTunMenu = protoMenu
	return func() tea.Msg {
		return dialog.OpenDialog(protoMenu,
			dialog.WithTitle(dnstunmenu.Title),
		)
	}
}

// openFormDialog opens a formkit-based form as a dialog. The frame
// title is read from the form itself so "New X" vs "Edit name" stays
// in one place; hints are the shared form hints. Extra options
// (e.g. WithOnClose) still apply after the defaults.
func openFormDialog(session *ui.Session, frm ui.Component, opts ...dialog.DialogOption) dialog.OpenDialogMsg {
	title := ""
	if t, ok := frm.(interface{ Title() string }); ok {
		title = t.Title()
	}
	base := []dialog.DialogOption{
		dialog.WithTitle(title),
		dialog.WithHints(form.Hints(keys.OrDefault(*session.Env().Keys))...),
	}
	return dialog.OpenDialog(frm, append(base, opts...)...)
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
		frm, err := slipstream.New(m.session, original)
		if err != nil {
			return notice.NewNoticeCmd(m.session.Env(), "Error", err.Error(), notice.NOTICE_ERROR)
		}
		return openFormDialog(m.session, frm,
			dialog.WithOnClose(func() tea.Msg { return crud.RefreshMsg{} }))
	}
}

func (m *Model) openDNSTTForm(original *dns.DNSTunConfigFile) tea.Cmd {
	return func() tea.Msg {
		frm, err := dnstt.New(m.session, original)
		if err != nil {
			return notice.NewNoticeCmd(m.session.Env(), "Error", err.Error(), notice.NOTICE_ERROR)
		}
		return openFormDialog(m.session, frm, dialog.WithOnClose(func() tea.Msg { return crud.RefreshMsg{} }))
	}
}

func (m *Model) openVayDNSForm(original *dns.DNSTunConfigFile) tea.Cmd {
	return func() tea.Msg {
		frm, err := vaydns.New(m.session, original)
		if err != nil {
			return notice.NewNoticeCmd(m.session.Env(), "Error", err.Error(), notice.NOTICE_ERROR)
		}
		return openFormDialog(m.session, frm, dialog.WithOnClose(func() tea.Msg { return crud.RefreshMsg{} }))
	}
}

func (m *Model) openMasterDNSForm(original *dns.DNSTunConfigFile) tea.Cmd {
	return func() tea.Msg {
		frm, err := masterdns.New(m.session, original)
		if err != nil {
			return notice.NewNoticeCmd(m.session.Env(), "Error", err.Error(), notice.NOTICE_ERROR)
		}
		return openFormDialog(m.session, frm, dialog.WithOnClose(func() tea.Msg { return crud.RefreshMsg{} }))
	}
}

func (m *Model) openStormDNSForm(original *dns.DNSTunConfigFile) tea.Cmd {
	return func() tea.Msg {
		frm, err := stormdns.New(m.session, original)
		if err != nil {
			return notice.NewNoticeCmd(m.session.Env(), "Error", err.Error(), notice.NOTICE_ERROR)
		}
		return openFormDialog(m.session, frm, dialog.WithOnClose(func() tea.Msg { return crud.RefreshMsg{} }))
	}
}

func (m *Model) openTheFeedForm(original *dns.DNSTunConfigFile) tea.Cmd {
	return func() tea.Msg {
		frm, err := thefeed.New(m.session, original)
		if err != nil {
			return notice.NewNoticeCmd(m.session.Env(), "Error", err.Error(), notice.NOTICE_ERROR)
		}
		return openFormDialog(m.session, frm, dialog.WithOnClose(func() tea.Msg { return crud.RefreshMsg{} }))
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
			frm, err = dnstt.New(m.session, original)
		case dns.DNSTunProtocolVayDNS:
			frm, err = vaydns.New(m.session, original)
		case dns.DNSTunProtocolMasterDNS:
			frm, err = masterdns.New(m.session, original)
		case dns.DNSTunProtocolStormDNS:
			frm, err = stormdns.New(m.session, original)
		case dns.DNSTunProtocolTheFeed:
			frm, err = thefeed.New(m.session, original)
		default:
			frm, err = slipstream.New(m.session, original)
		}

		if err != nil {
			return notice.NewNoticeCmd(m.session.Env(),
				"Error",
				err.Error(),
				notice.NOTICE_ERROR,
			)
		}

		return openFormDialog(m.session, frm,
			dialog.WithOnClose(func() tea.Msg { return crud.RefreshMsg{} }))
	}
}
