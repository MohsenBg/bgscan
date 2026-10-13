package progress

import (
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

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

	env      *ui.Env
	progress progress.Model

	// percent represents the current progress value (0.0 → 1.0).
	percent float64
}

func New(env *ui.Env) *Model {
	p := progress.New(
		progress.WithScaled(true),
		progress.WithColors(
			env.Theme.ProgressStart,
			env.Theme.ProgressEnd,
		),
	)

	m := &Model{
		id:       ui.NewComponentID(),
		name:     "Progress",
		progress: p,
		env:      env,
		percent:  0,
	}

	m.progress.SetWidth(m.Width())
	m.progress.PercentFormat = " %0.2f%%"

	return m
}

// rebuild recreates the progress gradient from the live theme.
func (m *Model) rebuild() {
	width := m.progress.Width()
	p := progress.New(
		progress.WithScaled(true),
		progress.WithColors(m.env.Theme.ProgressStart, m.env.Theme.ProgressEnd),
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
	width := min(m.env.Layout.Body().Width, maxWidth)
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

func (m *Model) Mode() ui.KeyMode {
	return ui.KeyModeNormal
}
