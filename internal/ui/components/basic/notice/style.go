package notice

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// levelStyle holds the palette used to render a notice at a given severity.
type levelStyle struct {
	TitleColor  color.Color
	BorderColor color.Color
	AccentColor color.Color
	Background  color.Color

	Icon       string
	FooterText string
}

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

// levelPalette maps a notice level to its rendering palette.
func (m *Model) levelPalette(level LEVEL) levelStyle {
	th := m.env.Theme
	switch level {

	case NOTICE_ERROR:
		return levelStyle{
			TitleColor:  th.Error,
			BorderColor: th.Error,
			AccentColor: th.Error,
			Icon:        iconFor(level),
			FooterText:  "Continue",
		}

	case NOTICE_SUCCESS:
		return levelStyle{
			TitleColor:  th.Success,
			BorderColor: th.Success,
			AccentColor: th.Success,
			Icon:        iconFor(level),
			FooterText:  "Done",
		}

	case NOTICE_INFO:
		fallthrough

	default:
		return levelStyle{
			TitleColor:  th.Info,
			BorderColor: th.Info,
			AccentColor: th.Info,
			Icon:        iconFor(level),
			FooterText:  "Continue",
		}
	}
}

func containerStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Left, lipgloss.Top)
}

// CenterStyle centers content horizontally inside the notice body.
func CenterStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width - 2).
		Align(lipgloss.Center)
}

func (m *Model) titleStyle(width int, level LEVEL) lipgloss.Style {
	p := m.levelPalette(level)

	return lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center).
		Bold(true).
		Foreground(p.TitleColor).
		MarginBottom(1)
}

// buttonStyle renders the notice action button ("Continue", "Done", ...).
func (m *Model) buttonStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(m.env.Theme.Primary).
		Align(lipgloss.Center).
		Padding(0, 2).
		MarginTop(1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.env.Theme.BorderActive)
}

// ButtonStyle keeps backward compatibility for external callers.
func ButtonStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Align(lipgloss.Center).
		Padding(0, 2).
		MarginTop(1).
		Border(lipgloss.RoundedBorder())
}
