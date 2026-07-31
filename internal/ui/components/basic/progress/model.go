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
	// padding defines the horizontal padding applied around the progress bar.
	padding = 1

	// maxWidth limits the maximum width of the progress bar to prevent
	// overly wide rendering on large terminals.
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

// Width calculates the progress bar width based on the layout body.
//
// The width is constrained by:
//   - the layout body width
//   - the defined maximum width
//   - internal horizontal padding
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

// Mode returns the interaction mode of the component.
//
// The progress component operates in NormalMode and does not
// capture exclusive input.
func (m *Model) Mode() env.Mode {
	return env.NormalMode
}
