package textarea

import (
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

// Update handles textarea input.
//
// Enter submits without inserting a newline. When newlines are enabled
// (the default), Shift+Enter (or Ctrl+M) inserts a literal newline. When
// newlines are disabled via [WithNewlines], no key inserts a newline.
func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.textarea.SetWidth(m.Width())

	case tea.KeyPressMsg:
		switch msg.String() {
		case "shift+enter", "ctrl+m":
			if m.allowNewline {
				m.textarea.InsertRune('\n')
				return m, nil
			}
			return m, m.submit()
		}

		if msg.Code == tea.KeyEnter {
			return m, m.submit()
		}
	}

	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)

	if _, ok := msg.(tea.KeyPressMsg); ok &&
		m.dynamicValidation &&
		m.validationFunc != nil {
		m.errorMsg = ""

		if err := m.validationFunc(m.textarea.Value()); err != nil {
			m.errorMsg = err.Error()
		}
	}

	return m, cmd
}
