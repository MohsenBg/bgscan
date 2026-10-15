package menu

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/MohsenBg/bgscan/internal/ui/shared/anim"
)

const (
	rowSelector = "▐"
	rowSpacer   = " "

	sidebarLeftPct   = 45
	minListWidth     = 24
	minDetailWidth   = 16
	sidebarSeparator = "│"

	markerIndent       = "  "
	minLinesForMarkers = 4
	markerLines        = 2

	detailIndent  = "  "
	minDetailText = 12

	keycapTint   = 0.18
	selectorTint = 0.20
)

// View renders the menu as a block of exactly effectiveSize() cells.
func (m *Model) View() string {
	width, height := m.effectiveSize()
	innerWidth := max(1, width-2*contentPadding)

	visible := m.visibleLines(height)
	start, end, markers := m.window(visible)

	var head []string

	if m.title == "" {
		head = []string{""}
	} else {
		head = []string{
			lipgloss.PlaceHorizontal(
				innerWidth,
				lipgloss.Center,
				titleStyle(&m.env.Theme).Render(m.title),
			),
			"",
		}
	}

	var body []string

	if m.sidebarMode(innerWidth) {
		body = m.sidebarLines(innerWidth, visible, start, end, markers)
	} else {
		body = m.listLines(innerWidth, start, end, markers)
	}

	foot := []string{"", ""}

	if m.showHelp && m.help != nil {
		foot = append(foot, lipgloss.PlaceHorizontal(
			innerWidth,
			lipgloss.Center,
			m.help.View(),
		))
	}

	for range max(height-len(head)-len(body)-len(foot), 0) {
		body = append(body, "")
	}

	lines := append(head, body...)
	lines = append(lines, foot...)

	for i, line := range lines {
		switch w := lipgloss.Width(line); {
		case w < innerWidth:
			lines[i] = line + strings.Repeat(" ", innerWidth-w)
		case w > innerWidth:
			lines[i] = ansi.Truncate(line, innerWidth, "…")
		}
	}

	if contentPadding > 0 {
		pad := strings.Repeat(" ", contentPadding)

		for i := range lines {
			lines[i] = pad + lines[i] + pad
		}
	}

	return strings.Join(lines, "\n")
}

// sidebarMode enables the detail pane when there is enough room.
func (m *Model) sidebarMode(width int) bool {
	return width >= minSidebarWidth && m.hasDescriptions()
}

// window returns the visible item range for the given number of lines.
func (m *Model) window(lines int) (start, end int, markers bool) {
	n := len(m.items)

	if itemsHeight(n) <= lines {
		return 0, n, false
	}

	if lines >= minLinesForMarkers {
		markers = true
		lines -= markerLines
	}

	rows := itemsThatFit(lines)
	start = max(0, min(m.index-rows/2, n-rows))

	return start, start + rows, markers
}

// listLines renders item rows with gaps and scroll markers.
func (m *Model) listLines(width, start, end int, markers bool) []string {
	lines := make([]string, 0, itemsHeight(end-start)+markerLines)

	if markers && start > 0 {
		lines = append(lines, m.markerLine("↑"))
	}

	for i := start; i < end; i++ {
		if i > start {
			for range rowGap {
				lines = append(lines, " ")
			}
		}

		lines = append(lines, m.rowLine(m.items[i], i, width))
	}

	if markers && end < len(m.items) {
		lines = append(lines, m.markerLine("↓"))
	}

	return lines
}

// sidebarLines renders the item list next to a detail pane.
func (m *Model) sidebarLines(width, visible, start, end int, markers bool) []string {
	th := resolveTheme(&m.env.Theme)

	leftW := min(
		max(width*sidebarLeftPct/100, minListWidth),
		width-1-minDetailWidth,
	)
	leftW = max(leftW, minListWidth)

	rightW := max(width-leftW-1, minDetailWidth)

	left := m.listLines(leftW, start, end, markers)

	item, _ := m.GetSelected()
	detail := m.detailLines(item, rightW, visible)

	sep := lipgloss.NewStyle().
		Foreground(th.Border).
		Render(sidebarSeparator)

	lines := make([]string, 0, max(len(left), len(detail)))

	for i := range max(len(left), len(detail)) {
		lines = append(lines, padTo(left, i, leftW)+sep+padTo(detail, i, rightW))
	}

	return lines
}

