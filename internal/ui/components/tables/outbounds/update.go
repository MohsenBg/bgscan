package outbounds

import (
	"bgscan/internal/core/xray"
	"bgscan/internal/ui/components/basic/crud"
	"bgscan/internal/ui/components/basic/picker"
	"bgscan/internal/ui/components/menus/outboundmenu"
	"bgscan/internal/ui/shared/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	switch msg := msg.(type) {

	// Catch the clean hotkey notification intercepted from the inner controller
	case crud.MsgActionTrigger:
		if msg.ActionType == "add" {
			return m, m.ShowAdditionMethod()
		}

	case outboundmenu.MsgSelectImportMethod:
		if msg.Method == outboundmenu.MethodJSON {
			return m,
				tea.Sequence(m.closeOutboundMenu(),
					picker.NewOpenPickFileCmd(
						m.layout,
						"Select outbound template (.json)",
						"",
						[]string{".json"},
						m.handleFileSelect,
					))
		}

		if msg.Method == outboundmenu.MethodLink {
			return m, tea.Sequence(m.closeOutboundMenu(), m.handleLinkImport())
		}
	}

	updatedCrud, cmd := m.crudTable.Update(msg)
	m.crudTable = updatedCrud.(*crud.Model[xray.XrayOutboundsFile])
	return m, cmd
}
