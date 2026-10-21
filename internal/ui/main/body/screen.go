package body

import (
	"math"
	"strings"
	"time"

	"github.com/MohsenBg/bgscan/internal/ui/shared/anim"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
	"github.com/charmbracelet/x/ansi"

	tea "charm.land/bubbletea/v2"
)

const swapDuration = 150 * time.Millisecond

// Keep screen transitions body-local: body swaps render full-region pages,
// while dialog transitions compose positioned overlays.
type screen struct {
	component ui.Component
	closingAt time.Time
}

func (s screen) closing() bool {
	return !s.closingAt.IsZero()
}

func (s screen) settled(now time.Time) bool {
	return !s.closing() || now.Sub(s.closingAt) >= swapDuration
}

type swapKind int

const (
	pushOpen swapKind = iota
	pushClose
)

// swap holds the screens participating in a transition.
type swap struct {
	leaving  ui.Component
	entering ui.Component
	kind     swapKind
	at       time.Time
}

func (s *swap) progress(now time.Time) float64 {
	return anim.EaseOutCubic(anim.Progress(s.at, swapDuration))
}

func (s *swap) settled(now time.Time) bool {
	return s.progress(now) >= 1
}

type bodyFrameMsg time.Time

func tickSwaps() tea.Cmd {
	return tea.Tick(anim.FrameInterval, func(t time.Time) tea.Msg {
		return bodyFrameMsg(t)
	})
}

// swapView renders the active transition within the body viewport.
func (m *Model) swapView() string {
	body := m.session.Env().Layout.Body()

	if m.swap.progress(time.Now()) >= 1 {
		return ""
	}

	style := containerStyle(body.Width, body.Height)
	leaving := style.Render(m.swap.leaving.View())
	entering := style.Render(m.swap.entering.View())

	return composeSwap(leaving, entering, body.Width, m.swap.kind, m.swap.progress(time.Now()))
}

// composeSwap slides both screens through the viewport as a single page.
func composeSwap(leaving, entering string, width int, kind swapKind, progress float64) string {
	var (
		left, right []string
		offset      int
	)

	switch kind {
	case pushOpen:
		left, right = strings.Split(leaving, "\n"), strings.Split(entering, "\n")
		offset = int(math.Round(progress * float64(width)))
	case pushClose:
		left, right = strings.Split(entering, "\n"), strings.Split(leaving, "\n")
		offset = int(math.Round((1 - progress) * float64(width)))
	}

	var out strings.Builder
	out.Grow(len(leaving) + len(entering))

	for i := 0; i < max(len(left), len(right)); i++ {
		if i > 0 {
			out.WriteByte('\n')
		}

		row := ansi.Cut(line(left, i)+line(right, i), offset, offset+width)
		out.WriteString(padTo(row, width))
	}

	return out.String()
}

func line(lines []string, i int) string {
	if i >= len(lines) {
		return ""
	}

	return lines[i]
}

// padTo pads a row to the viewport width.
func padTo(s string, width int) string {
	if padding := width - ansi.StringWidth(s); padding > 0 {
		return s + strings.Repeat(" ", padding)
	}

	return s
}
