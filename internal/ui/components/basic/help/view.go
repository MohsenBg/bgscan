package help

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/MohsenBg/bgscan/internal/ui/shared/keys"
)

const ellipsis = "…"

func joinLines(lines []string) string {
	return strings.Join(lines, "\n")
}

func (m *Model) View() string {
	if m.mode == LongHelp {
		return m.longHelpStyle().Render(m.vp.View())
	}

	return joinLines(m.centerShortLines(m.shortLines()))
}

func (m *Model) renderHint(hint keys.Hint) string {
	return m.keyStyle().Render(hint.Key) +
		" " +
		m.labelStyle().Render(hint.Label)
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

type shortLine struct {
	text string
	used int
}

func (m *Model) shortLines() []string {
	contentWidth := m.shortContentWidth()
	if contentWidth <= 0 || m.shortHelpHeight <= 0 || len(m.help.Short) == 0 {
		return nil
	}

	separator := m.sepStyle().Render(" • ")
	separatorWidth := lipgloss.Width(separator)

	var (
		lines   []shortLine
		current strings.Builder
		curW    int
		used    int
	)

	flush := func() {
		if current.Len() == 0 {
			return
		}

		lines = append(lines, shortLine{
			text: current.String(),
			used: used,
		})

		current.Reset()
		curW = 0
	}

	for i, hint := range m.help.Short {
		item := m.renderHint(hint)
		itemW := lipgloss.Width(item)

		if itemW > contentWidth {
			item = ansi.Truncate(item, contentWidth, ellipsis)
			itemW = lipgloss.Width(item)
		}

		if curW == 0 {
			current.WriteString(item)
			curW = itemW
			used = i + 1
			continue
		}

		if curW+separatorWidth+itemW <= contentWidth {
			current.WriteString(separator)
			current.WriteString(item)
			curW += separatorWidth + itemW
			used = i + 1
			continue
		}

		flush()

		current.WriteString(item)
		curW = itemW
		used = i + 1
	}

	flush()

	if len(lines) == 0 {
		return nil
	}

	if len(lines) <= m.shortHelpHeight {
		out := make([]string, len(lines))
		for i, line := range lines {
			out[i] = line.text
		}
		return out
	}

	if m.shortHelpHeight == 1 {
		line := lines[0]
		if remaining := len(m.help.Short) - line.used; remaining > 0 {
			suffix := separator + m.labelStyle().Render(fmt.Sprintf("+%d more", remaining))
			line.text = ansi.Truncate(line.text+suffix, contentWidth, ellipsis)
		}

		return []string{line.text}
	}

	keep := m.shortHelpHeight - 1
	out := make([]string, 0, m.shortHelpHeight)

	for i := 0; i < keep && i < len(lines); i++ {
		out = append(out, lines[i].text)
	}

	used = lines[keep-1].used
	if remaining := len(m.help.Short) - used; remaining > 0 {
		out = append(out, ansi.Truncate(
			m.labelStyle().Render(fmt.Sprintf("+%d more", remaining)),
			contentWidth,
			ellipsis,
		))
	}

	if len(out) > m.shortHelpHeight {
		out = out[:m.shortHelpHeight]
	}

	return out
}

func (m *Model) centerShortLines(lines []string) []string {
	if len(lines) == 0 {
		return nil
	}

	padWidth := m.shortPadding()
	innerWidth := m.shortContentWidth()
	pad := strings.Repeat(" ", padWidth)

	out := make([]string, len(lines))

	for i, line := range lines {
		w := lipgloss.Width(line)

		if w > innerWidth {
			line = ansi.Truncate(line, innerWidth, ellipsis)
			w = lipgloss.Width(line)
		}

		gap := innerWidth - w
		left := gap / 2
		right := gap - left

		out[i] = pad +
			strings.Repeat(" ", left) +
			line +
			strings.Repeat(" ", right) +
			pad
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
			var line strings.Builder

			if i == 0 {
				line.WriteString(key)
				line.WriteString(keyPad)
				line.WriteString(gap)
			} else {
				line.WriteString(blank)
			}

			line.WriteString(m.labelStyle().Render(labelLine))
			lines = append(lines, line.String())
		}
	}

	return lines
}

// wrapText splits plain text into visible-width-aware lines.
func wrapText(s string, width int) []string {
	s = strings.TrimSpace(s)

	if width <= 0 {
		return []string{s}
	}

	if s == "" {
		return []string{""}
	}

	words := strings.Fields(s)

	var (
		lines     []string
		line      strings.Builder
		lineWidth int
	)

	flush := func() {
		if line.Len() == 0 {
			return
		}

		lines = append(lines, line.String())
		line.Reset()
		lineWidth = 0
	}

	for _, word := range words {
		wordWidth := lipgloss.Width(word)

		if wordWidth <= width {
			if lineWidth == 0 {
				line.WriteString(word)
				lineWidth = wordWidth
				continue
			}

			if lineWidth+1+wordWidth <= width {
				line.WriteString(" ")
				line.WriteString(word)
				lineWidth += 1 + wordWidth
				continue
			}

			flush()

			line.WriteString(word)
			lineWidth = wordWidth

			continue
		}

		flush()

		var (
			part      strings.Builder
			partWidth int
		)

		for _, r := range word {
			rWidth := lipgloss.Width(string(r))

			if partWidth+rWidth > width && part.Len() > 0 {
				lines = append(lines, part.String())
				part.Reset()
				partWidth = 0
			}

			part.WriteRune(r)
			partWidth += rWidth
		}

		if part.Len() > 0 {
			lines = append(lines, part.String())
		}
	}

	flush()

	if len(lines) == 0 {
		return []string{""}
	}

	return lines
}
