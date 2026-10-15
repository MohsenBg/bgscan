package outbounds

import (
	tea "charm.land/bubbletea/v2"
	"github.com/MohsenBg/bgscan/internal/core/xray"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/crud"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/menu"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/picker"
	"github.com/MohsenBg/bgscan/internal/ui/components/menus/outboundmenu"
	"github.com/MohsenBg/bgscan/internal/ui/shared/dialog"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
)

type openAddMenuMsg struct{}

const dialogTitle = "Add Outbound"

func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case openAddMenuMsg:
		m.outboundMenu = outboundmenu.New(m.session)
		return m, func() tea.Msg {
			return dialog.OpenDialog(
				m.outboundMenu,
				dialog.WithTitle(dialogTitle),
				dialog.WithStyle(dialog.StyleInfo),
				dialog.WithOffset(0, -5),
				dialog.WithHints(menu.DialogHints(m.session.Env().Keys.Menu)...),
			)
		}

	case outboundmenu.MsgSelectImportMethod:
		cmds := []tea.Cmd{m.closeMenuCmd()}
		switch msg.Method {
		case outboundmenu.MethodJSON:
			cmds = append(cmds, picker.OpenFilePickerCmd(
				m.session.Env(),
				"Select outbound template (.json)",
				"",
				[]string{".json"},
				m.handleFileSelect,
			))
		case outboundmenu.MethodLink:
			cmds = append(cmds, m.handleLinkImport())
		}
		return m, tea.Batch(cmds...)
	}

	updated, cmd := m.crudTable.Update(msg)
	if table, ok := updated.(*crud.Model[xray.OutboundFile]); ok {
		m.crudTable = table
	}
	return m, cmd
}
