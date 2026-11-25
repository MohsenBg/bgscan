package crud

// View renders the wrapped table.
func (m *Model[T]) View() string {
	return m.table.View()
}
