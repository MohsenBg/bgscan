package tabs

import (
	"bgscan/internal/ui/theme"

	"charm.land/lipgloss/v2"
)

func (m *Model[T]) View() string {
	if len(m.tabs) == 0 {
		return ""
	}

	var tabs []string

	active := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Current().Text).
		Background(theme.Current().Purple).
		Padding(0, 2)

	inactive := lipgloss.NewStyle().
		Foreground(theme.Current().Muted).
		Padding(0, 2)

	for i, tab := range m.tabs {
		if i == m.idx {
			tabs = append(tabs, active.Render(tab.Label))
		} else {
			tabs = append(tabs, inactive.Render(tab.Label))
		}
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)

	border := lipgloss.NewStyle().Width(min(m.layout.Body.Width-5, m.maxWidth)).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(theme.Current().Border)

	return lipgloss.NewStyle().Align(lipgloss.Center).Width(m.layout.Body.Width).Render(border.Render(row))
}
