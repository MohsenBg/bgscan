package body

import (
	"bgscan/internal/ui/components/basic/confirm"
	"bgscan/internal/ui/main/footer"
	"bgscan/internal/ui/shared/env"
	"bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

// Update routes messages through the component stack.
//
// Key behavior:
//   - back in a nested stack pops the top component
//   - quit from the root opens the exit confirmation dialog
//   - OpenComponentMsg pushes a new component
//   - CloseComponentMsg removes the matching component
//   - ResetComponentStacksMsg pops back to the root menu
//   - all other messages are forwarded to the active top component
func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	lastIdx := len(m.components) - 1

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if env.IsBackKey(msg, m.components[lastIdx].Mode()) && len(m.components) > 1 {
			return m.popComponent()
		}

		if env.IsQuitKey(msg, m.components[lastIdx].Mode()) {
			return m, confirm.ExitConfirmCmd(m.state.Layout)
		}

	case ui.OpenComponentMsg:
		if msg.Component != nil {
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

// pushComponent adds a component to the stack and initializes it.
func (m *Model) pushComponent(c ui.Component) (ui.Component, tea.Cmd) {
	m.components = append(m.components, c)

	return m, tea.Batch(
		c.Init(),
		m.forceResize(),
		m.updateStatusCmd(c.Name()),
	)
}

// popComponent removes the top component and returns to the previous one.
func (m *Model) popComponent() (ui.Component, tea.Cmd) {
	lastIdx := len(m.components) - 1
	c := m.components[lastIdx]

	closeCmd := c.OnClose()
	m.components[lastIdx] = nil
	m.components = m.components[:lastIdx]

	newTop := m.components[len(m.components)-1]

	return m, tea.Batch(
		closeCmd,
		m.updateStatusCmd(newTop.Name()),
	)
}

// updateStatusCmd returns a command that updates the footer status text.
func (m *Model) updateStatusCmd(name string) tea.Cmd {
	return func() tea.Msg {
		return footer.UpdateStatus{Status: name}
	}
}

// forceResize returns a command that re-emits the current terminal size so
// newly pushed components receive a resize event.
func (m *Model) forceResize() tea.Cmd {
	return func() tea.Msg {
		return tea.WindowSizeMsg{
			Width:  m.state.Layout.Terminal.Width,
			Height: m.state.Layout.Terminal.Height,
		}
	}
}
