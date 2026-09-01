package table

import (
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if cmd := m.Keys.Check(msg); cmd != nil {
			cmds = append(cmds, cmd)
		}

		// Toggle Full Help
		if msg.String() == "?" {
			m.FullHelp = !m.FullHelp
			m.updateTableSizeLocked() // Now safely inside the lock
		}

	case tea.WindowSizeMsg:
		m.updateTableSizeLocked()
		m.BubbleTable.SetStyles(tableStyles())
		return m, nil
	}

	var tableCmd tea.Cmd
	m.BubbleTable, tableCmd = m.BubbleTable.Update(msg)
	if tableCmd != nil {
		cmds = append(cmds, tableCmd)
	}

	return m, tea.Batch(cmds...)
}
