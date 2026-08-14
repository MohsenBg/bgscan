package speedtest

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Bits is a raw bit count. Use the constructor or arithmetic on BitsPerSec.
type Bits uint64

// BitsPerSec represents a network throughput measurement in bits per second.
type BitsPerSec uint64

const (
	Bps  BitsPerSec = 1
	Kbps BitsPerSec = 1_000
	Mbps BitsPerSec = 1_000_000
	Gbps BitsPerSec = 1_000_000_000
)

// String returns a human-readable representation, e.g. "23.45 Mbps".
func (s BitsPerSec) String() string {
	switch {
	case s >= Gbps:
		return fmt.Sprintf("%.2f Gbps", float64(s)/float64(Gbps))
	case s >= Mbps:
		return fmt.Sprintf("%.2f Mbps", float64(s)/float64(Mbps))
	case s >= Kbps:
		return fmt.Sprintf("%.2f Kbps", float64(s)/float64(Kbps))
	default:
		return fmt.Sprintf("%d bps", s)
	}
}

// Bps returns the raw bits-per-second value as uint64.
func (s BitsPerSec) Bps() uint64 {
	return uint64(s)
}

// SpeedResult holds the outcome of a single speed measurement.
type SpeedResult struct {
	Speed    BitsPerSec
	Bytes    uint64
	MinSpeed BitsPerSec // minimum required; zero means no minimum
}

// String delegates to Speed.String().
func (r SpeedResult) String() string {
	return r.Speed.String()
}

// BelowMinimum reports whether the measured speed is below the configured
// minimum. Always false when MinSpeed is zero.
func (r SpeedResult) BelowMinimum() bool {
	return r.MinSpeed > 0 && r.Speed < r.MinSpeed
}

// bitsPerSec converts bytes transferred over a duration in seconds to BitsPerSec.
// Returns 0 for non-positive inputs rather than panicking.
func bitsPerSec(bytes uint64, seconds float64) BitsPerSec {
	if bytes == 0 || seconds <= 0 {
		return 0
	}
	raw := float64(bytes) * 8 / seconds
	if raw > math.MaxUint64 {
		return BitsPerSec(math.MaxUint64)
	}
	return BitsPerSec(uint64(raw))
}

// ParseBitsPerSec parses strings like:
//
//	"123 bps"
//	"1 Kbps"
//	"12.5 Mbps"
//	"2 Gbps"
func ParseBitsPerSec(s string) (BitsPerSec, error) {
	s = strings.TrimSpace(strings.ToLower(s))

	var multiplier uint64
	var isBps bool

	switch {
	case strings.HasSuffix(s, "gbps"):
		multiplier = uint64(Gbps)
		s = strings.TrimSpace(strings.TrimSuffix(s, "gbps"))

	case strings.HasSuffix(s, "mbps"):
		multiplier = uint64(Mbps)
		s = strings.TrimSpace(strings.TrimSuffix(s, "mbps"))

	case strings.HasSuffix(s, "kbps"):
		multiplier = uint64(Kbps)
		s = strings.TrimSpace(strings.TrimSuffix(s, "kbps"))

	case strings.HasSuffix(s, "bps"):
		multiplier = uint64(Bps)
		isBps = true
		s = strings.TrimSpace(strings.TrimSuffix(s, "bps"))

	default:
		return 0, fmt.Errorf("unknown speed unit")
	}

	if s == "" {
		return 0, fmt.Errorf("missing speed value")
	}

	if isBps {
		value, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("speed exceeds uint64 range")
		}

		return BitsPerSec(value), nil
	}

	value, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid speed value")
	}

	if value < 0 || math.IsInf(value, 0) {
		return 0, fmt.Errorf("invalid speed value")
	}

	result := value * float64(multiplier)

	if result > float64(math.MaxUint64) {
		return 0, fmt.Errorf("speed exceeds uint64 range")
	}

	return BitsPerSec(result), nil
}
