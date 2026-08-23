package workspace

import (
	"charm.land/lipgloss/v2"
	overlay "github.com/rmhubbert/bubbletea-overlay"
)

func (m model) View() string {
	termWidth := m.state.Layout.Terminal.Width
	termHeight := m.state.Layout.Terminal.Height

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		m.header.View(),
		m.body.View(),
		m.footer.View(),
	)

	content = m.renderOverlays(content)
	view := containerStyle(termWidth, termHeight).Render(
		mainStyle(m.state.Layout.Content.Width, m.state.Layout.Content.Height).Render(content),
	)

	return view
}

func (m *model) renderOverlays(baseView string) string {
	view := baseView

	for _, layer := range m.dialog {
		placement := m.getDialogPlacement(layer.ID())

		view = overlay.Composite(
			WindowStyle(m.state.Layout.Body.Width).Render(layer.View()),
			view,
			placement.XPos,
			placement.YPos,
			placement.XOffset,
			placement.YOffset,
		)
	}

	return view
}
