package table

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
)

func (m *Model) View() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	width := m.tableWidthLocked()

	tableView := m.tableViewStyle().Render(m.renderTableBody())
	helpView := m.helpStyle().Render(m.help.View())
	statusView := m.renderStatus()

	blocks := make([]string, 0, 5)

	if t := m.renderTitle(); t != "" {
		blocks = append(blocks, t)
	}

	if f := m.renderFilter(); f != "" {
		blocks = append(blocks, f)
	}

	blocks = append(blocks, tableView, statusView, helpView)

	return lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center).
		Render(lipgloss.JoinVertical(lipgloss.Center, blocks...))
}

// renderTableBody renders the table body. When there are no rows, it keeps
// the header visible and injects a centered message into the empty body.
func (m *Model) renderTableBody() string {
	view := m.bubble.View()

	if !m.loading && len(m.bubble.Rows()) > 0 {
		return m.markSelected(view)
	}

	// bubbles/table renders header text, header border, then body lines.
	lines := strings.Split(view, "\n")

	const headerEnd = 2
	if len(lines) <= headerEnd {
		return view
	}

	bodyLines := len(lines) - headerEnd
	contentWidth := m.contentWidthLocked()

	base := lipgloss.NewStyle().
		Width(contentWidth).
		Align(lipgloss.Center)

	var msgLines []string

	switch {
	case m.loading:
		msgLines = append(
			msgLines,
			base.Foreground(m.theme.Text).Render(m.renderLoading()),
		)

	case m.filterValue() != "":
		filter := m.filterValue()
		if runes := []rune(filter); len(runes) > maxFilterChar {
			filter = string(runes[:maxFilterChar-3]) + "..."
		}

		msgLines = append(
			msgLines,
			base.Foreground(m.theme.Text).Render(fmt.Sprintf("No matches for %q", filter)),
			base.Foreground(m.theme.Muted).Render(fmt.Sprintf(
				"press %s to clear the filter",
				m.keys.Filter.Clear.Keys[0],
			)),
		)

	default:
		msgLines = append(
			msgLines,
			base.Foreground(m.theme.Muted).Render(m.emptyMessage),
		)
	}

	start := max(headerEnd, headerEnd+(bodyLines-len(msgLines))/2)
	msgLen := len(msgLines)

	for i := headerEnd; i < len(lines); i++ {
		idx := i - start

		if idx >= 0 && idx < msgLen {
			lines[i] = msgLines[idx]
		} else if m.loading {
			lines[i] = ""
		}
	}

	return strings.Join(lines, "\n")
}

func (m *Model) renderTitle() string {
	if m.title == "" {
		return ""
	}

	return m.titleStyles(m.tableWidthLocked()).Render(m.title)
}

func (m *Model) renderFilter() string {
	if !m.filtering && m.filterValue() == "" {
		return ""
	}

	left := " / " + lipgloss.NewStyle().
		Foreground(m.theme.Text).
		Render(m.filter.View())

	right := lipgloss.NewStyle().
		Foreground(m.theme.Secondary).
		Render(fmt.Sprintf(
			"%d / %d ",
			len(m.bubble.Rows()),
			len(m.originalRows),
		))

	style := m.filterStyle()
	gap := max(1, style.GetWidth()-lipgloss.Width(left)-lipgloss.Width(right))

	return style.Render(left + strings.Repeat(" ", gap) + right)
}

// renderStatus draws the bottom border and status text, continuing the
// table's rounded border.
func (m *Model) renderStatus() string {
	total := len(m.originalRows)

	selected := 0
	if !m.loading && total > 0 {
		selected = m.bubble.Cursor() + 1
	}

	text := fmt.Sprintf(" Row %d of %d ──╯", selected, total)
	width := m.tableWidthLocked()

	left := "╰"
	dashes := max(0, width-lipgloss.Width(text)-lipgloss.Width(left))
	line := left + strings.Repeat("─", dashes) + text

	return lipgloss.NewStyle().
		Width(width).
		Foreground(m.theme.Info).
		Render(line)
}
