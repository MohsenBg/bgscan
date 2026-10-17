package menu

import (
	"image/color"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/MohsenBg/bgscan/internal/ui/shared/anim"
)

const (
	rowDuration = 170 * time.Millisecond
	rowDelay    = 60 * time.Millisecond
)

type menuTickMsg time.Time

func (m *Model) rowDuration() time.Duration {
	if !m.env.Settings.Animated {
		return 0
	}
	return rowDuration
}

func (m *Model) rowDelay() time.Duration {
	if !m.env.Settings.Animated {
		return 0
	}
	return rowDelay
}

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
	rowDelay := m.rowDelay()
	rowDuration := m.rowDuration()
	entrance := rowDuration + time.Duration(len(m.items))*rowDelay
	return time.Since(m.enterAt) < entrance
}

// rowAlpha returns the entrance fade-in progress of row i.
func (m *Model) rowAlpha(i int) float64 {
	rowDelay := m.rowDelay()
	rowDuration := m.rowDuration()
	elapsed := time.Since(m.enterAt) - time.Duration(i)*rowDelay
	if elapsed <= 0 {
		return 0
	}

	return anim.EaseOutCubic(
		anim.Clamp01(float64(elapsed) / float64(rowDuration)),
	)
}

// fader blends a color in from the theme background by alpha.
func (m *Model) fader(alpha float64) func(color.Color) color.Color {
	bg := resolveTheme(&m.env.Theme).Background

	return func(c color.Color) color.Color {
		return anim.Mix(bg, c, alpha)
	}
}
