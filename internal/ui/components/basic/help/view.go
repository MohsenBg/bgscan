package help

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/MohsenBg/bgscan/internal/ui/shared/keys"
)

const ellipsis = "…"

func (m *Model) View() string {
	if m.mode == LongHelp {
		return m.longHelpStyle().Render(m.vp.View())
	}

	return strings.Join(m.centerShortLines(m.shortLines()), "\n")
}

func (m *Model) renderHint(hint keys.Hint) string {
	return m.keyStyle().Render(hint.Key) + " " + m.labelStyle().Render(hint.Label)
}

func (m *Model) separator() string {
	return m.sepStyle().Render(" • ")
}

func (m *Model) moreLabel(n int) string {
	return m.labelStyle().Render(fmt.Sprintf("+%d more", n))
}

func (m *Model) shortPadding() int {
	if m.shortHelpWidth <= 1 {
		return 0
	}

	return min(shortHelpPadding, (m.shortHelpWidth-1)/2)
}

func (m *Model) shortContentWidth() int {
	return max(1, m.shortHelpWidth-2*m.shortPadding())
}

// shortLine is one packed row of hints. used is the number of hints
type shortLine struct {
	text string
	used int
}

// shortLines packs the short hints into at most shortHelpHeight rows.
func (m *Model) shortLines() []string {
	width := m.shortContentWidth()
	height := m.shortHelpHeight
	total := len(m.help.Short)

	if width <= 0 || height <= 0 || total == 0 {
		return nil
	}

	lines := m.packShortHints(width)
	if len(lines) <= height {
		return lineTexts(lines)
	}

	// Everything fits on one row only if we squeeze the overflow marker in.
	if height == 1 {
		first := lines[0]
		text := first.text + m.separator() + m.moreLabel(total-first.used)

		return []string{ansi.Truncate(text, width, ellipsis)}
	}

	// Keep height-1 full rows and spend the last row on the overflow marker.
	keep := height - 1
	out := lineTexts(lines[:keep])
	more := m.moreLabel(total - lines[keep-1].used)

	return append(out, ansi.Truncate(more, width, ellipsis))
}

// packShortHints greedily fills rows of at most width cells.
func (m *Model) packShortHints(width int) []shortLine {
	sep := m.separator()
	sepW := lipgloss.Width(sep)

	var (
		lines []shortLine
		cur   strings.Builder
		curW  int
	)

	flush := func(used int) {
		if cur.Len() == 0 {
			return
		}

		lines = append(lines, shortLine{text: cur.String(), used: used})
		cur.Reset()
		curW = 0
	}

	for i, hint := range m.help.Short {
		item := ansi.Truncate(m.renderHint(hint), width, ellipsis)
		itemW := lipgloss.Width(item)

		if curW > 0 && curW+sepW+itemW > width {
			flush(i)
		}

		if curW > 0 {
			cur.WriteString(sep)
			curW += sepW
		}

		cur.WriteString(item)
		curW += itemW
	}

	flush(len(m.help.Short))

	return lines
}

func lineTexts(lines []shortLine) []string {
	out := make([]string, len(lines))
	for i, line := range lines {
		out[i] = line.text
	}

	return out
}

// centerShortLines centers each row inside the padded content area.
func (m *Model) centerShortLines(lines []string) []string {
	inner := m.shortContentWidth()
	pad := strings.Repeat(" ", m.shortPadding())

	out := make([]string, len(lines))
	for i, line := range lines {
		// PlaceHorizontal leaves over-wide lines alone, so clamp first.
		line = ansi.Truncate(line, inner, ellipsis)
		out[i] = pad + lipgloss.PlaceHorizontal(inner, lipgloss.Center, line) + pad
	}

	return out
}

// longLines renders aligned keys and wrapped labels within width.
func (m *Model) longLines(width int) []string {
	width = max(1, width)

	keyWidth := m.longKeyWidth()
	labelWidth := max(1, width-keyWidth-keyLabelGap)

	gap := strings.Repeat(" ", keyLabelGap)
	blank := strings.Repeat(" ", keyWidth+keyLabelGap)

	lines := make([]string, 0, len(m.help.Long))

	for _, hint := range m.help.Long {
		key := m.keyStyle().Render(hint.Key)
		keyPad := strings.Repeat(" ", max(0, keyWidth-lipgloss.Width(hint.Key)))

		for i, labelLine := range wrapText(hint.Label, labelWidth) {
			prefix := blank
			if i == 0 {
				prefix = key + keyPad + gap
			}

			lines = append(lines, prefix+m.labelStyle().Render(labelLine))
		}
	}

	return lines
}

// wrapText splits plain text into visible-width-aware lines.
func wrapText(s string, width int) []string {
	s = strings.TrimSpace(s)
	if width <= 0 || s == "" {
		return []string{s}
	}

	var (
		lines []string
		line  strings.Builder
		lineW int
	)

	flush := func() {
		if line.Len() == 0 {
			return
		}

		lines = append(lines, line.String())
		line.Reset()
		lineW = 0
	}

	for word := range strings.FieldsSeq(s) {
		wordW := lipgloss.Width(word)

		if wordW > width {
			flush()
			lines = append(lines, splitWord(word, width)...)

			continue
		}

		switch {
		case lineW == 0:
			// First word on the line: no leading space.
		case lineW+1+wordW <= width:
			line.WriteByte(' ')
			lineW++
		default:
			flush()
		}

		line.WriteString(word)
		lineW += wordW
	}

	flush()

	if len(lines) == 0 {
		return []string{""}
	}

	return lines
}

// splitWord breaks a single word into chunks of at most width cells.
func splitWord(word string, width int) []string {
	var (
		parts []string
		part  strings.Builder
		partW int
	)

	for _, r := range word {
		rw := lipgloss.Width(string(r))

		if partW+rw > width && part.Len() > 0 {
			parts = append(parts, part.String())
			part.Reset()
			partW = 0
		}

		part.WriteRune(r)
		partW += rw
	}

	if part.Len() > 0 {
		parts = append(parts, part.String())
	}

	return parts
}
