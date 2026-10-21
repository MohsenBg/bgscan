package layout

const frameBorderWidth = 2

// Layout represents the terminal screen geometry and component regions.
//
// It is recalculated on terminal resize so render paths can use precomputed
// regions instead of recalculating layout repeatedly.
type Layout struct {
	terminal    TerminalSize
	content     ContentSize
	minTerminal TerminalSize
	body        ComponentSize
	footer      ComponentSize
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
	l := &Layout{
		terminal: TerminalSize{
			Width:  80,
			Height: 35,
		},
		minTerminal: TerminalSize{
			Width:  75,
			Height: 35,
		},
	}

	l.recalculate()
	return l
}

// recalculate rebuilds the component regions based on the current terminal
// dimensions.
func (l *Layout) recalculate() {
	l.content = ContentSize{
		Width:   l.terminal.Width - 2,
		Height:  l.terminal.Height - 2,
		Padding: 1,
	}

	innerWidth := l.content.Width - frameBorderWidth

	l.body = ComponentSize{
		Width:  innerWidth,
		Height: l.content.Height - 2,
		X:      0,
		Y:      0,
	}

	l.footer = ComponentSize{
		Width:  innerWidth,
		Height: 2,
		X:      0,
		Y:      l.body.Y + l.body.Height,
	}
}

// Update recalculates regions from the current terminal dimensions.
// Call this on tea.WindowSizeMsg.
func (l *Layout) Update(termWidth, termHeight int) {
	l.terminal.Width = termWidth
	l.terminal.Height = termHeight
	l.recalculate()
}

// SetMinTerminal sets the minimum terminal dimensions required for the layout.
func (l *Layout) SetMinTerminal(width, height int) {
	l.minTerminal.Width = width
	l.minTerminal.Height = height
}

// Terminal returns the current raw terminal dimensions.
func (l *Layout) Terminal() TerminalSize {
	return l.terminal
}

// Content returns the drawable terminal content area.
func (l *Layout) Content() ContentSize {
	return l.content
}

// MinTerminal returns the minimum allowed terminal dimensions.
func (l *Layout) MinTerminal() TerminalSize {
	return l.minTerminal
}

// Body returns the body component's size and position.
func (l *Layout) Body() ComponentSize {
	return l.body
}

// Footer returns the footer component's size and position.
func (l *Layout) Footer() ComponentSize {
	return l.footer
}

// HasSpace reports whether the current terminal dimensions meet the minimum
// requirements.
func (l *Layout) HasSpace() bool {
	return l.terminal.Width >= l.minTerminal.Width && l.terminal.Height >= l.minTerminal.Height
}
