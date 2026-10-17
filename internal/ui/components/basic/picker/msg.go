package picker

import (
	tea "charm.land/bubbletea/v2"

	"github.com/MohsenBg/bgscan/internal/ui/shared/dialog"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
)

// OnSelect is called with the selected file path.
type OnSelect func(path string) tea.Cmd

// OpenFilePickerCmd opens the file picker in a dialog overlay.
func OpenFilePickerCmd(
	env *ui.Env,
	title string,
	baseDir string,
	allowedExt []string,
	onSelect OnSelect,
	options ...dialog.DialogOption,
) tea.Cmd {
	return func() tea.Msg {
		p := New(env, title, baseDir, allowedExt, onSelect)

		opts := []dialog.DialogOption{
			dialog.WithTitle(p.Name()),
			dialog.WithHints(p.DialogHints()...),
			dialog.WithAnimation(dialog.AnimRiseUp),
		}
		opts = append(opts, options...)

		return dialog.OpenDialog(p, opts...)
	}
}
