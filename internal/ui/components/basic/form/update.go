package form

import (
	"github.com/MohsenBg/bgscan/internal/logger"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/inspector"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.recalculateSize()

	case tea.KeyPressMsg:
		switch msg.String() {
		case "s":
			cmd = m.Save()

		case "esc", "b", "q":
			cmd = m.Cancel()
			logger.DebugInfo("Form cancelled")
		}
	}

	component, inspectorCmd := m.inspector.Update(msg)
	m.inspector = component.(*inspector.Model)
	return m, tea.Batch(cmd, inspectorCmd)
}
