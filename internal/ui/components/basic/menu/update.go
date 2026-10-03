package menu

import (
	"github.com/MohsenBg/bgscan/internal/ui/shared/env"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

// Update handles input and animation ticks.
func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	if ui.HandleTheme(msg, m) {
		return m, m.nextFrame()
	}

	if key, ok := msg.(tea.KeyPressMsg); ok {
		if cmd, handled := m.handleKey(key); handled {
			return m, cmd
		}
	}

	// Ticks and unhandled keys just advance the animation.
	return m, m.nextFrame()
}

// handleKey handles a key press and reports whether it was consumed.
func (m *Model) handleKey(msg tea.KeyPressMsg) (tea.Cmd, bool) {
	key := msg.String()
	switch key {
	case env.KeyEnter:
		if item, ok := m.GetSelected(); ok {
			return m.selectItem(item), true
		}
	case "q", env.KeyEsc:
		return m.nextFrame(), true
	default:
		if item, ok := m.itemForShortcut(key); ok {
			return m.selectItem(item), true
		}
		m.navigate(key)
	}
	return nil, false
}

func (m *Model) navigate(key string) {
	switch key {
	case "up", "k":
		m.setSelection(m.index - 1)
	case "down", "j":
		m.setSelection(m.index + 1)
	case "pgup":
		m.setSelection(m.index - m.visibleItems())
	case "pgdown":
		m.setSelection(m.index + m.visibleItems())
	case "home", "g":
		m.setSelection(0)
	case "end", "G":
		m.setSelection(len(m.items) - 1)
	}
}

func (m *Model) setSelection(to int) {
	if len(m.items) == 0 {
		return
	}
	m.index = max(0, min(to, len(m.items)-1))
}

// itemForShortcut finds the item bound to key and moves the selection to it.
func (m *Model) itemForShortcut(key string) (MenuItem, bool) {
	for i, item := range m.items {
		if item.shortcut != "" && item.shortcut == key {
			m.setSelection(i)
			return item, true
		}
	}
	return MenuItem{}, false
}

func (m *Model) selectItem(item MenuItem) tea.Cmd {
	if m.onSelect != nil {
		return m.onSelect(item)
	}
	return item.action
}
