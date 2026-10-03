package menu

import (
	"image/color"
	"strings"

	"github.com/MohsenBg/bgscan/internal/ui/shared/anim"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

const (
	// Row chrome: a 2-cell prefix ("  " or "▐ ") plus a 2-cell suffix.
	rowPrefix   = "  "
	rowSuffix   = "  "
	rowPadding  = 4
	minRowWidth = 10

	// Sidebar split: the list gets sidebarLeftPct% of the width, one
	// column goes to the separator, the rest is the detail pane.
	sidebarLeftPct   = 45
	minListWidth     = 24
	minDetailWidth   = 16
	sidebarSeparator = "│"

	// Scroll markers.
	markerIndent       = "  "
	minLinesForMarkers = 4
	markerLines        = 2

	// Detail pane.
	detailIndent  = "  "
	openHint      = "Press ⏎ to open"
	openHintShort = "⏎ open"
	minDetailText = 8

	// Selection tints blended onto the theme background.
	keycapTint   = 0.18
	selectorTint = 0.20
)

// View renders the menu as a block of exactly effectiveSize() cells.
func (m *Model) View() string {
	width, height := m.effectiveSize()
	visible := visibleLines(height)
	start, end, markers := m.window(visible)

	head := []string{
		lipgloss.PlaceHorizontal(width, lipgloss.Center, titleStyle(m.theme).Render(m.title)),
		"",
	}

	var body []string
	if m.sidebarMode(width) {
		body = m.sidebarLines(width, visible, start, end, markers)
	} else {
		body = m.listLines(width, start, end, markers)
	}

	foot := []string{"", m.helpLine(width)}

	for range min(height-len(head)-len(body)-len(foot), 1) {
		body = append(body, "")
	}

	lines := append(head, body...)
	lines = append(lines, foot...)

	// Keep every line exactly width cells wide.
	for i, line := range lines {
		switch w := lipgloss.Width(line); {
		case w < width:
			lines[i] = line + strings.Repeat(" ", width-w)
		case w > width:
			lines[i] = ansi.Truncate(line, width, "…")
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
		lines -= markerLines // room for the ↑ / ↓ markers
	}
	rows := itemsThatFit(lines)
	start = max(0, min(m.index-rows/2, n-rows))
	return start, start + rows, markers
}

// listLines renders the item rows (with gaps and scroll markers) in a
// column of the given width.
func (m *Model) listLines(width, start, end int, markers bool) []string {
	lines := make([]string, 0, itemsHeight(end-start)+markerLines)
	if markers && start > 0 {
		lines = append(lines, m.markerLine("↑"))
	}
	for i := start; i < end; i++ {
		if i > start {
			for range rowGap {
				lines = append(lines, "")
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
	th := resolveTheme(m.theme)
	// Guaranteed to satisfy minListWidth and minDetailWidth while keeping
	// leftW + separator + rightW == width exactly.
	leftW := min(max(width*sidebarLeftPct/100, minListWidth), width-1-minDetailWidth)
	rightW := width - leftW - 1

	left := m.listLines(leftW, start, end, markers)
	item, _ := m.GetSelected()
	detail := m.detailLines(item, rightW, visible)

	sep := lipgloss.NewStyle().Foreground(th.Border).Render(sidebarSeparator)
	lines := make([]string, 0, max(len(left), len(detail)))
	for i := range max(len(left), len(detail)) {
		lines = append(lines, padTo(left, i, leftW)+sep+padTo(detail, i, rightW))
	}
	return lines
}

// rowLine renders a single menu row, truncated to width.
func (m *Model) rowLine(item MenuItem, i, width int) string {
	th := resolveTheme(m.theme)
	selected := i == m.index

	// The entrance animation fades the whole row in from the background.
	fade := m.fader(m.rowAlpha(i))

	// Colors come from the theme at runtime, so these stay vars.
	iconFg, titleFg, keyFg, keyBg := th.Muted, th.Muted, th.Muted, th.Background
	if selected {
		iconFg, titleFg, keyFg, keyBg = th.Primary, th.Text, th.Primary,
			anim.Mix(th.Background, th.Primary, keycapTint)
	}

	var left string
	if m.showIcon {
		left += rowIconStyle(fade(iconFg), selected).Render(item.icon)
	}
	left += rowTitleStyle(fade(titleFg), selected).Render(item.title)

	var right string
	if m.showShortcut && item.shortcut != "" {
		right = keycapStyle(fade(keyFg), fade(keyBg)).Render(item.shortcut)
	}

	prefix := rowPrefix
	if selected {
		tint := fade(anim.Mix(th.Background, th.Primary, selectorTint))
		prefix = lipgloss.NewStyle().Foreground(tint).Render("▐") + " "
	}
	line := prefix + fillRow(left, right, max(width-rowPadding, minRowWidth)) + rowSuffix
	return ansi.Truncate(line, width, "…")
}

// detailLines renders the selected item's details: title, description,
// then the "open" hint and its shortcut key.
func (m *Model) detailLines(item MenuItem, w, avail int) []string {
	if avail < 1 {
		return nil
	}
	th := resolveTheme(m.theme)
	fade := m.fader(m.rowAlpha(m.index))
	text := func(c color.Color, bold bool) lipgloss.Style {
		return lipgloss.NewStyle().Foreground(fade(c)).Bold(bold)
	}
	inner := max(w-len(detailIndent), minDetailText)

	// Head: icon + title.
	head := detailIndent
	if m.showIcon && item.icon != "" {
		head += text(th.Primary, true).Render(item.icon) + "  "
	}
	head = ansi.Truncate(head+text(th.Text, true).Render(item.title), w, "…")

	// Description.
	var desc []string
	for _, line := range wrapWords(item.desc, max(inner-len(detailIndent), minDetailText)) {
		desc = append(desc, detailIndent+text(th.Muted, false).Render(line))
	}

	// Footer: rule, open hint, shortcut.
	rule := detailIndent + text(th.Border, false).Render(strings.Repeat("─", inner))

	openText := openHint
	if lipgloss.Width(openText)+len(detailIndent) > w {
		openText = openHintShort
	}
	footer := []string{"", rule, "", detailIndent + text(th.Text, false).Render(openText)}

	if m.showShortcut && item.shortcut != "" {
		keyBg := fade(anim.Mix(th.Background, th.Primary, keycapTint))
		key := keycapStyle(fade(th.Primary), keyBg).Render(item.shortcut)
		footer = append(footer, "", detailIndent+text(th.Muted, false).Render("Shortcut ")+key)
	}

	lines := append([]string{"", head, ""}, desc...)
	lines = append(lines, footer...)

	// Not enough room: drop the top spacer first, then cut from the bottom
	// so the title and description survive longest.
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
	return helpStyle(m.theme).Render(markerIndent + arrow)
}

// helpLine renders the key hints below the menu.
func (m *Model) helpLine(width int) string {
	hint := helpStyle(m.theme).Render("↑↓ move   ⏎ select   q quit")
	return lipgloss.PlaceHorizontal(width, lipgloss.Center, hint)
}

// fillRow fills the space between the left and right content.
func fillRow(left, right string, w int) string {
	const minGap = 1 // keep at least one space before the shortcut chip
	gap := max(w-lipgloss.Width(left)-lipgloss.Width(right), minGap)
	return left + strings.Repeat(" ", gap) + right
}

// padTo returns lines[i] padded (or truncated) to w cells, or blanks if
// i is out of range.
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

	var lines, current []string
	lineWidth := 0
	for _, word := range words {
		wordWidth := lipgloss.Width(word)
		if len(current) > 0 && lineWidth+1+wordWidth > w {
			lines = append(lines, strings.Join(current, " "))
			current, lineWidth = nil, 0
		}
		if len(current) > 0 {
			lineWidth++
		}
		current = append(current, word)
		lineWidth += wordWidth
	}
	return append(lines, strings.Join(current, " "))
}
