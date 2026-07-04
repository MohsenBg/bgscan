package toggleinput

import (
	"bgscan/internal/ui/components/basic/input"

	"charm.land/lipgloss/v2"
)

// View renders the toggle component.
func (m *Model) View() string {
	content := make([]string, 0, 4)

	if m.title != "" {
		content = append(content, input.MessageStyle().Render(m.title))
	}

	content = append(content, m.huhInput.View())

	if m.errorMsg != "" {
		content = append(content, input.ErrorStyle().Render("✗ "+m.errorMsg))
	}

	hints := input.KeyHintStyle().Render("←/→ to toggle • Enter to confirm • Esc to cancel")
	content = append(content, hints)

	body := lipgloss.JoinVertical(lipgloss.Top, content...)
	return input.ContainerStyle(m.Width()).Render(body)
}
