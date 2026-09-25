package table

import (
	"fmt"

	"charm.land/lipgloss/v2"
)

func (m *Model) View() string {
	width := m.Layout.Body.Width

	tableView := m.tableViewStyle(width).Render(m.BubbleTable.View())

	return lipgloss.NewStyle().
		Width(width).
		Render(lipgloss.JoinVertical(
			lipgloss.Center,
			m.renderTitle(),
			m.renderFilter(),
			tableView,
			m.renderHelpView(),
		))
}

func (m *Model) renderHelpView() string {
	width := m.Layout.Body.Width
	m.Help.SetWidth(width)

	helpView := ""
	if m.FullHelp {
		helpView = m.helpViewStyle(width).Render(m.Help.FullHelpView(m.Keys.FullHelp(m.Layout.Body.Width)))
	} else {
		helpView = m.helpViewStyle(width).Render(m.Help.ShortHelpView(m.Keys.ShortHelp()))
	}

	return helpView
}

func (m *Model) renderTitle() string {
	width := m.Layout.Body.Width
	if m.Title != "" {
		return m.titleStyles(width).Render(m.Title)
	}
	return ""
}

// renderFilter shows "filter: text [matched/total]" while editing or while
// a filter is applied; empty otherwise so the layout doesn't shift.
func (m *Model) renderFilter() string {
	if !m.filtering && m.filter == "" {
		return ""
	}
	text := m.filter
	if m.filtering {
		text += "▊"
	}
	return m.filterStyle().Render(fmt.Sprintf("filter: %s [%d/%d]",
		text, len(m.BubbleTable.Rows()), len(m.originalRows)))
}
