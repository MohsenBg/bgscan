package resultlist

import (
	"bgscan/internal/core/result"
	"bgscan/internal/ui/components/basic/crud"
	"bgscan/internal/ui/shared/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	updatedCrud, cmd := m.crudTable.Update(msg)
	m.crudTable = updatedCrud.(*crud.Model[result.ResultFile])
	return m, cmd
}
