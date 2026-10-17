package tabs

import (
	"image/color"
	"math"
	"strings"

	"charm.land/lipgloss/v2"
)

func (m *Model[T]) View() string {
	if len(m.tabs) == 0 {
		return ""
	}

	if m.animated {
		return m.viewSlideAccent()
	}

	row, widths := m.renderLabels()
	width := m.frameWidth()
	out := row

	if m.underline {
		line := m.underlineRow(widths, width)
		out = lipgloss.JoinVertical(lipgloss.Left, row, line)
	}

	return lipgloss.NewStyle().
		Width(width).
		Align(m.alignment).
		Render(out)
}

// viewSlideAccent draws an underline whose accent segment slides between tabs.
func (m *Model[T]) viewSlideAccent() string {
	row, widths := m.renderLabels()
	width := m.frameWidth()

	start := int(math.Round(m.slide))
	segment := widths[m.idx]
	lineWidth := max(width, start+segment)

	th := m.env.Theme
	line := columns(lineWidth, func(i int) color.Color {
		if i >= start && i < start+segment {
			return th.Primary
		}
		return th.Border
	})

	out := lipgloss.JoinVertical(lipgloss.Left, row, line)

	return lipgloss.NewStyle().
		Width(width).
		Align(m.alignment).
		Render(out)
}

func (m *Model[T]) renderLabels() (string, []int) {
	parts := make([]string, len(m.tabs))
	widths := make([]int, len(m.tabs))

	for i, tab := range m.tabs {
		rendered := m.styleFor(i).Render(tab.Label)
		parts[i] = rendered
		widths[i] = lipgloss.Width(rendered)
	}

	return strings.Join(parts, m.separator), widths
}

func (m *Model[T]) underlineRow(widths []int, width int) string {
	sepWidth := lipgloss.Width(m.separator)

	offset := 0
	for i := 0; i < m.idx; i++ {
		offset += widths[i] + sepWidth
	}

	activeWidth := widths[m.idx]
	activeEnd := offset + activeWidth
	lineWidth := max(width, activeEnd)

	var b strings.Builder

	if offset > 0 {
		b.WriteString(m.inactiveBorderStyle().Render(strings.Repeat("─", offset)))
	}
	if activeWidth > 0 {
		b.WriteString(m.activeBorderStyle().Render(strings.Repeat("─", activeWidth)))
	}
	if rest := lineWidth - activeEnd; rest > 0 {
		b.WriteString(m.inactiveBorderStyle().Render(strings.Repeat("─", rest)))
	}

	return b.String()
}

// tabOffset returns the start column of the tab at idx.
func (m *Model[T]) tabOffset(idx int) int {
	sepWidth := lipgloss.Width(m.separator)

	offset := 0
	for i := 0; i < idx && i < len(m.tabs); i++ {
		offset += m.labelWidth(i) + sepWidth
	}
	return offset
}

func (m *Model[T]) labelWidth(i int) int {
	return lipgloss.Width(m.styleFor(i).Render(m.tabs[i].Label))
}

// columns renders ─ across width columns, grouping adjacent columns with the
// same color to keep the output compact.
func columns(width int, colorAt func(i int) color.Color) string {
	var b strings.Builder

	for i := 0; i < width; {
		c := colorAt(i)

		j := i + 1
		for j < width && sameColor(c, colorAt(j)) {
			j++
		}

		b.WriteString(
			lipgloss.NewStyle().
				Foreground(c).
				Render(strings.Repeat("─", j-i)),
		)

		i = j
	}

	return b.String()
}

func sameColor(a, b color.Color) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}

	ar, ag, ab, aa := a.RGBA()
	br, bg, bb, ba := b.RGBA()

	return ar == br && ag == bg && ab == bb && aa == ba
}
