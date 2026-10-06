package app

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/MohsenBg/bgscan/internal/ui/shared/keys"
)

const hintSeparator = "   "

// renderHints renders key hints along the frame's bottom edge.
func (m *model) renderHints(bindings []keys.Hint) string {
	hints := make([]string, 0, len(bindings))

	for _, binding := range bindings {
		hints = append(hints, m.renderHint(binding))
	}

	return strings.Join(hints, hintSeparator)
}

// renderHint renders a key and its description.
func (m *model) renderHint(binding keys.Hint) string {
	th := m.state.Theme()

	return lipgloss.NewStyle().
		Foreground(th.Primary).
		Bold(true).
		Render(binding.Key) +
		" " +
		lipgloss.NewStyle().
			Foreground(th.Muted).
			Render(binding.Label)
}
