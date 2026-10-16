package inspector

import "charm.land/lipgloss/v2"

func (m *Model) View() string {
	sections := make([]string, 0, 4)

	if len(m.groups) > 1 {
		sections = append(sections, m.tabs.View())
	}
	if m.Title != "" {
		sections = append(sections, titleStyle(&m.env.Theme).Render(m.Title))
	}

	centered := lipgloss.NewStyle().Width(m.Width()).Align(lipgloss.Center)
	sections = append(sections, centered.Render(m.list.View()))

	if m.help != nil {
		sections = append(sections, centered.Render(m.help.View()))
	}

	return lipgloss.JoinVertical(lipgloss.Center, sections...)
}
