package dialog

import (
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

// OpenDialogMsg requests opening a dialog overlay containing the given component.
type OpenDialogMsg struct {
	Component ui.Component

	XPos DialogPosition
	YPos DialogPosition

	XOffset int
	YOffset int

	OnClose tea.Cmd
}

// OpenDialog builds an OpenDialogMsg. Position options override the default
// centered placement.
func OpenDialog(component ui.Component, opts ...DialogOption) OpenDialogMsg {
	msg := OpenDialogMsg{
		Component: component,
		XPos:      Center,
		YPos:      Center,
		XOffset:   0,
		YOffset:   0,
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

// WithOnClose sets the command to execute when the dialog closes.
func WithOnClose(cmd tea.Cmd) DialogOption {
	return func(m *OpenDialogMsg) {
		m.OnClose = cmd
	}
}
