package general

import (
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

// Update passes the message to the wrapped inspector and returns its command.
func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	var cmd tea.Cmd
	m.inspector, cmd = m.inspector.Update(msg)
	return m, cmd
}
