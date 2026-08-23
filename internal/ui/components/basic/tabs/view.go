package tabs

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// View renders the tabs row.
func (m *Model[T]) View() string {
	if len(m.tabs) == 0 {
		return ""
	}

	// Build the row of styled tab labels, tracking rendered widths so the
	// border line beneath can line up with the active tab exactly.
	var parts []string
	var widths []int
	for i, tab := range m.tabs {
		active := i == m.idx

		label := tab.Label
		style := m.inactiveStyle
		if active {
			style = m.activeStyle
		}

		styled := style.Render(label)
		parts = append(parts, styled)
		widths = append(widths, lipgloss.Width(styled))
	}

	sepWidth := lipgloss.Width(m.separator)
	row := strings.Join(parts, m.separator)

	// Compute overall width up front, since the border line needs it
	// whether or not it also gets used for the final Width/Align pass.
	width := m.layout.Body.Width
	if m.maxWidth > 0 && width > m.maxWidth {
		width = m.maxWidth
	}

	out := row

	if m.underline {
		// One continuous line, full width. The segment beneath the active
		// tab is drawn in the accent color; everything else — including
		// the gap after the last tab — is muted, so it reads as a single
		// border with a highlighted stretch rather than separate bars.
		offset := 0
		for i := 0; i < m.idx; i++ {
			offset += widths[i] + sepWidth
		}
		activeEnd := offset + widths[m.idx]

		lineWidth := width
		if lineWidth < offset+widths[m.idx] {
			lineWidth = offset + widths[m.idx]
		}

		var b strings.Builder
		b.WriteString(inactiveBorderStyle().Render(strings.Repeat("─", offset)))
		b.WriteString(activeBorderStyle().Render(strings.Repeat("─", widths[m.idx])))
		if rest := lineWidth - activeEnd; rest > 0 {
			b.WriteString(inactiveBorderStyle().Render(strings.Repeat("─", rest)))
		}

		out = lipgloss.JoinVertical(lipgloss.Left, row, b.String())
	}

	out = lipgloss.NewStyle().
		Width(width).
		Align(m.alignment).
		Render(out)

	return out
}
