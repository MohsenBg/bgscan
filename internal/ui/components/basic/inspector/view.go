package inspector

import "charm.land/lipgloss/v2"

// View renders the tabs followed by the field list, stacked vertically
// inside a styled container sized to the inspector's width.
func (m *Model) View() string {
	sections := make([]string, 0, 2)
	if len(m.groups) > 1 {
		sections = append(sections, m.tabs.View())
	}
	if m.Title != "" {
		sections = append(sections, titleStyle().Render(m.Title))
	}

	sections = append(
		sections,
		lipgloss.NewStyle().Width(m.Width()).Align(lipgloss.Center).Render(m.list.View()),
	)

	return lipgloss.JoinVertical(lipgloss.Center, sections...)
}
