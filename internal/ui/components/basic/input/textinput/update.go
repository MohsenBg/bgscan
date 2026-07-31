package textinput

import (
	"bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

// Update handles Bubble Tea messages. It updates the underlying text input
// first, then on Enter submits the value (validating it), and optionally runs
// dynamic validation on other keys.
func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	var cmd tea.Cmd

	// Always update the underlying text input first
	m.textinput, cmd = m.textinput.Update(msg)

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.textinput.SetWidth(m.Width())
		return m, cmd

	case tea.KeyPressMsg:
		value := m.textinput.Value()

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
