package dns

import (
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

// Update forwards msg to the inner inspector and returns the updated
// model and any command the inspector produces.
func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	var cmd tea.Cmd
	m.inspector, cmd = m.inspector.Update(msg)
	return m, cmd
}
