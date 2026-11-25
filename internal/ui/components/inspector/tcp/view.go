package tcp

// View returns the rendered view of the wrapped inspector.
func (m *Model) View() string {
	return m.inspector.View()
}
