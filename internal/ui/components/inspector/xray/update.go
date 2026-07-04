package xray

import (
	"bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	var cmd tea.Cmd
	m.inspector, cmd = m.inspector.Update(msg)
	return m, cmd
}
