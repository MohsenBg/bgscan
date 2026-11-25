package ipviewer

// View renders the underlying table.
func (m *Model) View() string {
	return m.table.View()
}
