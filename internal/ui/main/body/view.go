package body

import "charm.land/lipgloss/v2"

func (m *Model) View() string {
	env := m.session.Env()
	term := env.Layout.Body()

	content := ""
	if m.swap != nil {
		content = m.swapView()
	}
	if content == "" {
		content = m.screens[len(m.screens)-1].component.View()
	}

	// A screen that already fills the body needs no centering box.
	if lipgloss.Width(content) == term.Width && lipgloss.Height(content) == term.Height {
		return content
	}

	return containerStyle(
		term.Width,
		term.Height,
	).Render(content)
}
