package tabs

import (
	"bgscan/internal/ui/shared/env"
	"bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

func (m *Model[T]) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == env.KeyTab {
			m.NextTab()
			cmd = m.selectTabCmd()
		}
		if msg.String() == env.KeyShiftTab {
			m.BackTab()
			cmd = m.selectTabCmd()
		}
	}
	return m, cmd
}
