package logview

import (
	"bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// Update resizes the viewport, refreshes content on LogUpdateTickMsg, and
// forwards other messages to the viewport for scrolling.
func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg.(type) {

	case tea.WindowSizeMsg:
		m.setSize()
		m.viewport.SetContent(m.renderContent())

	case LogUpdateTickMsg:
		m.mu.Lock()

		if m.needUpdate {
			isEmpty := m.viewport.TotalLineCount() == 0
			shouldGoToBottom := isEmpty || m.viewport.AtBottom()
			m.viewport.SetContent(m.renderContent())
			if shouldGoToBottom {
				m.viewport.GotoBottom()
			}
			m.needUpdate = false
		}

		m.mu.Unlock()

		cmds = append(cmds, m.tick())
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)

	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// renderContent builds the viewport content string from buffered messages.
func (m *Model) renderContent() string {
	if len(m.messages) == 0 {
		return ""
	}

	width := max(0, m.viewport.Width()-5)

	content := formatLogLines(m.messages)

	return lipgloss.NewStyle().
		Width(width).
		Render(content)
}
