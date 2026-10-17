package tabs

import (
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/MohsenBg/bgscan/internal/ui/shared/anim"
)

const (
	slideDuration = 220 * time.Millisecond
	chevronSep    = " › "
)

type frameMsg time.Time

func (m *Model[T]) tick() tea.Cmd {
	if !m.animated || len(m.tabs) == 0 {
		return nil
	}
	if anim.Progress(m.slideStart, slideDuration) >= 1 {
		return nil
	}
	return tea.Tick(anim.FrameInterval, func(t time.Time) tea.Msg {
		return frameMsg(t)
	})
}

func (m *Model[T]) step() {
	p := anim.Progress(m.slideStart, slideDuration)
	p = anim.EaseInOutCubic(p)
	m.slide = m.slideFrom + (m.slideTo-m.slideFrom)*p
}

func (m *Model[T]) startSlide() {
	if !m.animated {
		return
	}

	m.slideFrom = m.slide
	m.slideTo = float64(m.tabOffset(m.idx))
	m.slideStart = time.Now()
}
