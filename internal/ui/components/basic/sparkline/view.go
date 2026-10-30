package sparkline

import (
	"math"
	"strings"

	"charm.land/lipgloss/v2"
)

const brailleBase rune = 0x2800

// Dot patterns filled from bottom to top.
var (
	leftFill  = [5]rune{0x00, 0x40, 0x44, 0x46, 0x47}
	rightFill = [5]rune{0x00, 0x80, 0xA0, 0xB0, 0xB8}
)

// View renders a fixed-width sparkline, aligned to the right.
// The result is memoised until a mutation marks the model dirty.
func (m *Model) View() string {
	if m.dirty {
		m.cache = m.render()
		m.dirty = false
	}
	return m.cache
}

func (m *Model) render() string {
	styles := m.getStyles()
	lo, hi := m.bounds()

	heights := make([]int, m.capacity())
	offset := len(heights) - len(m.samples)
	for i, v := range m.samples {
		heights[offset+i] = m.height(v, lo, hi)
	}

	var out strings.Builder
	out.Grow(m.width * 4)
	for c := 0; c < m.width; c++ {
		glyph := brailleBase + leftFill[heights[2*c]] + rightFill[heights[2*c+1]]
		out.WriteString(styles[c].Render(string(glyph)))
	}
	return out.String()
}

// styles returns the cached gradient, rebuilding it only when the width changes.
func (m *Model) getStyles() []lipgloss.Style {
	if len(m.styles) != m.width {
		m.styles = gradientStyles(&m.env.Theme, m.width)
	}
	return m.styles
}

// height maps a sample to 1..4 dots. NaN represents missing data.
func (m *Model) height(v, lo, hi float64) int {
	if math.IsNaN(v) {
		return 0
	}
	if hi <= lo {
		if hi <= 0 {
			return 1
		}
		return 2
	}

	ratio := math.Max(0, math.Min(1, (v-lo)/(hi-lo)))
	return 1 + int(math.Round(ratio*3))
}

// bounds returns the range used to scale samples.
func (m *Model) bounds() (lo, hi float64) {
	lo, hi = math.Inf(1), math.Inf(-1)
	for _, v := range m.samples {
		if math.IsNaN(v) {
			continue
		}
		lo = math.Min(lo, v)
		hi = math.Max(hi, v)
	}
	if math.IsInf(hi, -1) {
		return 0, 0
	}

	minSpan := hi * m.minSpan
	if hi-lo < minSpan {
		mid := (hi + lo) / 2
		lo, hi = mid-minSpan/2, mid+minSpan/2
	}

	return math.Max(lo, 0), hi
}
