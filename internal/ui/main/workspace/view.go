package workspace

import (
	"charm.land/lipgloss/v2"
)

// View renders the workspace layout.
func (m *model) View() string {
	env := m.session.Env()
	term := env.Layout.Terminal()
	content := env.Layout.Content()

	parts := make([]string, 0, 3)
	if env.Layout.ShowHeader() {
		parts = append(parts, m.header.View())
	}
	parts = append(parts, m.body.View(), m.footer.View())

	joined := lipgloss.JoinVertical(lipgloss.Left, parts...)

	main := m.contentStyle(content.Width, content.Height).Render(joined)
	return screenStyle(term.Width, term.Height).Render(main)
}
