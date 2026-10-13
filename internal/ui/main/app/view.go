package app

import (
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/MohsenBg/bgscan/internal/ui/shared/dialog"
	"github.com/MohsenBg/bgscan/internal/ui/shared/keys"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
	"github.com/MohsenBg/bgscan/internal/ui/theme"
)

func (m *model) View() tea.View {
	env := m.session.Env()
	th := env.Theme

	if !env.Layout.HasSpace() {
		term := env.Layout.Terminal()
		t := tea.NewView(m.renderLimitSize(env, &th, term.Width, term.Height))
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
	view = m.renderDialogs(env, &th, view)

	t := tea.NewView(view)
	if m.session.Settings.SolidBackground {
		t.BackgroundColor = th.Background
		t.ForegroundColor = th.Text
	}
	t.AltScreen = true
	return t
}

func (m *model) renderLimitSize(env *ui.Env, th *theme.Theme, termWidth, termHeight int) string {
	minTerm := env.Layout.MinTerminal()
	msg := fmt.Sprintf(
		"Terminal too small\nMinimum size is %dx%d\nPlease resize your terminal to have more space.",
		minTerm.Width,
		minTerm.Height,
	)

	return centerStyle().
		Width(termWidth).
		Height(termHeight).
		Render(
			warningStyle(th).Render(msg),
		)
}

func (m *model) renderDialogs(env *ui.Env, th *theme.Theme, base string) string {
	if len(m.layers) == 0 {
		return base
	}

	now := time.Now()
	cancel := env.Keys.Dialog.Cancel.ShortItem()
	width := env.Layout.Body().Width

	view := base
	for _, l := range m.layers {
		opts := l.frame()

		hints := make([]keys.Hint, 0, len(opts.Hints)+1)
		hints = append(hints, opts.Hints...)
		hints = append(hints, cancel)

		view = dialog.Composite(
			m.frameView(th, width, l.component.View(), opts, hints),
			view,
			opts.XPos, opts.YPos,
			opts.XOffset, opts.YOffset,
			dialog.AnimSplit,
			l.progress(now),
		)
	}
	return view
}

func (m *model) frameView(th *theme.Theme, width int, content string, opts *frameOptions, hints []keys.Hint) string {
	return dialog.RenderFrame(content, dialog.FrameOptions{
		Title: opts.Title,
		Style: opts.Style,
		Hints: hints,
		Width: width,
	}, th)
}
