package textarea

import (
	"bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

// Update forwards the message to the underlying textarea, then on Enter
// submits (validating) and optionally runs dynamic validation on other keys.
func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	var cmd tea.Cmd
	// Always update the underlying textarea first
	m.textarea, cmd = m.textarea.Update(msg)
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.textarea.SetWidth(m.Width())
		return m, cmd
	case tea.KeyPressMsg:
		value := m.textarea.Value()
		switch msg.Code {
		case tea.KeyEnter:
			return m, m.submit()
		default:
			if m.dynamicValidation && m.validationFunc != nil {
				m.errorMsg = ""
				if err := m.validationFunc(value); err != nil {
					m.errorMsg = err.Error()
				}
			}
		}
	}
	return m, cmd
}
