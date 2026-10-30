package app

import (
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/MohsenBg/bgscan/internal/ui/shared/dialog"
	"github.com/MohsenBg/bgscan/internal/ui/shared/keys"
	"github.com/MohsenBg/bgscan/internal/ui/shared/perf"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
	"github.com/MohsenBg/bgscan/internal/ui/theme"
)

func (m *model) View() tea.View {
	start := time.Now()
	defer func() { m.stats.recordFrame(time.Since(start)) }()

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
		view = m.workspaceView()
	}
	view = m.renderDialogs(env, &th, view)

	t := tea.NewView(view)
	if m.session.Env().Settings.SolidBackground {
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

// baseInterval caps how often the workspace behind an open dialog is rebuilt.
// The dialog covers part of it, so a copy this old is invisible, while
// rebuilding it on every animation frame and every key press is what made the
// event loop fall behind the keyboard.
const baseInterval = 150 * time.Millisecond

// workspaceView renders the workspace, reusing the last render for a short
// while when a dialog is open. Anything that changes the shape of the screen
// drops the copy through invalidateBase.
func (m *model) workspaceView() string {
	defer perf.Begin("wscache")()

	if len(m.layers) == 0 {
		m.invalidateBase()
		return m.workspace.View()
	}

	if m.baseOK && time.Since(m.baseAt) < baseInterval {
		perf.Record("wscache.hit", 0)
		return m.base
	}

	m.base = m.workspace.View()
	m.baseAt = time.Now()
	m.baseOK = true

	return m.base
}

// invalidateBase drops the cached workspace render.
func (m *model) invalidateBase() {
	m.base = ""
	m.baseOK = false
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

		// check if main animation disable
		animation := opts.Animation
		if !env.Settings.Animated {
			animation = dialog.AnimNone
		}

		stopComponent := perf.Begin("dlg.component")
		inner := l.component.View()
		stopComponent()

		stopFrame := perf.Begin("dlg.frame")
		boxed := m.frameView(th, width, inner, opts, hints)
		stopFrame()

		stopComposite := perf.Begin("dlg.composite")
		view = dialog.Composite(
			boxed,
			view,
			opts.XPos, opts.YPos,
			opts.XOffset, opts.YOffset,
			animation,
			l.progress(now),
		)
		stopComposite()
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
