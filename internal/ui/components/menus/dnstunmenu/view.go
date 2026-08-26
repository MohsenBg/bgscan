package dnstunmenu

import "charm.land/lipgloss/v2"

func (m *Model) View() string {
	return lipgloss.NewStyle().Padding(0, 2).Render(m.menu.View())
}
