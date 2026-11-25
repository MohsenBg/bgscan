package targetsource

// View renders the underlying menu.
func (m *Model) View() string {
	return m.menu.View()
}
