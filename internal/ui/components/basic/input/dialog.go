package input

import (
	"bgscan/internal/ui/shared/dialog"
	"bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

// OpenInputDialog opens input as an overlay, closing it automatically
// after the input's own onSubmit handler succeeds.
func OpenInputDialog(input Dialog, optOverlay ...dialog.DialogOption) tea.Cmd {
	input.AppendOnSubmit(func() tea.Cmd {
		return func() tea.Msg {
			return ui.CloseComponentMsg{ID: input.ID()}
		}
	})

	return func() tea.Msg {
		return dialog.OpenDialog(input, optOverlay...)
	}
}
