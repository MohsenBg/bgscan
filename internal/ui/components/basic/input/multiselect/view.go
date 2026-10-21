package multiselect

import (
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/input"

	"charm.land/lipgloss/v2"
)

// View renders the multi-select component.
func (m *Model[T]) View() string {
	content := make([]string, 0, 4)

	if m.title != "" {
		content = append(content, input.MessageStyle(&m.env.Theme).Render(m.title))
	}

	content = append(content, m.huhInput.View())

	if m.errorMsg != "" {
		content = append(content, input.ErrorStyle(&m.env.Theme).Render("✗ "+m.errorMsg))
	}

	body := lipgloss.JoinVertical(lipgloss.Top, content...)
	return input.ContainerStyle(m.Width()).Render(body)
}
