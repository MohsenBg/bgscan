package textarea

import (
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

// Update handles textarea input.
func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.textarea.SetWidth(m.Width())

	case tea.KeyPressMsg:
		if m.env.Keys.TextArea.NewLine.Matches(msg) {
			if m.allowNewline {
				m.textarea.InsertRune('\n')
				return m, nil
			}
			return m, m.submit()
		}

		if m.env.Keys.TextArea.Submit.Matches(msg) {
			return m, m.submit()
		}
	}

	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)

	if _, ok := msg.(tea.KeyMsg); ok &&
		m.dynamicValidation &&
		m.validationFunc != nil {
		m.errorMsg = ""

		if err := m.validationFunc(m.textarea.Value()); err != nil {
			m.errorMsg = err.Error()
		}
	}

	return m, cmd
}
