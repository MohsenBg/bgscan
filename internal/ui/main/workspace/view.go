package workspace

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// View renders the workspace layout.
func (m *model) View() string {
	env := m.session.Env()
	term := env.Layout.Terminal()
	content := env.Layout.Content()

	bodyView := m.body.View()
	footerView := m.footer.View()

	parts := []string{bodyView, footerView}

	joined := strings.Join(parts, "\n")
	main := m.contentStyle(content.Width, content.Height).Render(joined)

	// The framed content already fills the terminal exactly, so centering it
	// inside another full-screen box would be a no-op that costs a whole
	// render.
	if lipgloss.Width(main) == term.Width && lipgloss.Height(main) == term.Height {
		return main
	}

	return screenStyle(term.Width, term.Height).Render(main)
}
