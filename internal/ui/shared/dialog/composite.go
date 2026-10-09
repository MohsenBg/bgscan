package dialog

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// Composite draws fg over bg, revealing only the cells covered by a at
// progress t.
//
// Cells outside the reveal box keep the background, so a partially open
// dialog does not punch a hole in the screen behind it.
func Composite(
	fg, bg string,
	xPos, yPos DialogPosition,
	xOff, yOff int,
	a Animation,
	t float64,
) string {
	if fg == "" {
		return bg
	}

	if bg == "" {
		return fg
	}

	fgW, fgH := lipgloss.Size(fg)
	bgW, bgH := lipgloss.Size(bg)

	if fgW >= bgW && fgH >= bgH {
		return fg
	}

	x, y := offsets(fgW, fgH, bgW, bgH, xPos, yPos, xOff, yOff)
	x = clamp(x, 0, bgW-fgW)
	y = clamp(y, 0, bgH-fgH)

	fgLines := lines(fg)
	bgLines := lines(bg)
	rev := a.reveal(t, fgW, fgH)
	_ = rev
	var sb strings.Builder
	sb.Grow(len(bg))

	for i, bgLine := range bgLines {
		if i > 0 {
			sb.WriteByte('\n')
		}

		if i < y || i >= y+fgH {
			sb.WriteString(bgLine)
			continue
		}

		sb.WriteString(revealRow(bgLine, fgLines[i-y], x, rev, i-y, fgW))
	}

	return sb.String()
}

// revealRow splices the visible part of one dialog row over the background.
func revealRow(bgLine, fgLine string, x int, rev box, row, fgW int) string {
	if row < rev.top || row >= rev.bottom {
		return bgLine
	}

	left := max(rev.left, 0)
	right := min(rev.right, fgW)

	if left >= right {
		return bgLine
	}

	mid := sliceCols(fgLine, left, right)
	midW := ansi.StringWidth(mid)
	start := x + left

	var sb strings.Builder

	sb.WriteString(keepLeft(bgLine, start))
	sb.WriteString(mid)
	sb.WriteString(keepRight(bgLine, start+midW))

	return sb.String()
}

// sliceCols returns visible cells [from, to) of s, preserving styles.
func sliceCols(s string, from, to int) string {
	if to <= from {
		return ""
	}

	return ansi.TruncateLeft(ansi.Truncate(s, to, ""), from, "")
}

// keepLeft returns the first n cells of s, padding if s is shorter.
func keepLeft(s string, n int) string {
	left := ansi.Truncate(s, n, "")

	if pad := n - ansi.StringWidth(left); pad > 0 {
		return left + whitespace(pad)
	}

	return left
}

// keepRight returns s from cell n onward, padding on the left when needed.
func keepRight(s string, n int) string {
	right := ansi.TruncateLeft(s, n, "")

	if pad := ansi.StringWidth(s) - ansi.StringWidth(right) - n; pad > 0 {
		return whitespace(pad) + right
	}

	return right
}

// offsets calculates the foreground position inside the background.
func offsets(
	fgW, fgH, bgW, bgH int,
	xPos, yPos DialogPosition,
	xOff, yOff int,
) (x, y int) {
	switch xPos {
	case Left:
		x = 0
	case Center:
		x = bgW/2 - fgW/2
	case Right:
		x = bgW - fgW
	}

	switch yPos {
	case Top:
		y = 0
	case Center:
		y = bgH/2 - fgH/2
	case Bottom:
		y = bgH - fgH
	}

	return x + xOff, y + yOff
}

// clamp limits v to [lower, upper], tolerating inverted bounds.
func clamp(v, lower, upper int) int {
	if lower > upper {
		lower, upper = upper, lower
	}

	return max(lower, min(upper, v))
}

// lines normalizes newlines and splits s.
func lines(s string) []string {
	return strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n")
}

func whitespace(n int) string {
	return strings.Repeat(" ", n)
}
