package picker

import (
	"charm.land/lipgloss/v2"
)

// View renders the file picker overlay.
//
// Layout structure:
//
//	Title
//	Current Directory
//	File List
//	Help Bar
//
// The entire block is wrapped in a styled container sized
// according to the current layout constraints.
func (m *Model) View() string {
	// Limit picker width so it doesn't become too wide
	width := min(70, m.Layout.Content.Width-10)
	height := pickerHeight(m.Layout)

	title := ""
	if m.Title != "" {
		title = m.titleStyle(width).Render(m.Title)
	}

	currentDir := m.currentDirStyle(width).Render(
		m.FilePicker.CurrentDirectory,
	)

	content := lipgloss.JoinVertical(
		lipgloss.Top,
		title,
		currentDir,
		m.FilePicker.View(),
		m.helpStyle(width).Render(m.helpView()),
	)

	return containerStyle(width, height).Render(content)
}

// helpView renders the picker keyboard help bar.
func (m *Model) helpView() string {
	return lipgloss.JoinHorizontal(
		lipgloss.Top,

		m.helpKeyStyle().Render("← →"),
		" dir  ",

		m.helpKeyStyle().Render("↑ ↓"),
		" move  ",

		m.helpKeyStyle().Render("enter"),
		" select  ",

		m.helpKeyStyle().Render("b/esc"),
		" close",
	)
}
