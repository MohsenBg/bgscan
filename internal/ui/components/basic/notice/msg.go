package notice

import (
	"bgscan/internal/ui/shared/dialog"
	"bgscan/internal/ui/shared/layout"

	tea "charm.land/bubbletea/v2"
)

// NewNoticeCmd opens a Notice overlay centered horizontally near the top
// of the screen.
func NewNoticeCmd(
	l *layout.Layout,
	title string,
	message string,
	level LEVEL,
	options ...dialog.DialogOption,
) tea.Cmd {
	return func() tea.Msg {
		opts := []dialog.DialogOption{
			dialog.WithPosition(dialog.Center, dialog.Top),
			dialog.WithOffset(0, 5),
		}

		opts = append(opts, options...)

		return dialog.OpenDialog(
			New(l, title, message, level),
			opts...,
		)
	}
}
