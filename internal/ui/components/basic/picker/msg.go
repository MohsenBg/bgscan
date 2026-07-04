package picker

import (
	"bgscan/internal/ui/shared/dialog"
	"bgscan/internal/ui/shared/layout"

	tea "charm.land/bubbletea/v2"
)

// OnSelect is called when the user selects a file.
// It receives the selected file path and may return a Bubble Tea command
// to trigger further application logic.
type OnSelect func(path string) tea.Cmd

// OpenFilePickerCmd returns a Bubble Tea command that opens a file picker
// inside a dialog overlay.
//
// The picker is created and pushed into the dialog system when the command
// is executed.
//
// Parameters:
//   - l           : layout manager used for sizing and positioning
//   - title       : title displayed in the picker header
//   - baseDir     : initial directory (defaults to home directory if empty)
//   - allowedExt  : list of allowed file extensions (e.g. ".txt", ".csv")
//   - onSelect    : callback executed when a file is selected
//   - options     : optional dialog configuration (positioning, style, etc.)
//
// The picker is centered by default unless overridden via options.
func OpenFilePickerCmd(
	l *layout.Layout,
	title string,
	baseDir string,
	allowedExt []string,
	onSelect OnSelect,
	options ...dialog.DialogOption,
) tea.Cmd {
	return func() tea.Msg {
		p := New(l, title, baseDir, allowedExt, onSelect)

		return dialog.OpenDialog(
			p,
			options...,
		)
	}
}
