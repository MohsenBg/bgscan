package table

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
)

func (m *Model) View() string {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.refreshCacheLocked()

	helpView := m.helpStyle().Render(m.help.View())

	blocks := make([]string, 0, 5)

	if m.cacheTitle != "" {
		blocks = append(blocks, m.cacheTitle)
	}

	if f := m.renderFilter(); f != "" {
		blocks = append(blocks, f)
	}

	blocks = append(blocks, m.cacheTable, m.cacheStatus, helpView)

	return lipgloss.NewStyle().
		Width(m.tableWidthLocked()).
		Align(lipgloss.Center).
		Render(lipgloss.JoinVertical(lipgloss.Center, blocks...))
}

// refreshCacheLocked re-renders the cached blocks when something changed.
func (m *Model) refreshCacheLocked() {
	if !m.dirty {
		return
	}

	m.cacheTitle = m.renderTitle()
	m.cacheTable = m.tableViewStyle().Render(m.renderTableBody())
	m.cacheStatus = m.renderStatus()
	m.dirty = false
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

	base := lipgloss.NewStyle().
		Width(m.contentWidthLocked()).
		Align(lipgloss.Center)

	var msgLines []string

	switch {
	case m.loading:
		msgLines = append(
			msgLines,
			base.Foreground(m.env.Theme.Text).Render(m.renderLoading()),
		)

	case m.filterValue() != "":
		filter := m.filterValue()
		if runes := []rune(filter); len(runes) > maxFilterChar {
			filter = string(runes[:maxFilterChar-3]) + "..."
		}

		msgLines = append(
			msgLines,
			base.Foreground(m.env.Theme.Text).Render(fmt.Sprintf("No matches for %q", filter)),
			base.Foreground(m.env.Theme.Muted).Render(fmt.Sprintf(
				"press %s to clear the filter",
				m.env.Keys.Filter.Clear.Keys[0],
			)),
		)

	default:
		msgLines = append(
			msgLines,
			base.Foreground(m.env.Theme.Muted).Render(m.emptyMessage),
		)
	}

	start := max(headerEnd, headerEnd+(bodyLines-len(msgLines))/2)

	for i := headerEnd; i < len(lines); i++ {
		idx := i - start

		switch {
		case idx >= 0 && idx < len(msgLines):
			lines[i] = msgLines[idx]
		case m.loading:
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
		Foreground(m.env.Theme.Text).
		Render(m.filter.View())

	right := lipgloss.NewStyle().
		Foreground(m.env.Theme.Secondary).
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
		Foreground(m.env.Theme.Info).
		Render(line)
}
