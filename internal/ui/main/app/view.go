package app

import (
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/MohsenBg/bgscan/internal/ui/shared/dialog"
	"github.com/MohsenBg/bgscan/internal/ui/shared/keys"
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

// renderDialogs composites open dialogs over the base view.
func (m *model) renderDialogs(base string) string {
	view := base
	now := time.Now()

	for _, l := range m.layers {
		opts := l.frame()
		hints := append(opts.Hints, m.state.Deps().Keys.Dialog.Cancel.ShortItem())

		view = dialog.Composite(
			m.frameView(l.component.View(), opts, hints),
			view,
			opts.XPos, opts.YPos,
			opts.XOffset, opts.YOffset,
			dialog.AnimSplit,
			l.progress(now),
		)
	}

	return view
}

func (m *model) frameView(content string, opts *frameOptions, hints []keys.Hint) string {
	return dialog.RenderFrame(content, dialog.FrameOptions{
		Title: opts.Title,
		Style: opts.Style,
		Hints: hints,
		Width: m.state.Layout.Body.Width,
	}, m.state.Theme())
}
