package body

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

	return containerStyle(
		term.Width,
		term.Height,
	).Render(content)
}
