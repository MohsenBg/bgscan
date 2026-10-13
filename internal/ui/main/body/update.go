package body

import (
	"slices"

	"github.com/MohsenBg/bgscan/internal/ui/components/basic/confirm"
	"github.com/MohsenBg/bgscan/internal/ui/main/footer"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
	"github.com/MohsenBg/bgscan/internal/ui/theme"

	tea "charm.land/bubbletea/v2"
)

const mainMenuName = "Main Menu"

func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	n := len(m.components)
	if n == 0 {
		return m, nil
	}
	top := m.components[n-1]
	env := m.session.Env()

	switch msg := msg.(type) {
	case theme.ThemeChanged:
		cmds := make([]tea.Cmd, 0, n)
		for i, c := range m.components {
			var cmd tea.Cmd
			m.components[i], cmd = c.Update(msg)
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)

	case tea.KeyPressMsg:
		k := env.Keys
		switch top.Mode() {
		case ui.KeyModeNormal:
			if n > 1 && k.Global.Back.Matches(msg) {
				return m.popComponent()
			}
			if k.Global.Quit.Matches(msg) {
				return m, confirm.ExitConfirmCmd(env)
			}
		case ui.KeyModeInput:
			if k.Global.BackFromInput.Matches(msg) {
				return m.popComponent()
			}
		}

	case ui.OpenComponentMsg:
		if msg.Component != nil {
			env.Log.Info("Opening screen: %s", msg.Component.Name())
			return m.pushComponent(msg.Component)
		}

	case ui.CloseComponentMsg:
		for i, c := range m.components {
			if c.ID() != msg.ID {
				continue
			}
			if i == 0 { // never close the root screen
				return m, nil
			}
			cmd := c.OnClose()
			m.components = slices.Delete(m.components, i, i+1) // zeroes the tail
			return m, tea.Batch(cmd, m.stackChangedCmd())
		}

	case ui.ResetComponentStacksMsg:
		cmds := make([]tea.Cmd, 0, n+1)
		for i := 1; i < n; i++ {
			cmds = append(cmds, m.components[i].OnClose())
			m.components[i] = nil
		}
		m.components = m.components[:1]
		cmds = append(cmds, m.stackChangedCmd())
		return m, tea.Batch(cmds...)
	}

	activeComp, cmd := top.Update(msg)
	m.components[n-1] = activeComp
	return m, cmd
}

func (m *Model) pushComponent(c ui.Component) (ui.Component, tea.Cmd) {
	m.components = append(m.components, c)
	return m, tea.Batch(c.Init(), m.forceResize(), m.stackChangedCmd())
}

func (m *Model) popComponent() (ui.Component, tea.Cmd) {
	n := len(m.components)
	if n <= 1 { // keep the root screen
		return m, nil
	}

	closeCmd := m.components[n-1].OnClose()
	m.components[n-1] = nil
	m.components = m.components[:n-1]

	return m, tea.Batch(closeCmd, m.stackChangedCmd())
}

func (m *Model) topName() string {
	if len(m.components) == 0 {
		return ""
	}
	return m.components[len(m.components)-1].Name()
}

// stackChangedCmd tells the header and footer about the new top screen.
// The header is shown only on the main menu. The name is read once, and
// both messages are built from that same value.
func (m *Model) stackChangedCmd() tea.Cmd {
	name := m.topName()
	visible := name == mainMenuName
	return tea.Batch(
		func() tea.Msg { return ui.HeaderVisibilityMsg{Visible: visible} },
		func() tea.Msg { return footer.UpdateStatus{Status: name} },
	)
}

// forceResize captures the size now, not when the command runs.
func (m *Model) forceResize() tea.Cmd {
	term := m.session.Env().Layout.Terminal()
	msg := tea.WindowSizeMsg{Width: term.Width, Height: term.Height}
	return func() tea.Msg { return msg }
}
