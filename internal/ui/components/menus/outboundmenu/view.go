package outboundmenu

import "charm.land/lipgloss/v2"

// View renders the inner menu with horizontal padding.
func (m *Model) View() string {
	return lipgloss.NewStyle().Padding(0, 2).Render(m.menu.View())
}