// rowLine renders a single menu row.
func (m *Model) rowLine(item MenuItem, i, width int) string {
	th := resolveTheme(&m.env.Theme)
	selected := i == m.index

	fade := m.fader(m.rowAlpha(i))

	iconFg, titleFg, keyFg, keyBg := th.Muted, th.Muted, th.Muted, th.Background

	if selected {
		iconFg, titleFg, keyFg = th.Primary, th.Text, th.Primary
		keyBg = anim.Mix(th.Background, th.Primary, keycapTint)
	}

	prefix := rowSpacer

	if selected {
		tint := fade(anim.Mix(th.Background, th.Primary, selectorTint))
		prefix = lipgloss.NewStyle().
			Foreground(tint).
			Render(rowSelector)
	}

	left := rowSpacer

	if m.showIcon {
		left += rowIconStyle(fade(iconFg), selected).Render(item.icon)
	}

	left += rowTitleStyle(fade(titleFg), selected).Render(item.title)

	var right string

	if m.showShortcut && item.shortcut != "" {
		right = keycapStyle(fade(keyFg), fade(keyBg)).Render(item.shortcut)
	}

	core := fillRow(
		left,
		right,
		max(width-rowOverheadWidth, minRowWidth),
	)

	line := prefix + core + rowSpacer

	return ansi.Truncate(line, width, "…")
}

// detailLines renders the selected item details.
func (m *Model) detailLines(item MenuItem, w, avail int) []string {
	if avail < 1 {
		return nil
	}

	th := resolveTheme(&m.env.Theme)
	fade := m.fader(m.rowAlpha(m.index))

	text := func(c color.Color, bold bool) lipgloss.Style {
		return lipgloss.NewStyle().
			Foreground(fade(c)).
			Bold(bold)
	}

	inner := max(w-len(detailIndent), minDetailText)

	head := detailIndent

	if m.showIcon && item.icon != "" {
		head += text(th.Primary, true).Render(item.icon) + "  "
	}

	head = ansi.Truncate(
		head+text(th.Text, true).Render(item.title),
		w,
		"…",
	)

	var desc []string

	for _, line := range wrapWords(item.desc, max(inner-len(detailIndent), minDetailText)) {
		desc = append(desc, detailIndent+text(th.Muted, false).Render(line))
	}

	rule := detailIndent + text(th.Border, false).Render(strings.Repeat("─", inner))

	footer := []string{"", rule}

	if m.showShortcut && item.shortcut != "" {
		keyBg := fade(anim.Mix(th.Background, th.Primary, keycapTint))
		key := keycapStyle(fade(th.Primary), keyBg).Render(item.shortcut)

		footer = append(
			footer,
			"",
			detailIndent+text(th.Muted, false).Render("Shortcut ")+key,
		)
	}

	lines := append([]string{"", head, ""}, desc...)
	lines = append(lines, footer...)

	if len(lines) > avail {
		lines = lines[1:]
	}

	if len(lines) > avail {
		lines = lines[:avail]
	}

	return lines
}

// markerLine renders a scroll marker.
func (m *Model) markerLine(arrow string) string {
	return helpStyle(&m.env.Theme).Render(markerIndent + arrow)
}

// fillRow fills the space between left and right content.
func fillRow(left, right string, w int) string {
	if right == "" {
		return left
	}

	gap := max(
		w-lipgloss.Width(left)-lipgloss.Width(right),
		shortcutGapWidth,
	)

	return left + strings.Repeat(" ", gap) + right
}

// padTo returns lines[i] padded or truncated to w cells.
func padTo(lines []string, i, w int) string {
	if i >= len(lines) {
		return strings.Repeat(" ", w)
	}

	s := lines[i]

	if pad := w - lipgloss.Width(s); pad > 0 {
		return s + strings.Repeat(" ", pad)
	}

	return ansi.Truncate(s, w, "")
}

// wrapWords wraps text at w cells.
func wrapWords(text string, w int) []string {
	w = max(w, 1)

	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}

	var (
		lines   []string
		current []string

		lineWidth int
	)

	for _, word := range words {
		wordWidth := lipgloss.Width(word)

		if len(current) > 0 && lineWidth+1+wordWidth > w {
			lines = append(lines, strings.Join(current, " "))
			current = nil
			lineWidth = 0
		}

		if len(current) > 0 {
			lineWidth++
		}

		current = append(current, word)
		lineWidth += wordWidth
	}

	return append(lines, strings.Join(current, " "))
}
