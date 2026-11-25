// Package layout computes the terminal screen geometry for the app. A
// Layout holds the raw terminal size and the precomputed regions for
// the content, body, and footer, so render code reads sizes instead of
// recalculating them. Callers update it on every terminal resize.
package layout

// frameBorderWidth is the number of cells the content frame border uses
// across the width, one cell on each side.
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
	Padding int // Padding is the margin inside the content area, in cells.
}

// ComponentSize describes a UI component's position and size in terminal cells.
type ComponentSize struct {
	Width  int
	Height int
	X      int // X is the column of the top left corner, in cells.
	Y      int // Y is the row of the top left corner, in cells.
}

// New returns a Layout with fallback terminal dimensions. Callers update
// it with the real terminal size through Update before use.
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

// Update recalculates the regions from the given terminal dimensions.
// Callers call it when a tea.WindowSizeMsg arrives.
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
