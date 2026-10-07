package input

import (
	"github.com/MohsenBg/bgscan/internal/ui/shared/dialog"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

// OpenInputDialog opens input as an overlay, closing it automatically
// after the input's own onSubmit handler succeeds. The input's own
// title and hints become the frame chrome by default; optOverlay may
// override them with WithTitle/WithHints.
func OpenInputDialog(input Dialog, optOverlay ...dialog.DialogOption) tea.Cmd {
	input.AppendOnSubmit(func() tea.Cmd {
		return func() tea.Msg {
			return ui.CloseComponentMsg{ID: input.ID()}
		}
	})

	base := []dialog.DialogOption{
		dialog.WithTitle(input.Title()),
		dialog.WithHints(input.Hints()...),
	}

	return func() tea.Msg {
		return dialog.OpenDialog(input, append(base, optOverlay...)...)
	}
}
