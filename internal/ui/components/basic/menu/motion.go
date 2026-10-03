package menu

import (
	"image/color"
	"time"

	"github.com/MohsenBg/bgscan/internal/ui/shared/anim"

	tea "charm.land/bubbletea/v2"
)

type menuTickMsg time.Time

// nextFrame schedules the next animation frame, or nothing once the
// entrance animation is done.
func (m *Model) nextFrame() tea.Cmd {
	if !m.animating() {
		return nil
	}
	return tea.Tick(anim.FrameInterval, func(t time.Time) tea.Msg {
		return menuTickMsg(t)
	})
}

func (m *Model) animating() bool {
	entrance := anim.RowDuration + time.Duration(len(m.items))*anim.RowStagger
	return time.Since(m.enterAt) < entrance
}

// rowAlpha is the entrance fade-in progress (0..1) of row i.
func (m *Model) rowAlpha(i int) float64 {
	elapsed := time.Since(m.enterAt) - time.Duration(i)*anim.RowStagger
	if elapsed <= 0 {
		return 0
	}
	return anim.EaseOutCubic(anim.Clamp01(float64(elapsed) / float64(anim.RowDuration)))
}

// fader returns a function that blends a color in from the theme
// background by alpha.
func (m *Model) fader(alpha float64) func(color.Color) color.Color {
	bg := resolveTheme(m.theme).Background
	return func(c color.Color) color.Color { return anim.Mix(bg, c, alpha) }
}
