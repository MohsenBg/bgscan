package dialog

import (
	"github.com/MohsenBg/bgscan/internal/ui/shared/keys"
	ui "github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
	bubbleTeaOverlay "github.com/rmhubbert/bubbletea-overlay"
)

// DialogPosition specifies where a dialog is anchored on screen.
type DialogPosition = bubbleTeaOverlay.Position

const (
	Top    = bubbleTeaOverlay.Top
	Right  = bubbleTeaOverlay.Right
	Bottom = bubbleTeaOverlay.Bottom
	Left   = bubbleTeaOverlay.Left
	Center = bubbleTeaOverlay.Center
)

// DialogStyle controls the dialog frame style.
type DialogStyle uint8

const (
	StylePrimary DialogStyle = iota
	StyleInfo
	StyleWarn
	StyleDanger
)

// OpenDialogMsg requests opening a dialog overlay containing the given component.
type OpenDialogMsg struct {
	Component ui.Component

	XPos DialogPosition
	YPos DialogPosition

	XOffset int
	YOffset int

	Title string
	Style DialogStyle

	Hints []keys.Hint

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
	}

	for _, opt := range opts {
		opt(&msg)
	}

	return msg
}

// DialogOption mutates an OpenDialogMsg.
type DialogOption func(*OpenDialogMsg)

// WithPosition sets the dialog anchor positions.
func WithPosition(x, y DialogPosition) DialogOption {
	return func(m *OpenDialogMsg) {
		m.XPos = x
		m.YPos = y
	}
}

// WithOffset adds pixel/cell offsets to the dialog position.
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

// WithOnClose sets the command to execute when the dialog closes.
func WithOnClose(cmd tea.Cmd) DialogOption {
	return func(m *OpenDialogMsg) {
		m.OnClose = cmd
	}
}
