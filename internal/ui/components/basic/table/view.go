package table

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
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
		helpView = m.helpViewStyle(width).Render(
			truncateLines(m.Help.FullHelpView(m.Keys.FullHelp(width)), width-8),
		)
	} else {
		helpView = m.helpViewStyle(width).Render(m.shortHelpText(width - 8))
	}

	return helpView
}

// shortHelpText builds a single-line help string that always fits within
// width cells. Items that don't fit are dropped and truncation is marked
// with an ellipsis, so the help never wraps or overflows its line.
func (m *Model) shortHelpText(width int) string {
	var items []string
	for _, a := range m.Keys.Actions {
		if a.ShortHelp == "" {
			continue
		}
		items = append(items, a.ShortHelp)
	}

	const sep = " • "
	out := ""
	shown := 0
	for _, it := range items {
		candidate := it
		if shown > 0 {
			candidate = out + sep + it
		}
		if lipgloss.Width(candidate) > width {
			break
		}
		out = candidate
		shown++
	}

	if shown < len(items) {
		if out != "" {
			out += sep
		}
		out += "…"
	}
	return out
}

// truncateLines crops every line of s to at most width cells, preserving
// ANSI styling and never producing wrapped output.
func truncateLines(s string, width int) string {
	if width < 1 {
		return ""
	}
	lines := strings.Split(s, "\n")
	for i := range lines {
		lines[i] = ansi.Truncate(lines[i], width, "")
	}
	return strings.Join(lines, "\n")
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
