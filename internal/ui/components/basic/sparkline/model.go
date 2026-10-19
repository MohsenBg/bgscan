package sparkline

import (
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
)

// Model is a tiny inline chart drawn with braille characters.
//
// Every braille cell is a 2x4 dot grid, so one cell shows two samples side by
// side with a natural gap between them. A chart `width` cells wide therefore
// shows the last width*2 samples.
type Model struct {
	id   ui.ComponentID
	env  *ui.Env
	name string

	samples []float64
	width   int // number of terminal cells

	// minSpan is the smallest value range the chart zooms into, as a fraction
	// of the highest sample. It keeps tiny wobbles from looking dramatic.
	minSpan float64

	// smoothing is the weight of the newest sample in an exponential moving
	// average. 0 disables smoothing; lower values are smoother.
	smoothing float64
}

// Option customizes a Model during creation.
type Option func(*Model)

// New creates a sparkline chart.
func New(env *ui.Env, opts ...Option) *Model {
	m := &Model{
		id:      ui.NewComponentID(),
		name:    "Sparkline",
		env:     env,
		width:   12,
		minSpan: 0.25,
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// WithWidth sets the chart width in terminal cells.
// Each cell shows two samples, so the chart keeps width*2 samples.
func WithWidth(width int) Option {
	return func(m *Model) {
		if width >= 1 {
			m.width = width
		}
	}
}

// WithMinSpan sets how far the chart may zoom in, as a fraction of the highest
// sample (0..1). Smaller values react to small changes, larger values are calmer.
// Values outside 0..1 are ignored.
func WithMinSpan(ratio float64) Option {
	return func(m *Model) {
		if ratio >= 0 && ratio <= 1 {
			m.minSpan = ratio
		}
	}
}

// WithSmoothing enables an exponential moving average on pushed samples.
// Alpha is the weight of the new sample, between 0 and 1 exclusive.
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

// capacity is the number of samples the chart keeps (two per cell).
func (m *Model) capacity() int { return m.width * 2 }
