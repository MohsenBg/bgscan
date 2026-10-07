package form

import (
	"charm.land/lipgloss/v2"
)

func (m *Model) renderTitle() string {
	return lipgloss.NewStyle().
		Foreground(m.theme.Primary).
		Bold(true).Padding(1, 0).
		Align(lipgloss.Center).
		Width(m.width).
		Render(m.name)
}

func (m *Model) renderInspector() string {
	if m.inspector == nil {
		return ""
	}

	return m.inspector.View()
}

func (m *Model) renderKeyHints() string {
	save := m.keyHintStyle().Render("•s save")
	cancel := m.keyHintStyle().Render("•esc/b cancel")

	return lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Center).
		Render(
			lipgloss.JoinHorizontal(lipgloss.Top, save, "  ", cancel),
		)
}

func (m *Model) View() string {
	inspector := m.renderInspector()

	if m.inDialog {
		return lipgloss.NewStyle().Padding(0, 2).Render(inspector)
	}

	title := m.renderTitle()
	hints := m.renderKeyHints()

	return lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		lipgloss.NewStyle().Padding(0, 2).Render(inspector),
		hints,
	)
}
