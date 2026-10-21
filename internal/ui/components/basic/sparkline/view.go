package sparkline

import (
	"math"
	"strings"
)

const brailleBase rune = 0x2800

// Braille dots filled from the bottom up, for 0..4 dots in one column.
// Left column dots are 7,3,2,1 and right column dots are 8,6,5,4.
var (
	leftFill  = [5]rune{0x00, 0x40, 0x44, 0x46, 0x47}
	rightFill = [5]rune{0x00, 0x80, 0xA0, 0xB0, 0xB8}
)

// View renders the sparkline, always exactly `width` cells wide.
// Until enough samples have arrived, the left side stays blank so the chart
// grows in from the right edge.
func (m *Model) View() string {
	th := m.env.Theme
	lo, hi := m.bounds()
	styles := gradientStyles(&th, m.width)

	// One height (0..4) per dot column, newest on the right. 0 means no data yet.
	heights := make([]int, m.capacity())
	offset := len(heights) - len(m.samples)
	for i, v := range m.samples {
		heights[offset+i] = m.height(v, lo, hi)
	}

	var out strings.Builder
	for c := 0; c < m.width; c++ {
		glyph := brailleBase + leftFill[heights[2*c]] + rightFill[heights[2*c+1]]
		out.WriteString(styles[c].Render(string(glyph)))
	}

	return out.String()
}

// height maps a sample to 1..4 dots. Even a zero keeps one baseline dot, so
// "stopped" looks different from "no data yet" (0 dots).
func (m *Model) height(v, lo, hi float64) int {
	if math.IsNaN(v) {
		return 0
	}
	if hi <= lo {
		if hi <= 0 {
			return 1 // everything is zero
		}
		return 2 // flat data: draw it in the middle
	}

	ratio := math.Max(0, math.Min(1, (v-lo)/(hi-lo)))
	return 1 + int(math.Round(ratio*3))
}

// bounds returns the values that map to the lowest and highest bar.
func (m *Model) bounds() (lo, hi float64) {
	lo, hi = math.Inf(1), math.Inf(-1)
	for _, v := range m.samples {
		if math.IsNaN(v) {
			continue
		}
		lo = math.Min(lo, v)
		hi = math.Max(hi, v)
	}
	if math.IsInf(hi, -1) { // no usable samples
		return 0, 0
	}

	// Never zoom in closer than minSpan of the max, so tiny wobble stays tiny.
	// If the data is flatter than that, center the range on the middle value.
	minSpan := hi * m.minSpan
	if hi-lo < minSpan {
		mid := (hi + lo) / 2
		lo, hi = mid-minSpan/2, mid+minSpan/2
	}

	return math.Max(lo, 0), hi
}
