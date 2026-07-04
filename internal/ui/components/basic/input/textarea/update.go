package textarea

import (
	"bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

// Update processes incoming Bubble Tea messages and updates the state
// of the textarea component.
//
// The method delegates message handling to the underlying textarea
// component first, then processes higher-level input dialog behavior
// such as submission and validation.
//
// Behavior:
//   - tea.WindowSizeMsg: adjusts the input width based on the layout.
//   - tea.KeyEnter: attempts to submit the input value.
//   - Other keys: may trigger dynamic validation if enabled.
//
// It returns the updated component and an optional command to execute.
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
