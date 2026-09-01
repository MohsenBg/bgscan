package outbounds

import (
	"github.com/MohsenBg/bgscan/internal/core/xray"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/crud"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/picker"
	"github.com/MohsenBg/bgscan/internal/ui/components/menus/outboundmenu"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case crud.MsgActionTrigger:
		if msg.ActionType == "add" {
			return m, m.ShowAdditionMethod()
		}

	case outboundmenu.MsgSelectImportMethod:
		switch msg.Method {
		case outboundmenu.MethodJSON:
			return m, tea.Sequence(
				m.closeOutboundMenu(),
				picker.OpenFilePickerCmd(
					m.layout,
					"Select outbound template (.json)",
					"",
					[]string{".json"},
					m.handleFileSelect,
				),
			)

		case outboundmenu.MethodLink:
			return m, tea.Sequence(
				m.closeOutboundMenu(),
				m.handleLinkImport(),
			)
		}
	}

	updated, cmd := m.crudTable.Update(msg)

	if table, ok := updated.(*crud.Model[xray.XrayOutboundsFile]); ok {
		m.crudTable = table
	}

	return m, cmd
}
