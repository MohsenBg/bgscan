// dialog.go
package input

import (
	tea "charm.land/bubbletea/v2"

	"github.com/MohsenBg/bgscan/internal/ui/shared/dialog"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
)

// OpenInputDialog wraps an input component in a dialog overlay.
// It automatically closes the dialog when the input's submit handler succeeds.
// The input's Title() and DialogHints() are used for the dialog frame by default,
// but can be overridden via opts.
func OpenInputDialog(input Dialog, opts ...dialog.DialogOption) tea.Cmd {
	input.AppendOnSubmit(func() tea.Cmd {
		return func() tea.Msg {
			return ui.CloseComponentMsg{ID: input.ID()}
		}
	})

	base := []dialog.DialogOption{
		dialog.WithHints(input.DialogHints()...),
	}

	return func() tea.Msg {
		return dialog.OpenDialog(input, append(base, opts...)...)
	}
}
