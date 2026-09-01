package settings

import (
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

// Update routes messages to the underlying menu component.
func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	var cmd tea.Cmd
	m.menu, cmd = m.menu.Update(msg)
	return m, cmd
}
