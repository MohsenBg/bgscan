package workspace

import (
	"charm.land/lipgloss/v2"

	"github.com/MohsenBg/bgscan/internal/ui/shared/perf"
)

// View renders the workspace layout.
func (m *model) View() string {
	env := m.session.Env()
	term := env.Layout.Terminal()
	content := env.Layout.Content()

	defer perf.Begin("ws")()
	stopBody := perf.Begin("ws.body")
	bodyView := m.body.View()
	stopBody()

	stopFooter := perf.Begin("ws.footer")
	footerView := m.footer.View()
	stopFooter()

	parts := []string{bodyView, footerView}

	stopStyle := perf.Begin("ws.style")
	joined := lipgloss.JoinVertical(lipgloss.Left, parts...)
	main := m.contentStyle(content.Width, content.Height).Render(joined)
	stopStyle()

	// The framed content already fills the terminal exactly, so centering it
	// inside another full-screen box would be a no-op that costs a whole
	// render.
	if lipgloss.Width(main) == term.Width && lipgloss.Height(main) == term.Height {
		return main
	}

	return screenStyle(term.Width, term.Height).Render(main)
}
