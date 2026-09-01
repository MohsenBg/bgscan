package vaydns

import (
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/form"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	switch msg.(type) {
	case tea.WindowSizeMsg:
		m.calculateSize()
		m.form.SetWidth(m.width)
		m.form.SetHeight(m.height)
	}

	component, cmd := m.form.Update(msg)
	m.form = component.(*form.Model)
	return m, cmd
}
