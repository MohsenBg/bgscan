package menu

import (
	"image/color"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/MohsenBg/bgscan/internal/ui/shared/anim"
)

type menuTickMsg time.Time

// nextFrame schedules the next animation frame.
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

// rowAlpha returns the entrance fade-in progress of row i.
func (m *Model) rowAlpha(i int) float64 {
	elapsed := time.Since(m.enterAt) - time.Duration(i)*anim.RowStagger
	if elapsed <= 0 {
		return 0
	}

	return anim.EaseOutCubic(
		anim.Clamp01(float64(elapsed) / float64(anim.RowDuration)),
	)
}

// fader blends a color in from the theme background by alpha.
func (m *Model) fader(alpha float64) func(color.Color) color.Color {
	bg := resolveTheme(m.theme).Background

	return func(c color.Color) color.Color {
		return anim.Mix(bg, c, alpha)
	}
}
