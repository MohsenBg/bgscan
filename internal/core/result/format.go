package result

import (
	"fmt"
	"time"
)

// FormatDuration formats a duration with 2 decimal digits.
// Examples: "2.43ms", "1.34s", "500.00µs"
func FormatDuration(d time.Duration) string {
	us := d.Microseconds()
	if us >= 1_000_000 {
		return fmt.Sprintf("%.2fs", float64(us)/1_000_000)
	}
	if us >= 1000 {
		return fmt.Sprintf("%.2fms", float64(us)/1000)
	}
	return fmt.Sprintf("%dµs", us)
}
