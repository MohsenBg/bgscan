package notice

import (
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

// Update closes the notice on Enter and forwards all messages to the internal
// viewport so long messages remain scrollable.
func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.env.Keys.Notice.Continue.Matches(msg) {
			return m, m.CloseCmd()
		}
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)

	return m, cmd
}
