package picker

import (
	"bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

// Update handles terminal resize, forwards messages to the file picker,
// and closes the overlay after a successful selection.
func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	var cmds []tea.Cmd

	// Handle terminal resize
	if _, ok := msg.(tea.WindowSizeMsg); ok {
		m.FilePicker.SetHeight(pickerHeight(m.Layout))
	}

	// Forward message to Bubble filepicker
	var cmd tea.Cmd
	m.FilePicker, cmd = m.FilePicker.Update(msg)

	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	// Handle file selection
	if didSelect, path := m.FilePicker.DidSelectFile(msg); didSelect && m.OnSelect != nil {
		cmds = append(cmds, m.OnSelect(path))
		cmds = append(cmds, m.CloseCmd())
	}

	return m, tea.Batch(cmds...)
}
