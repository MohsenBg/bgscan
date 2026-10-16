package form

import (
	tea "charm.land/bubbletea/v2"

	"github.com/MohsenBg/bgscan/internal/ui/components/basic/inspector"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
)

func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.recalculateSize()

	case tea.KeyPressMsg:
		if cmd, handled := m.handleKey(msg); handled {
			return m, cmd
		}
	}

	if m.inspector == nil {
		return m, nil
	}

	component, cmd := m.inspector.Update(msg)
	if im, ok := component.(*inspector.Model); ok {
		m.inspector = im
	}

	return m, cmd
}

func (m *Model) handleKey(msg tea.KeyMsg) (tea.Cmd, bool) {
	f := m.env.Keys.Form

	if !m.formKeysActive() {
		return nil, false
	}

	if f.Save.Matches(msg) {
		return m.Save(), true
	}

	if f.Cancel.Matches(msg) {
		return m.Cancel(), true
	}

	return nil, false
}
