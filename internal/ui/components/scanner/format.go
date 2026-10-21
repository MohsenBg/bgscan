package scanner

import (
	"fmt"
	"time"
)

// formatCount renders a count with a K suffix from one thousand up.
func formatCount(n uint64) string {
	switch {
	case n < 1_000:
		return fmt.Sprintf("%d", n)
	case n < 1_000_000:
		return fmt.Sprintf("%.2fK", float64(n)/1_000)
	case n < 1_000_000_000:
		return fmt.Sprintf("%.2fM", float64(n)/1_000_000)
	case n < 1_000_000_000_000:
		return fmt.Sprintf("%.2fB", float64(n)/1_000_000_000)
	case n < 1_000_000_000_000_000:
		return fmt.Sprintf("%.2fT", float64(n)/1_000_000_000_000)
	case n < 1_000_000_000_000_000_000:
		return fmt.Sprintf("%.2fP", float64(n)/1_000_000_000_000_000)
	default:
		return fmt.Sprintf("%.2fE", float64(n)/1_000_000_000_000_000_000)
	}
}

func formatDuration(d time.Duration) string {
	if d <= 0 {
		return "0s"
	}
	const (
		minute = time.Minute
		hour   = time.Hour
		day    = 24 * time.Hour
		year   = 365*day + 6*time.Hour
	)
	switch {
	case d < minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < hour:
		m := d / minute
		s := (d % minute) / time.Second
		return fmt.Sprintf("%dm %ds", m, s)
	case d < day:
		h := d / hour
		m := (d % hour) / minute
		s := (d % minute) / time.Second
		return fmt.Sprintf("%dh %dm %ds", h, m, s)
	case d < year:
		days := d / day
		h := (d % day) / hour
		m := (d % hour) / minute
		return fmt.Sprintf("%dd %dh %dm", days, h, m)
	default:
		years := d / year
		days := (d % year) / day
		h := (d % day) / hour
		return fmt.Sprintf("%dy %dd %dh", years, days, h)
	}
}
