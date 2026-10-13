package workspace

import (
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
	"github.com/MohsenBg/bgscan/internal/ui/theme"

	tea "charm.land/bubbletea/v2"
)

// Update routes messages through the workspace.
func (m *model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.session.Env().Layout.Update(msg.Width, msg.Height)

	case ui.HeaderVisibilityMsg:
		cmds = append(cmds, m.setHeaderVisibility(msg))
	}

	if _, themed := msg.(theme.ThemeChanged); themed {
		cmds = append(cmds, m.forwardToLayout(msg)...)
		return m, tea.Batch(cmds...)
	}

	cmds = append(cmds, m.forwardToLayout(msg)...)
	return m, tea.Batch(cmds...)
}

// setHeaderVisibility updates the header visibility and refreshes the layout.
func (m *model) setHeaderVisibility(msg ui.HeaderVisibilityMsg) tea.Cmd {
	env := m.session.Env()
	if env.Layout.ShowHeader() == msg.Visible {
		return nil
	}

	env.Layout.SetHeaderVisible(msg.Visible)

	term := env.Layout.Terminal()
	w, h := term.Width, term.Height
	return func() tea.Msg {
		return tea.WindowSizeMsg{Width: w, Height: h}
	}
}

// forwardToLayout passes msg to the header, body, and footer.
func (m *model) forwardToLayout(msg tea.Msg) []tea.Cmd {
	var headerCmd, bodyCmd, footerCmd tea.Cmd
	m.header, headerCmd = m.header.Update(msg)
	m.body, bodyCmd = m.body.Update(msg)
	m.footer, footerCmd = m.footer.Update(msg)
	return []tea.Cmd{headerCmd, bodyCmd, footerCmd}
}
