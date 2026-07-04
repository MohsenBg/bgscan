package confirm

import (
	"bgscan/internal/ui/shared/dialog"
	"bgscan/internal/ui/shared/layout"
	"bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

// ExitConfirmCmd opens a confirmation dialog asking whether the user wants to
// exit the application.
//
// If the user confirms, the program will terminate via tea.Quit.
//
// The dialog is shown as an overlay (top-center by default).
func ExitConfirmCmd(l *layout.Layout, options ...dialog.DialogOption) tea.Cmd {
	return func() tea.Msg {
		opts := []dialog.DialogOption{
			dialog.WithPosition(dialog.Center, dialog.Top),
			dialog.WithOffset(0, 1),
		}
		opts = append(opts, options...)

		return dialog.OpenDialog(
			New(
				l,
				"Are you sure you want to exit?",
				func() tea.Cmd { return tea.Quit },
				false,
			),
			opts...,
		)
	}
}

// ConfirmCmd opens a generic confirmation dialog overlay.
//
// Parameters:
//   - l           : layout manager used for sizing/positioning
//   - message     : message shown in the dialog
//   - confirm     : command executed when user confirms
//   - defaultYes  : initial selection state (true = Yes, false = No)
//   - options     : optional dialog configuration
//
// The dialog is displayed as an overlay managed by the UI system.
func ConfirmCmd(
	l *layout.Layout,
	message string,
	confirm tea.Cmd,
	defaultYes bool,
	options ...dialog.DialogOption,
) tea.Cmd {
	return func() tea.Msg {
		opts := []dialog.DialogOption{
			dialog.WithPosition(dialog.Center, dialog.Top),
			dialog.WithOffset(0, 1),
		}
		opts = append(opts, options...)

		return dialog.OpenDialog(
			New(
				l,
				message,
				func() tea.Cmd { return confirm },
				defaultYes,
			),
			opts...,
		)
	}
}

// CloseCmd closes the confirmation dialog by emitting a UI close message.
// The overlay manager removes the component from the stack.
func (m *Model) CloseCmd() tea.Cmd {
	return func() tea.Msg {
		return ui.CloseComponentMsg{ID: m.ID()}
	}
}
