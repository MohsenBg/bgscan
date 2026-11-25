package resultlist

import (
	"github.com/MohsenBg/bgscan/internal/core/result"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/crud"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

// Update forwards the message to the underlying CRUD table.
func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	updatedCrud, cmd := m.crudTable.Update(msg)
	m.crudTable = updatedCrud.(*crud.Model[result.ResultFile])
	return m, cmd
}
