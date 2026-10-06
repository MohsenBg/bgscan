package app

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	overlay "github.com/rmhubbert/bubbletea-overlay"
)

func (m *model) View() tea.View {
	termWidth := m.state.Layout.Terminal.Width
	termHeight := m.state.Layout.Terminal.Height

	if !m.state.Layout.HasSpace() {
		t := tea.NewView(m.renderLimitSize(termWidth, termHeight))
		t.AltScreen = true
		return t
	}

	var view string
	switch m.stage {
	case StageSplash:
		view = m.splash.View()
	case StageStartUP:
		view = m.startup.View()
	case StageWorkspace:
		view = m.workspace.View()
	}
	view = m.renderDialogs(view)
	t := tea.NewView(view)
	if m.state.Settings.SolidBackground {
		th := m.state.Theme()
		t.BackgroundColor = th.Background
		t.ForegroundColor = th.Text
	}

	t.AltScreen = true
	return t
}

func (m *model) renderLimitSize(termWidth, termHeight int) string {
	msg := fmt.Sprintf(
		"Terminal too small\nMinimum size is %dx%d\nPlease resize your terminal to have more space.",
		m.state.Layout.MinTerminal.Width,
		m.state.Layout.MinTerminal.Height,
	)

	return m.centerStyle().
		Width(termWidth).
		Height(termHeight).
		Render(
			m.warningStyle().Render(msg),
		)
}

// renderDialogs composites the open dialogs over the base view.
func (m *model) renderDialogs(base string) string {
	view := base

	for _, layer := range m.layers {
		opts := m.frameFor(layer.ID())
		hints := m.renderHints(append(opts.Hints, m.state.Deps().Keys.Dialog.Cancel.ShortItem()))

		view = overlay.Composite(
			m.frameView(layer.View(), opts, hints),
			view,
			opts.XPos, opts.YPos,
			opts.XOffset, opts.YOffset,
		)
	}

	return view
}
