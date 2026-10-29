package logview

import (
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
	"github.com/MohsenBg/bgscan/internal/ui/theme"

	tea "charm.land/bubbletea/v2"
)

// Update resizes the viewport, appends incoming log batches, and forwards
// other messages to the viewport for scrolling.
func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg, theme.ThemeChanged:
		m.setSize()

	case LogAppendMsg:
		m.appendLines(msg.Lines)
		m.pushRows()
	}

	if k, ok := msg.(tea.KeyMsg); ok {
		switch {
		case m.env.Keys.LogView.Top.Matches(k):
			m.viewport.GotoTop()
		case m.env.Keys.LogView.Bottom.Matches(k):
			m.viewport.GotoBottom()
		}
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)

	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	// The help bar handles "?" (full help) and resize/theme changes.
	_, cmd = m.help.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}
