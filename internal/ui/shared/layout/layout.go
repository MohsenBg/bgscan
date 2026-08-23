package layout

// Layout represents the terminal screen geometry and component regions.
//
// It is recalculated on terminal resize so render paths can use precomputed
// regions instead of recalculating layout repeatedly.
type Layout struct {
	Terminal TerminalSize
	Content  ContentSize

	MinTerminal TerminalSize

	Header ComponentSize
	Body   ComponentSize
	Footer ComponentSize
}

// TerminalSize represents the raw terminal dimensions.
type TerminalSize struct {
	Width  int
	Height int
}

// ContentSize is the drawable terminal content area, excluding margins.
type ContentSize struct {
	Width   int
	Height  int
	Padding int
}

// ComponentSize describes a UI component's position and size in terminal cells.
type ComponentSize struct {
	Width  int
	Height int
	X      int
	Y      int
}

// New returns a Layout with fallback terminal dimensions. It should be
// updated with the real terminal size via Update before use.
func New() *Layout {
	return &Layout{
		Terminal: TerminalSize{
			Width:  80,
			Height: 35,
		},
		MinTerminal: TerminalSize{
			Width:  75,
			Height: 35,
		},
	}
}

// Update recalculates regions from the current terminal dimensions.
// Call this on tea.WindowSizeMsg.
func (l *Layout) Update(termWidth, termHeight int) {
	l.Terminal.Width = termWidth
	l.Terminal.Height = termHeight

	l.Content = ContentSize{
		Width:   termWidth - 2,
		Height:  termHeight - 2,
		Padding: 1,
	}

	l.Header = ComponentSize{
		Width:  l.Content.Width,
		Height: 8,
		X:      0,
		Y:      0,
	}

	l.Body = ComponentSize{
		Width:  l.Content.Width,
		Height: l.Content.Height - (l.Header.Height + 2),
		X:      0,
		Y:      l.Header.Height,
	}

	l.Footer = ComponentSize{
		Width:  l.Content.Width,
		Height: 2,
		X:      0,
		Y:      l.Body.Y + l.Body.Height,
	}
}

func (l *Layout) BodyContentWidth() int {
	return l.Body.Width
}

func (l *Layout) BodyContentHeight() int {
	return l.Body.Height
}

func (l *Layout) HasSpace() bool {
	termWidth := l.Terminal.Width
	termHeight := l.Terminal.Height

	minWidth := l.MinTerminal.Width
	minHeight := l.MinTerminal.Height

	return termWidth >= minWidth && termHeight >= minHeight
}
