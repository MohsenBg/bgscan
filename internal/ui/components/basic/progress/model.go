package progress

import (
	"bgscan/internal/ui/shared/env"
	"bgscan/internal/ui/shared/layout"
	"bgscan/internal/ui/shared/ui"
	"bgscan/internal/ui/theme"

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
	progress progress.Model

	// percent represents the current progress value (0.0 → 1.0).
	percent float64
}

func New(layout *layout.Layout) *Model {
	p := progress.New(
		progress.WithScaled(true),
		progress.WithColors(
			theme.Current().ProgressStart,
			theme.Current().ProgressEnd,
		),
	)

	m := &Model{
		id:       ui.NewComponentID(),
		name:     "Progress",
		progress: p,
		layout:   layout,
		percent:  0,
	}

	m.progress.SetWidth(m.Width())
	m.progress.PercentFormat = " %0.2f%%"

	return m
}

// Init initializes the component.
//
// The progress component does not require any startup commands.
func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Width() int {
	width := min(m.layout.Body.Width, maxWidth)
	return width - padding*10
}

// ID returns the unique component identifier.
func (m *Model) ID() ui.ComponentID {
	return m.id
}

// Name returns the human‑readable component name.
func (m *Model) Name() string {
	return m.name
}

// OnClose is called when the component is removed from the UI stack.
func (m *Model) OnClose() tea.Cmd {
	return nil
}

func (m *Model) Mode() env.Mode {
	return env.NormalMode
}
