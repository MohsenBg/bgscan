package iplist

import (
	"bgscan/internal/core/iplist"
	"bgscan/internal/ui/components/basic/crud"
	"bgscan/internal/ui/components/basic/picker"
	"bgscan/internal/ui/shared/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// Update intercepts framework runtime messages and forwards state signals down to active widgets.
func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case crud.MsgActionTrigger:
		if msg.ActionType == "add" {
			return m, picker.NewOpenPickFileCmd(
				m.layout,
				"Select IP File (.txt)",
				"",
				[]string{".txt"},
				m.handleFileSelect,
			)
		}
	}

	updatedCrud, cmd := m.crudTable.Update(msg)
	m.crudTable = updatedCrud.(*crud.Model[iplist.IPFileInfo])
	return m, cmd
}
