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

// NoticeUnderDevelopment returns a command that displays a standard
// "Under Development" informational notice.
//
// This helper is used for UI sections that are not yet implemented.
func NoticeUnderDevelopment(layout *layout.Layout) tea.Cmd {
	title := "Under Development"

	message := "This section is currently being built.\n" +
		"Thank you for your patience. Stay tuned for future updates."

	return NewNoticeCmd(layout, title, message, NOTICE_INFO)
}
