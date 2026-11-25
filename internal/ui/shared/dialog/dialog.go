package dialog

import (
	"github.com/MohsenBg/bgscan/internal/ui/shared/keys"
	ui "github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

// DialogPosition specifies where a dialog is anchored on screen.
type DialogPosition int

// Dialog anchor positions. Center is valid on both axes, and the other
// values name the edge the dialog is anchored to.
const (
	Top DialogPosition = iota + 1
	Right
	Bottom
	Left
	Center
)

// DialogStyle controls the dialog frame style. Each style picks a
// border and title color from the theme.
type DialogStyle uint8

// Dialog frame styles.
const (
	StylePrimary DialogStyle = iota
	StyleInfo
	StyleWarn
	StyleDanger
	StyleSuccess
)

// DefaultAnimation is the reveal transition used when a dialog does not
// request a specific one.
const DefaultAnimation = AnimSplit

// OpenDialogMsg requests opening a dialog overlay that contains the
// given component.
type OpenDialogMsg struct {
	// Component is the dialog content.
	Component ui.Component

	// XPos and YPos anchor the dialog on screen.
	XPos DialogPosition
	YPos DialogPosition

	// XOffset and YOffset shift the dialog from its anchor, in cells.
	// Positive values move it right and down.
	XOffset int
	YOffset int

	// Title is drawn in the frame's top border.
	Title string

	// Style picks the frame colors.
	Style DialogStyle

	// Animation is the reveal transition for opening and closing.
	Animation Animation

	// Hints are drawn in the frame's bottom border.
	Hints []keys.Hint

	// OnClose runs when the dialog closes. It may be nil.
	OnClose tea.Cmd
}

// OpenDialog builds an OpenDialogMsg. Position options override the default
// centered placement.
func OpenDialog(component ui.Component, opts ...DialogOption) OpenDialogMsg {
	msg := OpenDialogMsg{
		Component: component,
		XPos:      Center,
		YPos:      Center,
		Style:     StylePrimary,
		Animation: DefaultAnimation,
	}

	for _, opt := range opts {
		opt(&msg)
	}

	return msg
}

// DialogOption mutates an OpenDialogMsg. OpenDialog applies the options in order.
type DialogOption func(*OpenDialogMsg)

// WithPosition sets the dialog anchor positions.
func WithPosition(x, y DialogPosition) DialogOption {
	return func(m *OpenDialogMsg) {
		m.XPos = x
		m.YPos = y
	}
}

// WithOffset adds offsets, in cells, to the dialog position.
func WithOffset(x, y int) DialogOption {
	return func(m *OpenDialogMsg) {
		m.XOffset = x
		m.YOffset = y
	}
}

// WithTitle sets the title embedded in the dialog frame's top rule.
func WithTitle(title string) DialogOption {
	return func(m *OpenDialogMsg) {
		m.Title = title
	}
}

// WithHints sets the key bindings shown in the frame's bottom border.
func WithHints(h ...keys.Hint) DialogOption {
	return func(m *OpenDialogMsg) {
		m.Hints = append(m.Hints, h...)
	}
}

// WithStyle sets the dialog style.
func WithStyle(style DialogStyle) DialogOption {
	return func(m *OpenDialogMsg) {
		m.Style = style
	}
}

// WithDanger marks the dialog as destructive (danger frame style).
func WithDanger() DialogOption {
	return WithStyle(StyleDanger)
}

// WithSuccess marks the dialog as a success (success frame style).
func WithSuccess() DialogOption {
	return WithStyle(StyleSuccess)
}

// WithOnClose sets the command to execute when the dialog closes.
func WithOnClose(cmd tea.Cmd) DialogOption {
	return func(m *OpenDialogMsg) {
		m.OnClose = cmd
	}
}

// WithAnimation sets the reveal transition for the dialog's open and close.
func WithAnimation(a Animation) DialogOption {
	return func(m *OpenDialogMsg) {
		m.Animation = a
	}
}
