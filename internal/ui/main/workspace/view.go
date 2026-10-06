package workspace

import (
	"charm.land/lipgloss/v2"
)

// View renders the workspace layout.
func (m *model) View() string {
	layout := m.state.Layout

	parts := make([]string, 0, 3)
	if layout.ShowHeader {
		parts = append(parts, m.header.View())
	}
	parts = append(parts, m.body.View(), m.footer.View())

	content := lipgloss.JoinVertical(lipgloss.Left, parts...)

	main := m.contentStyle(layout.Content.Width, layout.Content.Height).Render(content)
	return screenStyle(layout.Terminal.Width, layout.Terminal.Height).Render(main)
}
