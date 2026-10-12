package table

import (
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"

	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
)

func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	if ui.HandleTheme(msg, m) {
		return m, nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tea.KeyMsg:
		if m.loading {
			return m, nil
		}

		if cmd, handled := m.handleKey(msg); handled {
			return m, cmd
		}

	case tea.WindowSizeMsg:
		m.updateTableSizeLocked()
		m.bubble.SetStyles(m.tableStyles())
	}

	var cmds []tea.Cmd
	var cmd tea.Cmd

	m.bubble, cmd = m.bubble.Update(msg)
	cmds = append(cmds, cmd)

	m.help, cmd = m.help.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// handleFilterKey processes filter input.
// It must be called with m.mu held.
func (m *Model) handleFilterKey(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch {
	case m.keys.Filter.Clear.Matches(msg):
		m.clearFilterLocked()
		return nil, true

	case m.keys.Filter.Accept.Matches(msg):
		m.filtering = false
		m.hideCursor()
		return nil, true
	}

	var cmd tea.Cmd
	m.filter, cmd = m.filter.Update(msg)
	m.refreshDisplayLocked(false)

	return cmd, true
}

// handleKey processes table navigation and filter toggling.
// It must be called with m.mu held.
func (m *Model) handleKey(msg tea.KeyMsg) (tea.Cmd, bool) {
	if m.filtering {
		return m.handleFilterKey(msg)
	}

	k := m.keys.Table

	switch {
	case k.Up.Matches(msg):
		m.bubble.MoveUp(1)

	case k.Down.Matches(msg):
		m.bubble.MoveDown(1)

	case k.PageUp.Matches(msg):
		m.bubble.MoveUp(m.bubble.Height())

	case k.PageDown.Matches(msg):
		m.bubble.MoveDown(m.bubble.Height())

	case k.HalfPageUp.Matches(msg):
		m.bubble.MoveUp(max(1, m.bubble.Height()/2))

	case k.HalfPageDown.Matches(msg):
		m.bubble.MoveDown(max(1, m.bubble.Height()/2))

	case k.ChangeSort.Matches(msg):
		m.cycleSortColumnLocked()

	case k.ToggleOrder.Matches(msg):
		m.toggleSortDirectionLocked()

	case k.Filter.Matches(msg):
		m.filtering = true
		return m.showCursor(), true

	case m.keys.Filter.Clear.Matches(msg):
		m.clearFilterLocked()
		return nil, true

	default:
		return nil, false
	}

	return nil, true
}
