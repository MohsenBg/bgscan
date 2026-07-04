package selectinput

import (
	"bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
)

// Update processes incoming Bubble Tea messages and updates the state
// of the select component.
func (m *Model[T]) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	prev := m.value

	updated, cmd := m.huhInput.Update(msg)
	if sel, ok := updated.(*huh.Select[T]); ok {
		m.huhInput = sel
	}

	if m.value != prev && m.onChange != nil {
		cmd = tea.Batch(cmd, m.onChange(m.value))
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		inp := m.huhInput.WithWidth(m.Width())
		m.huhInput = inp.(*huh.Select[T])

		return m, cmd
	case tea.KeyPressMsg:
		if msg.Code == tea.KeyEnter {
			return m, tea.Batch(cmd, m.submit())
		}
	}

	return m, cmd
}
