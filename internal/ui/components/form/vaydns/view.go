package vaydns

func (m *Model) View() string {
	if m.form == nil {
		return ""
	}
	return m.form.View()
}
