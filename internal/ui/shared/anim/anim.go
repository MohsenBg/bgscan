// Package anim provides timing, easing, and color helpers for animated TUI components.
package anim

import (
	"image/color"
	"time"
)

const (
	FrameInterval = 16 * time.Millisecond
	FadeDuration  = 300 * time.Millisecond
	RowDuration   = 170 * time.Millisecond
	RowStagger    = 60 * time.Millisecond
	FlashDuration = 130 * time.Millisecond
)

// Clamp01 limits t to [0, 1].
func Clamp01(t float64) float64 {
	if t < 0 {
		return 0
	}
	if t > 1 {
		return 1
	}
	return t
}

// EaseOutCubic starts fast and slows down toward the end.
func EaseOutCubic(t float64) float64 {
	l := 1 - Clamp01(t)
	return 1 - l*l*l
}

// Progress returns the progress of an animation.
func Progress(start time.Time, d time.Duration) float64 {
	if start.IsZero() || d <= 0 {
		return 1
	}
	return Clamp01(float64(time.Since(start)) / float64(d))
}

// Mix blends two colors.
func Mix(a, b color.Color, t float64) color.Color {
	t = Clamp01(t)
	if t <= 0 || a == nil {
		return a
	}
	if t >= 1 || b == nil {
		return b
	}

	ar, ag, ab, aa := nrgba(a)
	br, bg, bb, ba := nrgba(b)
	return color.NRGBA{
		R: uint8(float64(ar)*(1-t) + float64(br)*t),
		G: uint8(float64(ag)*(1-t) + float64(bg)*t),
		B: uint8(float64(ab)*(1-t) + float64(bb)*t),
		A: uint8(float64(aa)*(1-t) + float64(ba)*t),
	}
}

func nrgba(c color.Color) (r, g, b, a uint8) {
	if c == nil {
		return 0, 0, 0, 0
	}

	R, G, B, A := c.RGBA()
	if A == 0 {
		return 0, 0, 0, 0
	}
	if A < 0xffff {
		R = uint32(uint64(R) * 0xffff / uint64(A))
		G = uint32(uint64(G) * 0xffff / uint64(A))
		B = uint32(uint64(B) * 0xffff / uint64(A))
	}

	return uint8(R >> 8), uint8(G >> 8), uint8(B >> 8), uint8(A >> 8)
}
