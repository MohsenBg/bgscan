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
		switch m.components[lastIdx].Mode() {
		case env.NormalMode:
			if len(m.components) > 1 && m.state.Deps().Keys.Global.Back.Matches(msg) {
				return m.popComponent()
			}
			if m.state.Deps().Keys.Global.Quit.Matches(msg) {
				return m, confirm.ExitConfirmCmd(m.state.Deps())
			}
		case env.InputMode:
			if m.state.Deps().Keys.Global.BackFromInput.Matches(msg) {
				return m.popComponent()
			}
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
				return m, tea.Batch(cmd, m.headerVisibilityCmd(), m.updateStatusCmd(m.topName()))
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
		return m, tea.Batch(append(cmds, m.headerVisibilityCmd(), m.updateStatusCmd(m.topName()))...)
	}

	activeComp, cmd := m.components[lastIdx].Update(msg)
	m.components[lastIdx] = activeComp

	return m, cmd
}

func (m *Model) pushComponent(c ui.Component) (ui.Component, tea.Cmd) {
	m.components = append(m.components, c)
	return m, tea.Batch(c.Init(), m.forceResize(), m.updateStatusCmd(c.Name()), m.headerVisibilityCmd())
}

func (m *Model) popComponent() (ui.Component, tea.Cmd) {
	lastIdx := len(m.components) - 1
	c := m.components[lastIdx]

	closeCmd := c.OnClose()
	m.components[lastIdx] = nil
	m.components = m.components[:lastIdx]

	newTop := m.components[len(m.components)-1]

	return m, tea.Batch(closeCmd, m.updateStatusCmd(newTop.Name()), m.headerVisibilityCmd())
}

func (m *Model) topName() string {
	if len(m.components) == 0 {
		return ""
	}
	return m.components[len(m.components)-1].Name()
}

// headerVisibilityCmd requests the workspace to show the header only on
// the main menu screen.
func (m *Model) headerVisibilityCmd() tea.Cmd {
	visible := m.topName() == "Main Menu"
	return func() tea.Msg { return ui.HeaderVisibilityMsg{Visible: visible} }
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
