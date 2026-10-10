package confirm

import (
	"github.com/MohsenBg/bgscan/internal/ui/shared/dialog"
	"github.com/MohsenBg/bgscan/internal/ui/shared/keys"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

func ConfirmHints(k keys.Keys) keys.Hints {
	return keys.Hints{
		keys.Combine("move", k.Nav.Left, k.Nav.Right),
		k.Confirm.Accept.ShortItem(),
	}
}

// ExitConfirmCmd returns a command that opens the exit confirmation dialog.
// Confirming the dialog quits the application.
func ExitConfirmCmd(deps ui.Deps, options ...dialog.DialogOption) tea.Cmd {
	return func() tea.Msg {
		opts := []dialog.DialogOption{
			dialog.WithPosition(dialog.Center, dialog.Top),
			dialog.WithOffset(0, 1),
			dialog.WithTitle("Exit"),
			dialog.WithHints(ConfirmHints(deps.Keys)...),
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

// ConfirmCmd opens a confirmation dialog and runs confirm when accepted.
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
			dialog.WithHints(ConfirmHints(deps.Keys)...),
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

// CloseCmd closes the confirmation dialog.
func (m *Model) CloseCmd() tea.Cmd {
	return func() tea.Msg {
		return ui.CloseComponentMsg{ID: m.ID()}
	}
}
