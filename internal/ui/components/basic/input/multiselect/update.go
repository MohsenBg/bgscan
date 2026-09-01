package multiselect

import (
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
)

// Update processes incoming Bubble Tea messages and updates the state
// of the multi-select component.
func (m *Model[T]) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	prev := make([]T, len(m.value))
	copy(prev, m.value)

	updated, cmd := m.huhInput.Update(msg)
	if ms, ok := updated.(*huh.MultiSelect[T]); ok {
		m.huhInput = ms
	}

	if !slicesEqual(m.value, prev) && m.onChange != nil {
		cmd = tea.Batch(cmd, m.onChange(m.value))
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		inp := m.huhInput.WithWidth(m.Width())
		m.huhInput = inp.(*huh.MultiSelect[T])

		return m, cmd
	case tea.KeyPressMsg:
		if msg.Code == tea.KeyEnter {
			return m, tea.Batch(cmd, m.submit())
		}
	}

	return m, cmd
}
