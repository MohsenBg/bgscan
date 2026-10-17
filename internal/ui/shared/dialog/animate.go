// Package dialog provides dialog frames and open/close transitions.
package dialog

import (
	"math"
	"time"

	"github.com/MohsenBg/bgscan/internal/ui/shared/anim"
)

// Duration is the dialog open/close transition time.
const Duration = 120 * time.Millisecond

// FrameInterval is the transition tick interval.
const FrameInterval = anim.FrameInterval

// Animation is a dialog reveal transition.
type Animation uint8

// Dialog transitions.
const (
	AnimNone Animation = iota
	AnimExpand
	AnimDropIn
	AnimRiseUp
	AnimWipe
	AnimSplit
	AnimCurtain
)

// All returns animations in picker order.
func All() []Animation {
	return []Animation{
		AnimNone,
		AnimExpand,
		AnimDropIn,
		AnimRiseUp,
		AnimWipe,
		AnimSplit,
		AnimCurtain,
	}
}

// String returns the picker label.
func (a Animation) String() string {
	switch a {
	case AnimExpand:
		return "Expand"
	case AnimDropIn:
		return "Drop in"
	case AnimRiseUp:
		return "Rise up"
	case AnimWipe:
		return "Wipe"
	case AnimSplit:
		return "Split"
	case AnimCurtain:
		return "Curtain"
	default:
		return "None (instant)"
	}
}

// box is the visible rectangle of a dialog.
type box struct {
	top, left, bottom, right int
}

// reveal returns the visible rectangle at progress t.
//
// t is expected to move from 0 to 1 while opening, and from 1 to 0 while
// closing.
func (a Animation) reveal(t float64, w, h int) box {
	p := anim.EaseInOutCubic(t)

	switch a {
	case AnimExpand:
		return centered(cells(p, h), cells(p, w), w, h)

	case AnimDropIn:
		return box{
			bottom: cells(p, h),
			right:  w,
		}

	case AnimRiseUp:
		rows := cells(p, h)
		return box{
			top:    h - rows,
			bottom: h,
			right:  w,
		}

	case AnimWipe:
		return box{
			bottom: h,
			right:  cells(p, w),
		}

	case AnimSplit:
		cols := cells(p, w)
		left := (w - cols) / 2
		return box{
			bottom: h,
			left:   left,
			right:  left + cols,
		}

	case AnimCurtain:
		return centered(cells(p, h), w, w, h)

	default:
		return box{
			bottom: h,
			right:  w,
		}
	}
}

// centered places a rows by cols box inside a w by h area.
func centered(rows, cols, w, h int) box {
	top := (h - rows) / 2
	left := (w - cols) / 2

	return box{
		top:    top,
		left:   left,
		bottom: top + rows,
		right:  left + cols,
	}
}

// cells returns how many of n cells are visible at eased progress p.
//
// The final cell is held until p reaches 1, except when n == 1.
func cells(p float64, n int) int {
	switch {
	case n <= 0 || p <= 0:
		return 0
	case p >= 1:
		return n
	default:
		return max(1, min(n-1, int(math.Ceil(p*float64(n)))))
	}
}
