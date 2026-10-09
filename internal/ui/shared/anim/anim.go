// Package anim provides timing, easing, and color helpers for animated TUI components.
package anim

import (
	"image/color"
	"time"
)

// Animation timing constants.
const (
	FrameInterval = 16 * time.Millisecond
)

// Clamp01 constrains t to the range [0, 1].
func Clamp01(t float64) float64 {
	switch {
	case t < 0:
		return 0
	case t > 1:
		return 1
	default:
		return t
	}
}

// EaseOutCubic applies a cubic ease-out curve.
func EaseOutCubic(t float64) float64 {
	u := 1 - Clamp01(t)
	return 1 - u*u*u
}

// EaseInOutCubic applies a cubic ease-in-out curve.
func EaseInOutCubic(t float64) float64 {
	t = Clamp01(t)

	if t < 0.5 {
		return 4 * t * t * t
	}

	u := -2*t + 2
	return 1 - u*u*u/2
}

// Progress returns normalized elapsed progress for an animation.
func Progress(start time.Time, d time.Duration) float64 {
	if start.IsZero() || d <= 0 {
		return 1
	}

	return Clamp01(float64(time.Since(start)) / float64(d))
}

// Mix linearly blends a into b by t.
//
// Nil colors are treated as transparent, except at t == 0 or t == 1,
// where the original endpoint color is returned.
func Mix(a, b color.Color, t float64) color.Color {
	t = Clamp01(t)

	if t == 0 {
		return a
	}

	if t == 1 {
		return b
	}

	ar, ag, ab, aa := nrgba(a)
	br, bg, bb, ba := nrgba(b)

	return color.NRGBA{
		R: uint8(float64(ar)*(1-t) + float64(br)*t + 0.5),
		G: uint8(float64(ag)*(1-t) + float64(bg)*t + 0.5),
		B: uint8(float64(ab)*(1-t) + float64(bb)*t + 0.5),
		A: uint8(float64(aa)*(1-t) + float64(ba)*t + 0.5),
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
