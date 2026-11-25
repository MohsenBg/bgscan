package dns

// View renders the DNS settings inspector.
func (m *Model) View() string {
	return m.inspector.View()
}
