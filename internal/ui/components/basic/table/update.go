package table

import (
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
	case tea.KeyMsg:
		if cmd, handled := m.handleKey(msg); handled {
			return m, cmd
		}
	case tea.WindowSizeMsg:
		m.updateTableSizeLocked()
		m.BubbleTable.SetStyles(m.tableStyles())
	}

	var cmds []tea.Cmd

	var cmd tea.Cmd
	m.BubbleTable, cmd = m.BubbleTable.Update(msg)
	cmds = append(cmds, cmd)

	m.help, cmd = m.help.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// handleFilterKey processes filter input.
// It must be called with m.mu held.
func (m *Model) handleFilterKey(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch {
	case m.Keys.Filter.Clear.Matches(msg):
		m.clearFilterLocked()
		return nil, true
	case m.Keys.Filter.Accept.Matches(msg):
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

	k := m.Keys.Table
	switch {
	case k.Up.Matches(msg):
		m.BubbleTable.MoveUp(1)
	case k.Down.Matches(msg):
		m.BubbleTable.MoveDown(1)
	case k.PageUp.Matches(msg):
		m.BubbleTable.MoveUp(m.BubbleTable.Height())
	case k.PageDown.Matches(msg):
		m.BubbleTable.MoveDown(m.BubbleTable.Height())
	case k.HalfPageUp.Matches(msg):
		m.BubbleTable.MoveUp(max(1, m.BubbleTable.Height()/2))
	case k.HalfPageDown.Matches(msg):
		m.BubbleTable.MoveDown(max(1, m.BubbleTable.Height()/2))
	case k.ChangeSort.Matches(msg):
		m.cycleSortColumnLocked()
	case k.ToggleOrder.Matches(msg):
		m.toggleSortDirectionLocked()
	case k.Filter.Matches(msg):
		m.filtering = true
		return m.showCursor(), true
	case m.Keys.Filter.Clear.Matches(msg):
		m.clearFilterLocked()
		return nil, true
	default:
		return nil, false
	}

	return nil, true
}
