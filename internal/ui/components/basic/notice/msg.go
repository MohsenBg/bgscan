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
			dialog.WithHints(NoticeHints(*env.Keys)...),
		}

		// Errors and other destructive feedback get the danger frame.
		if level == NOTICE_ERROR {
			opts = append(opts, dialog.WithDanger())
		}

		opts = append(opts, options...)

		return dialog.OpenDialog(
			New(env, title, message, level),
			opts...,
		)
	}
}
