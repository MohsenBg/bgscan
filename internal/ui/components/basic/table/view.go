package table

import (
	"strings"

	"bgscan/internal/logger"

	"charm.land/lipgloss/v2"
)

func (m *Model) View() string {
	width := m.Layout.Body.Width

	tableView := tableViewStyle(width).Render(m.BubbleTable.View())
	// Temporary — in View()
	logger.DebugInfo("DEBUG layout w=%d tableW=%d tablePHad=%d", m.Layout.Body.Width, m.tableWidthLocked(), tableHPad)
	rendered := m.BubbleTable.View()
	lines := strings.Split(rendered, "\n")
	if len(lines) > 0 {
		logger.DebugInfo("DEBUG rendered line width=%d", len([]rune(lines[0])))
	}

	return lipgloss.NewStyle().
		Width(width).
		Render(lipgloss.JoinVertical(
			lipgloss.Center,
			m.renderTitle(),
			tableView,
			m.renderHelpView(),
		))
}

func (m *Model) renderHelpView() string {
	width := m.Layout.Body.Width
	m.Help.SetWidth(width)

	helpView := ""
	if m.FullHelp {
		helpView = helpViewStyle(width).Render(m.Help.FullHelpView(m.Keys.FullHelp(m.Layout.Body.Width)))
	} else {
		helpView = helpViewStyle(width).Render(m.Help.ShortHelpView(m.Keys.ShortHelp()))
	}

	return helpView
}

func (m *Model) renderTitle() string {
	width := m.Layout.Body.Width
	if m.Title != "" {
		return titleStyles(width).Render(m.Title)
	}
	return ""
}
