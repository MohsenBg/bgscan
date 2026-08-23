package dnstun

import (
	"bgscan/internal/core/dns"
	"bgscan/internal/ui/components/basic/crud"
	"bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	updated, cmd := m.crudTable.Update(msg)

	if table, ok := updated.(*crud.Model[dns.DNSTunConfigFile]); ok {
		m.crudTable = table
	}
	return m, cmd
}
