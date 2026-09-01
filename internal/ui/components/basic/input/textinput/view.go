package textinput

import (
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/input"

	"charm.land/lipgloss/v2"
)

// View renders the input dialog (title, field, validation error, key hints).
func (m *Model) View() string {
	content := make([]string, 0, 4)

	// Optional message
	if m.title != "" {
		content = append(content, input.MessageStyle().Render(m.title))
	}

	// Input field
	content = append(content, m.textinput.View())

	// Validation error (if present)
	if m.errorMsg != "" {
		content = append(content, input.ErrorStyle().Render("✗ "+m.errorMsg))
	}

	// Key hints
	hints := input.KeyHintStyle().Render("Enter to confirm • Esc to cancel")
	content = append(content, hints)

	body := lipgloss.JoinVertical(
		lipgloss.Top,
		content...,
	)

	return input.ContainerStyle(m.Width()).Render(body)
}
