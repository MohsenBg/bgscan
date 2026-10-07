package menu

import (
	tea "charm.land/bubbletea/v2"

	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
)

func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	if ui.HandleTheme(msg, m) {
		if m.help != nil {
			m.help, _ = m.help.Update(msg)
		}

		return m, m.nextFrame()
	}

	switch msg := msg.(type) {
	case menuTickMsg:
		return m, m.nextFrame()

	case tea.WindowSizeMsg:
		m.handleResize()
		return m, m.nextFrame()

	case tea.KeyMsg:
		if cmd, handled := m.handleKey(msg); handled {
			return m, cmd
		}
	}

	var helpCmd tea.Cmd

	if m.help != nil {
		m.help, helpCmd = m.help.Update(msg)
	}

	next := m.nextFrame()

	switch {
	case next == nil:
		return m, helpCmd
	case helpCmd == nil:
		return m, next
	default:
		return m, tea.Sequence(next, helpCmd)
	}
}

// handleKey handles a key press and reports whether it was consumed.
func (m *Model) handleKey(key tea.KeyMsg) (tea.Cmd, bool) {
	if m.keys == nil {
		return nil, false
	}

	if m.keys.Menu.Select.Matches(key) {
		if item, ok := m.getSelected(); ok {
			return m.selectItem(item), true
		}

		return nil, true
	}

	if item, ok := m.itemForShortcut(key); ok {
		return m.selectItem(item), true
	}

	if m.navigate(key) {
		return nil, true
	}

	return nil, false
}

func (m *Model) navigate(key tea.KeyMsg) bool {
	if m.keys == nil {
		return false
	}

	nav := &m.keys.Nav

	switch {
	case nav.Up.Matches(key):
		m.setSelection(m.index - 1)

	case nav.Down.Matches(key):
		m.setSelection(m.index + 1)

	case nav.PageUp.Matches(key):
		m.setSelection(m.index - m.visibleItems())

	case nav.PageDown.Matches(key):
		m.setSelection(m.index + m.visibleItems())

	case nav.Top.Matches(key):
		m.setSelection(0)

	case nav.Bottom.Matches(key):
		m.setSelection(len(m.items) - 1)

	default:
		return false
	}

	return true
}

func (m *Model) setSelection(to int) {
	if len(m.items) == 0 {
		return
	}

	m.index = max(0, min(to, len(m.items)-1))
}

// itemForShortcut finds the item bound to key and moves selection to it.
func (m *Model) itemForShortcut(msg tea.KeyMsg) (MenuItem, bool) {
	key := msg.String()

	for i, item := range m.items {
		if item.shortcut != "" && item.shortcut == key {
			m.setSelection(i)
			return item, true
		}
	}

	return MenuItem{}, false
}

func (m *Model) getSelected() (MenuItem, bool) {
	return m.GetSelected()
}

func (m *Model) selectItem(item MenuItem) tea.Cmd {
	if m.onSelect != nil {
		return m.onSelect(item)
	}

	return item.action
}
