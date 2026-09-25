package progress

import (
	"github.com/MohsenBg/bgscan/internal/logger"
	"github.com/MohsenBg/bgscan/internal/ui/shared/env"
	"github.com/MohsenBg/bgscan/internal/ui/shared/layout"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
	"github.com/MohsenBg/bgscan/internal/ui/theme"

	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
)

const (
	padding  = 1
	maxWidth = 90
)

// Model is a reusable progress bar. It wraps the BubbleTea progress model,
// adapts its width to the layout, and tracks the current percentage.
type Model struct {
	id   ui.ComponentID
	name string

	layout   *layout.Layout
	theme    *theme.Theme
	log      logger.Set
	progress progress.Model

	// percent represents the current progress value (0.0 → 1.0).
	percent float64
}

func New(deps ui.Deps) *Model {
	p := progress.New(
		progress.WithScaled(true),
		progress.WithColors(
			deps.Theme.ProgressStart,
			deps.Theme.ProgressEnd,
		),
	)

	m := &Model{
		id:       ui.NewComponentID(),
		name:     "Progress",
		progress: p,
		layout:   deps.Layout,
		theme:    deps.Theme,
		log:      deps.Log,
		percent:  0,
	}

	m.progress.SetWidth(m.Width())
	m.progress.PercentFormat = " %0.2f%%"

	return m
}

func (m *Model) Theme() *theme.Theme { return m.theme }

// SetTheme rebuilds the progress gradient.
func (m *Model) SetTheme(th *theme.Theme) {
	m.theme = th
	width := m.progress.Width()
	p := progress.New(
		progress.WithScaled(true),
		progress.WithColors(th.ProgressStart, th.ProgressEnd),
	)
	p.SetWidth(width)
	p.PercentFormat = m.progress.PercentFormat
	_ = p.SetPercent(m.percent)
	p.Full = m.progress.Full
	p.Empty = m.progress.Empty
	p.EmptyColor = m.progress.EmptyColor
	m.progress = p
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Width() int {
	width := min(m.layout.Body.Width, maxWidth)
	return width - padding*10
}

func (m *Model) ID() ui.ComponentID {
	return m.id
}

func (m *Model) Name() string {
	return m.name
}

func (m *Model) OnClose() tea.Cmd {
	return nil
}

func (m *Model) Mode() env.Mode {
	return env.NormalMode
}
