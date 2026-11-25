package progress

import (
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
)

// Model is a reusable progress bar. It wraps the BubbleTea progress model,
// adapts its width to the layout, and tracks the current percentage.
type Model struct {
	id   ui.ComponentID
	name string

	env      *ui.Env
	progress progress.Model

	width int

	// percent represents the current progress value (0.0 → 1.0).
	percent float64
}

// Option configures a Model.
type Option func(*Model)

// WithProgressVisable sets whether the bar shows its percentage text.
func WithProgressVisable(show bool) Option {
	return func(m *Model) {
		m.progress.ShowPercentage = show
	}
}

// WithWidth sets the preferred bar width in cells.
func WithWidth(width int) Option {
	return func(m *Model) {
		m.width = width
	}
}

// SetWidth sets the bar width in cells, capped at the layout body width.
func (m *Model) SetWidth(width int) {
	m.width = width
	m.progress.SetWidth(min(width, m.env.Layout.Body().Width))
}

// New creates a progress bar that uses the theme gradient colors and
// starts at zero percent.
func New(env *ui.Env, opts ...Option) *Model {
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
		width:    min(90, env.Layout.Body().Width),
	}
	for _, opt := range opts {
		opt(m)
	}

	m.progress.SetWidth(m.width)
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

// Init returns no command, because the bar only changes on messages.
func (m *Model) Init() tea.Cmd {
	return nil
}

// ID returns the component ID.
func (m *Model) ID() ui.ComponentID {
	return m.id
}

// Name returns the component name.
func (m *Model) Name() string {
	return m.name
}

// OnClose returns no command, because the bar has nothing to release.
func (m *Model) OnClose() tea.Cmd {
	return nil
}

// Mode returns the normal key mode, because the bar takes no key input.
func (m *Model) Mode() ui.KeyMode {
	return ui.KeyModeNormal
}
