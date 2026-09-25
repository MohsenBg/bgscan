package confirm

import (
	"github.com/MohsenBg/bgscan/internal/ui/shared/dialog"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

// ExitConfirmCmd returns a command that opens an exit confirmation dialog.
// If the user confirms, the program exits via tea.Quit. The dialog is shown as
// a top-center overlay by default.
func ExitConfirmCmd(deps ui.Deps, options ...dialog.DialogOption) tea.Cmd {
	return func() tea.Msg {
		opts := []dialog.DialogOption{
			dialog.WithPosition(dialog.Center, dialog.Top),
			dialog.WithOffset(0, 1),
		}
		opts = append(opts, options...)

		return dialog.OpenDialog(
			New(
				deps,
				"Are you sure you want to exit?",
				tea.Quit,
				false,
			),
			opts...,
		)
	}
}

// ConfirmCmd returns a command that opens a generic confirmation dialog.
// message is displayed in the dialog and confirm is executed when the user
// confirms. defaultYes sets the initial selection. Additional options configure
// the dialog. The dialog is shown as a top-center overlay.
func ConfirmCmd(
	deps ui.Deps,
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
				deps,
				message,
				confirm,
				defaultYes,
			),
			opts...,
		)
	}
}

// CloseCmd returns a command that closes the confirmation dialog.
// It emits ui.CloseComponentMsg, which causes the overlay manager to remove the
// component from the stack.
func (m *Model) CloseCmd() tea.Cmd {
	return func() tea.Msg {
		return ui.CloseComponentMsg{ID: m.ID()}
	}
}
