package body

import (
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/confirm"
	"github.com/MohsenBg/bgscan/internal/ui/main/footer"
	"github.com/MohsenBg/bgscan/internal/ui/shared/env"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	lastIdx := len(m.components) - 1

	switch msg := msg.(type) {
	case ui.ThemeChangedMsg:
		var cmds []tea.Cmd
		for i := 0; i < len(m.components); i++ {
			var cmd tea.Cmd
			m.components[i], cmd = m.components[i].Update(msg)
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)

	case tea.KeyPressMsg:
		if env.IsBackKey(msg, m.components[lastIdx].Mode()) && len(m.components) > 1 {
			return m.popComponent()
		}

		if env.IsQuitKey(msg, m.components[lastIdx].Mode()) {
			return m, confirm.ExitConfirmCmd(m.state.Deps())
		}

	case ui.OpenComponentMsg:
		if msg.Component != nil {
			m.state.Log.UI.Info("Opening screen: %s", msg.Component.Name())
			return m.pushComponent(msg.Component)
		}

	case ui.CloseComponentMsg:
		for i, c := range m.components {
			if c.ID() == msg.ID {
				cmd := c.OnClose()
				m.components[i] = nil
				m.components = append(m.components[:i], m.components[i+1:]...)
				return m, cmd
			}
		}

	case ui.ResetComponentStacksMsg:
		var cmds []tea.Cmd
		for i := 1; i < len(m.components); i++ {
			cmd := m.components[i].OnClose()
			m.components[i] = nil
			cmds = append(cmds, cmd)
		}
		m.components = m.components[:1]
		return m, tea.Batch(cmds...)
	}

	activeComp, cmd := m.components[lastIdx].Update(msg)
	m.components[lastIdx] = activeComp

	return m, cmd
}

func (m *Model) pushComponent(c ui.Component) (ui.Component, tea.Cmd) {
	m.components = append(m.components, c)
	return m, tea.Batch(c.Init(), m.forceResize(), m.updateStatusCmd(c.Name()))
}

func (m *Model) popComponent() (ui.Component, tea.Cmd) {
	lastIdx := len(m.components) - 1
	c := m.components[lastIdx]

	closeCmd := c.OnClose()
	m.components[lastIdx] = nil
	m.components = m.components[:lastIdx]

	newTop := m.components[len(m.components)-1]

	return m, tea.Batch(closeCmd, m.updateStatusCmd(newTop.Name()))
}

func (m *Model) updateStatusCmd(name string) tea.Cmd {
	return func() tea.Msg { return footer.UpdateStatus{Status: name} }
}

func (m *Model) forceResize() tea.Cmd {
	return func() tea.Msg {
		return tea.WindowSizeMsg{
			Width:  m.state.Layout.Terminal.Width,
			Height: m.state.Layout.Terminal.Height,
		}
	}
}
