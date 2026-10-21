package workspace

import (
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
	"github.com/MohsenBg/bgscan/internal/ui/theme"

	tea "charm.land/bubbletea/v2"
)

// Update routes messages through the workspace.
func (m *model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	var cmds []tea.Cmd

	if _, themed := msg.(theme.ThemeChanged); themed {
		cmds = append(cmds, m.forwardToLayout(msg)...)
		return m, tea.Batch(cmds...)
	}

	cmds = append(cmds, m.forwardToLayout(msg)...)
	return m, tea.Batch(cmds...)
}

// forwardToLayout passes msg to the body and footer.
func (m *model) forwardToLayout(msg tea.Msg) []tea.Cmd {
	var bodyCmd, footerCmd tea.Cmd
	m.body, bodyCmd = m.body.Update(msg)
	m.footer, footerCmd = m.footer.Update(msg)
	return []tea.Cmd{bodyCmd, footerCmd}
}
