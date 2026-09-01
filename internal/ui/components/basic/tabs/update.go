package tabs

import (
	"github.com/MohsenBg/bgscan/internal/ui/shared/env"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

// Update handles key events to switch tabs.
func (m *Model[T]) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case env.KeyTab:
			m.NextTab()
			cmd = m.selectTabCmd()
		case env.KeyShiftTab:
			m.BackTab()
			cmd = m.selectTabCmd()
		}
	}
	return m, cmd
}
