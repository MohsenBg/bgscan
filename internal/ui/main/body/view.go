package body

// View renders the active top component inside the body container.
func (m *Model) View() string {
	view := containerStyle(
		m.state.Layout.Body.Width,
		m.state.Layout.Body.Height,
	).Render(m.components[len(m.components)-1].View())

	return view
}
