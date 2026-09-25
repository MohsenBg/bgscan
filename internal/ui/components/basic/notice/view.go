package notice

import (
	"charm.land/lipgloss/v2"
)

// View renders the notice component.
//
// Layout structure:
//
//	Header
//	Message viewport
//	Footer button
func (m *Model) View() string {
	width := m.Width()
	if width <= 0 {
		return ""
	}

	// Wrap message content to component width
	wrapped := lipgloss.NewStyle().
		Width(width).
		Render(m.message)

	m.viewport.SetContent(wrapped)

	content := lipgloss.JoinVertical(
		lipgloss.Top,
		m.headerView(width),
		m.viewport.View(),
		m.footerView(width),
	)

	return containerStyle(width).Render(content)
}

func (m *Model) headerView(width int) string {
	p := m.levelPalette(m.noticeType)

	return m.titleStyle(width, m.noticeType).Render(
		p.Icon + m.title,
	)
}

func (m *Model) footerView(width int) string {
	p := m.levelPalette(m.noticeType)

	button := m.buttonStyle().Render(p.FooterText)

	return CenterStyle(width).Render(button)
}
