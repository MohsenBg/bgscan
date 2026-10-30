package table

import (
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"

	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
	"github.com/MohsenBg/bgscan/internal/ui/theme"
)

func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	m.mu.Lock()
	defer m.mu.Unlock()

	before := captureBubble(m.bubble)
	defer func() {
		if captureBubble(m.bubble) != before {
			m.dirty = true
		}
	}()

	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case theme.ThemeChanged:
		m.applyStylesLocked()

	case spinner.TickMsg:
		if !m.loading {
			return m, nil
		}

		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		m.dirty = true
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

	default:
		// Non-key messages (cursor blink) belong to the filter input
		// while it is active.
		if m.filtering {
			var cmd tea.Cmd
			m.filter, cmd = m.filter.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

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
	case m.env.Keys.Filter.Clear.Matches(msg):
		m.clearFilterLocked()
		return nil, true

	case m.env.Keys.Filter.Accept.Matches(msg):
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

	k := m.env.Keys.Table

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

	case k.Filter.Matches(msg) && m.filterEnabled:
		m.filtering = true
		return m.showCursor(), true

	case m.env.Keys.Filter.Clear.Matches(msg):
		m.clearFilterLocked()
		return nil, true

	default:
		return nil, false
	}

	return nil, true
}

type bubbleState struct {
	cursor  int
	focused bool
	width   int
	height  int
}

func captureBubble(t table.Model) bubbleState {
	return bubbleState{
		cursor:  t.Cursor(),
		focused: t.Focused(),
		width:   t.Width(),
		height:  t.Height(),
	}
}
