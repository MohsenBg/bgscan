package notice

import (
	"charm.land/lipgloss/v2"
)

// iconFor returns the severity icon prefix for a notice level.
// Icons are static strings; colors stay with the workspace theme.
func iconFor(level LEVEL) string {
	switch level {
	case NOTICE_ERROR:
		return "[×] "
	case NOTICE_SUCCESS:
		return "[✓] "
	case NOTICE_INFO:
		fallthrough
	default:
		return "[i] "
	}
}

// FrameTitle returns the raw dialog-frame title for a notice:
// severity icon plus title. The workspace styles it.
func FrameTitle(level LEVEL, title string) string {
	return iconFor(level) + title
}

func containerStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Left, lipgloss.Top)
}
