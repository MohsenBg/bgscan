package notice

import (
	"github.com/MohsenBg/bgscan/internal/ui/shared/dialog"
	"github.com/MohsenBg/bgscan/internal/ui/shared/keys"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

func NoticeHints(k keys.Keys) keys.Hints {
	return keys.Hints{
		k.Notice.Continue.ShortItem(),
	}
}

// NewNoticeCmd opens a Notice overlay centered horizontally near the top
// of the screen.
func NewNoticeCmd(
	env *ui.Env,
	title string,
	message string,
	level LEVEL,
	options ...dialog.DialogOption,
) tea.Cmd {
	return func() tea.Msg {
		opts := []dialog.DialogOption{
			dialog.WithPosition(dialog.Center, dialog.Top),
			dialog.WithOffset(0, 5),
			dialog.WithTitle(FrameTitle(level, title)),
			dialog.WithAnimation(dialog.AnimSplit),
			dialog.WithHints(NoticeHints(*env.Keys)...),
		}

		// Errors and other destructive feedback get the danger frame.
		switch level {
		case NOTICE_ERROR:
			opts = append(opts, dialog.WithDanger())
		case NOTICE_SUCCESS:
			opts = append(opts, dialog.WithSuccess())
		case NOTICE_INFO:
			opts = append(opts, dialog.WithStyle(dialog.StyleInfo))
		}

		opts = append(opts, options...)

		return dialog.OpenDialog(
			new(env, message, level),
			opts...,
		)
	}
}

// NewKeyboardLayoutNoticeCmd warns the user that their keyboard layout
// is not English, so key shortcuts may not work.
func NewKeyboardLayoutNoticeCmd(env *ui.Env, close tea.Cmd) tea.Cmd {
	return NewNoticeCmd(
		env,
		"Keyboard Layout",
		"Your keyboard layout isn't English, so shortcuts may not work. Switch to English and try again.",
		NOTICE_INFO,
		dialog.WithStyle(dialog.StyleWarn),
		dialog.WithOnClose(close),
	)
}
