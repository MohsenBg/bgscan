package body

func (m *Model) View() string {
	return containerStyle(
		m.state.Layout.Body.Width,
		m.state.Layout.Body.Height,
	).Render(m.components[len(m.components)-1].View())
}
