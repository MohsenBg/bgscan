package logview

import (
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
	"github.com/MohsenBg/bgscan/internal/ui/theme"

	tea "charm.land/bubbletea/v2"
)

// Update resizes the viewport, refreshes content on LogUpdateTickMsg, and
// forwards other messages to the viewport for scrolling.
func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg.(type) {
	case tea.WindowSizeMsg, theme.ThemeChanged:
		m.setSize()

	case LogUpdateTickMsg:
		cmds = append(cmds, m.refreshView())
		cmds = append(cmds, m.tick())
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

	_, cmd = m.help.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// refreshView moves newly rendered rows into the viewport.
func (m *Model) refreshView() tea.Cmd {
	m.mu.Lock()
	if !m.needUpdate {
		m.mu.Unlock()
		return nil
	}
	m.needUpdate = false
	lines, drop := m.refreshLocked()
	m.mu.Unlock()

	atBottom := m.viewport.AtBottom()

	m.viewport.SetContentLines(lines)

	switch {
	case atBottom:
		m.viewport.GotoBottom()
	case drop > 0:
		m.viewport.SetYOffset(m.viewport.YOffset() - drop)
	}

	return nil
}
