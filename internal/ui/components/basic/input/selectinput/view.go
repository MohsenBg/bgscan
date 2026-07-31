package selectinput

import (
	"bgscan/internal/ui/components/basic/input"

	"charm.land/lipgloss/v2"
)

// View renders the select component: title, options, validation error, and key hints.
func (m *Model[T]) View() string {
	content := make([]string, 0, 4)

	// Optional message
	if m.title != "" {
		content = append(content, input.MessageStyle().Render(m.title))
	}

	// Select field
	content = append(content, m.huhInput.View())

	// Validation error (if present)
	if m.errorMsg != "" {
		content = append(content, input.ErrorStyle().Render("✗ "+m.errorMsg))
	}

	// Key hints
	hints := input.KeyHintStyle().Render("↑/↓ to move • Enter to confirm • Esc/b to cancel")
	content = append(content, hints)

	body := lipgloss.JoinVertical(
		lipgloss.Top,
		content...,
	)
	return input.ContainerStyle(m.Width()).Render(body)
}
