package outbounds

// View renders the underlying CRUD table.
func (m *Model) View() string {
	return m.crudTable.View()
}
