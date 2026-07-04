package textinput

import (
	"bgscan/internal/ui/components/basic/input"

	"charm.land/lipgloss/v2"
)

// View renders the input dialog UI.
//
// The view consists of:
//   - An optional message displayed above the input field
//   - The text input component
//   - An optional validation error message
//   - Key hints for user interaction
//
// The content is vertically stacked and wrapped inside a styled
// container whose width is determined by the layout.
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
