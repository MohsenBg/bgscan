package notice

import (
	"image/color"

	"bgscan/internal/ui/theme"

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

// levelPalette maps a notice level to its rendering palette.
func levelPalette(level LEVEL) levelStyle {
	switch level {

	case NOTICE_ERROR:
		return levelStyle{
			TitleColor:  theme.Current().Error,
			BorderColor: theme.Current().Error,
			AccentColor: theme.Current().Error,
			Icon:        "[×] ",
			FooterText:  "Continue",
		}

	case NOTICE_SUCCESS:
		return levelStyle{
			TitleColor:  theme.Current().Success,
			BorderColor: theme.Current().Success,
			AccentColor: theme.Current().Success,
			Icon:        "[✓] ",
			FooterText:  "Done",
		}

	case NOTICE_INFO:
		fallthrough

	default:
		return levelStyle{
			TitleColor:  theme.Current().Info,
			BorderColor: theme.Current().Info,
			AccentColor: theme.Current().Info,
			Icon:        "[i] ",
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

func titleStyle(width int, level LEVEL) lipgloss.Style {
	p := levelPalette(level)

	return lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center).
		Bold(true).
		Foreground(p.TitleColor).
		MarginBottom(1)
}

// ButtonStyle renders the notice action button ("Continue", "Done", ...).
func ButtonStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Primary).
		Align(lipgloss.Center).
		Padding(0, 2).
		MarginTop(1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Current().BorderActive)
}
