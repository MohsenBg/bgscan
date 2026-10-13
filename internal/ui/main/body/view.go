package body

func (m *Model) View() string {
	env := m.session.Env()
	term := env.Layout.Body()
	return containerStyle(
		term.Width,
		term.Height,
	).Render(m.components[len(m.components)-1].View())
}
