package targetlist

import (
	"github.com/MohsenBg/bgscan/internal/core/target"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/crud"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	updated, cmd := m.crudTable.Update(msg)
	if table, ok := updated.(*crud.Model[target.FileInfo]); ok {
		m.crudTable = table
	}

	return m, cmd
}
