package dialog

import (
	ui "bgscan/internal/ui/shared/ui"

	bubbleTeaOverlay "github.com/rmhubbert/bubbletea-overlay"
)

// DialogPosition represents a positioning option for the dialog window.
type DialogPosition = bubbleTeaOverlay.Position

const (
	Top    = bubbleTeaOverlay.Top
	Right  = bubbleTeaOverlay.Right
	Bottom = bubbleTeaOverlay.Bottom
	Left   = bubbleTeaOverlay.Left
	Center = bubbleTeaOverlay.Center
)

// OpenDialogMsg requests the UI manager to open a component
// as a popup window on top of the current interface.
type OpenDialogMsg struct {
	Component ui.Component

	XPos DialogPosition
	YPos DialogPosition

	XOffset int
	YOffset int
}

// OpenDialog creates a message requesting a new dialog popup.
// By default, it centers the popup on the screen.
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

type DialogOption func(*OpenDialogMsg)

func WithPosition(x, y DialogPosition) DialogOption {
	return func(m *OpenDialogMsg) {
		m.XPos = x
		m.YPos = y
	}
}

func WithOffset(x, y int) DialogOption {
	return func(m *OpenDialogMsg) {
		m.XOffset = x
		m.YOffset = y
	}
}
