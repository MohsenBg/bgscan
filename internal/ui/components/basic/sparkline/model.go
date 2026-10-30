package sparkline

import (
	"charm.land/lipgloss/v2"

	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
)

// Model is an inline chart rendered with braille characters.
// Each cell displays two samples, so the chart holds width*2 samples.
type Model struct {
	id   ui.ComponentID
	env  *ui.Env
	name string

	samples []float64
	width   int

	// minSpan limits zoom as a fraction of the highest sample.
	minSpan float64

	// smoothing controls the EMA weight of new samples. Zero disables it.
	smoothing float64

	// Render cache — invalidated whenever samples/width change.
	dirty bool
	cache string

	// Gradient styles, rebuilt only when width changes.
	styles []lipgloss.Style
}

// Option configures a Model.
type Option func(*Model)

// New creates a sparkline chart.
func New(env *ui.Env, opts ...Option) *Model {
	m := &Model{
		id:      ui.NewComponentID(),
		name:    "Sparkline",
		env:     env,
		width:   12,
		minSpan: 0.25,
		dirty:   true,
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// WithWidth sets the chart width in terminal cells.
func WithWidth(width int) Option {
	return func(m *Model) {
		if width >= 1 {
			m.width = width
		}
	}
}

// WithMinSpan sets the minimum chart range relative to the highest sample.
func WithMinSpan(ratio float64) Option {
	return func(m *Model) {
		if ratio >= 0 && ratio <= 1 {
			m.minSpan = ratio
		}
	}
}

// WithSmoothing sets the EMA weight of new samples (0 < alpha < 1).
func WithSmoothing(alpha float64) Option {
	return func(m *Model) {
		if alpha > 0 && alpha < 1 {
			m.smoothing = alpha
		} else {
			m.smoothing = 0
		}
	}
}

func (m *Model) ID() ui.ComponentID { return m.id }
func (m *Model) Name() string       { return m.name }

func (m *Model) capacity() int { return m.width * 2 }
