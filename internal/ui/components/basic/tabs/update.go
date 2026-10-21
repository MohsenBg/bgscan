package tabs

import (
	tea "charm.land/bubbletea/v2"

	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
	"github.com/MohsenBg/bgscan/internal/ui/theme"
)

func (m *Model[T]) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case theme.ThemeChanged:
		m.rebuild()
		return m, nil

	case frameMsg:
		m.step()
		return m, m.tick()

	case tea.KeyMsg:
		return m, m.handleKey(msg)
	}

	return m, nil
}

func (m *Model[T]) handleKey(msg tea.KeyMsg) tea.Cmd {
	switch {
	case m.env.Keys.Tab.Next.Matches(msg):
		m.NextTab()
	case m.env.Keys.Tab.Prev.Matches(msg):
		m.BackTab()
	default:
		return nil
	}

	return tea.Batch(m.selectTabCmd(), m.tick())
}
