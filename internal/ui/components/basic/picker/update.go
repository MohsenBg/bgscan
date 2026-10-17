package picker

import (
	tea "charm.land/bubbletea/v2"

	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
)

func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	if _, ok := msg.(tea.WindowSizeMsg); ok {
		m.FilePicker.SetHeight(pickerHeight(m.env))
	}

	var cmds []tea.Cmd

	updated, cmd := m.FilePicker.Update(msg)
	m.FilePicker = updated
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	if didSelect, path := m.FilePicker.DidSelectFile(msg); didSelect {
		if m.OnSelect != nil {
			cmds = append(cmds, m.OnSelect(path))
		}
		cmds = append(cmds, m.CloseCmd())
	}

	return m, tea.Batch(cmds...)
}
