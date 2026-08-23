package app

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
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
	t := tea.NewView(view)
	t.AltScreen = true
	return t
}

func (m *model) renderLimitSize(termWidth, termHeight int) string {
	msg := fmt.Sprintf(
		"Terminal too small\nMinimum size is %dx%d\nPlease resize your terminal to have more space.",
		m.state.Layout.MinTerminal.Width,
		m.state.Layout.MinTerminal.Height,
	)

	return centerStyle().
		Width(termWidth).
		Height(termHeight).
		Render(
			warningStyle().Render(msg),
		)
}
