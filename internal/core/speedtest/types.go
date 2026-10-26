package speedtest

import (
	"context"
	"fmt"
	"math"
	"net"
	"strconv"
	"strings"
	"time"
)

// Unit multipliers, in bits per second.
const (
	Bps  BitsPerSec = 1
	Kbps BitsPerSec = 1e3
	Mbps BitsPerSec = 1e6
	Gbps BitsPerSec = 1e9
)

// BitsPerSec represents a network throughput measurement in bits per second.
type BitsPerSec uint64

// Uint64 returns the raw bits-per-second value.
func (b BitsPerSec) Uint64() uint64 {
	return uint64(b)
}

func (b BitsPerSec) String() string {
	switch {
	case b >= Gbps:
		return fmt.Sprintf("%.2f Gbps", float64(b)/float64(Gbps))
	case b >= Mbps:
		return fmt.Sprintf("%.2f Mbps", float64(b)/float64(Mbps))
	case b >= Kbps:
		return fmt.Sprintf("%.2f Kbps", float64(b)/float64(Kbps))
	default:
		return fmt.Sprintf("%d bps", uint64(b))
	}
}

// SpeedResult holds the outcome of a single speed measurement.
type SpeedResult struct {
	Speed    BitsPerSec
	Bytes    uint64
	MinSpeed BitsPerSec
}

func (r SpeedResult) String() string {
	return r.Speed.String()
}

// BelowMinimum reports whether the measured speed is below the configured
// minimum. Always false when MinSpeed is zero.
func (r SpeedResult) BelowMinimum() bool {
	return r.MinSpeed > 0 && r.Speed < r.MinSpeed
}

// bitsPerSec converts a byte count transferred over a duration to BitsPerSec.
// Returns 0 for non-positive inputs rather than panicking.
func bitsPerSec(bytes uint64, seconds float64) BitsPerSec {
	if bytes == 0 || seconds <= 0 {
		return 0
	}
	raw := float64(bytes) * 8 / seconds
	if raw >= math.MaxUint64 {
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
//
// Units are case-insensitive and optional surrounding whitespace is ignored.
// Fractional bps truncates toward zero.
func ParseBitsPerSec(s string) (BitsPerSec, error) {
	s = strings.ToLower(strings.TrimSpace(s))

	// Order matters: longer suffixes must be checked before shorter ones.
	var multiplier BitsPerSec
	switch {
	case strings.HasSuffix(s, "gbps"):
		multiplier = Gbps
		s = strings.TrimSuffix(s, "gbps")
	case strings.HasSuffix(s, "mbps"):
		multiplier = Mbps
		s = strings.TrimSuffix(s, "mbps")
	case strings.HasSuffix(s, "kbps"):
		multiplier = Kbps
		s = strings.TrimSuffix(s, "kbps")
	case strings.HasSuffix(s, "bps"):
		multiplier = Bps
		s = strings.TrimSuffix(s, "bps")
	default:
		return 0, fmt.Errorf("speedtest: unknown speed unit")
	}

	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("speedtest: missing speed value")
	}

	// Parse whole bps values as integers first so the full uint64 range
	// (including math.MaxUint64) survives without float64 rounding.
	if multiplier == Bps {
		if value, err := strconv.ParseUint(s, 10, 64); err == nil {
			return BitsPerSec(value), nil
		}
	}

	value, err := strconv.ParseFloat(s, 64)
	if err != nil || value < 0 || math.IsInf(value, 0) || math.IsNaN(value) {
		return 0, fmt.Errorf("speedtest: invalid speed value %q", s)
	}

	result := value * float64(multiplier)
	if result >= float64(math.MaxUint64) {
		return 0, fmt.Errorf("speedtest: speed exceeds uint64 range")
	}

	return BitsPerSec(result), nil
}

// DownloadConfig controls a single download measurement.
type DownloadConfig struct {
	// Bytes is the number of bytes to request from the test endpoint. Must be > 0.
	Bytes int64

	// Timeout is the max time for the body transfer. Connection setup excluded.
	// Zero or negative means no transfer timeout.
	Timeout time.Duration

	// MinSpeed causes MeasureDownloadSpeed to error if throughput falls below it.
	MinSpeed BitsPerSec

	// ProxyPort is the local SOCKS5 proxy port used when DialContext is nil.
	ProxyPort uint16

	// DialContext provides the connection used to reach the test endpoint.
	// When set, ProxyPort is ignored.
	DialContext func(
		ctx context.Context,
		network string,
		address string,
	) (net.Conn, error)
}

// LatencyConfig controls a single latency measurement.
type LatencyConfig struct {
	// URL is the HTTP endpoint used for latency measurement.
	// Empty uses GoogleGenerate204HTTP.
	URL string

	// Timeout is the maximum time allowed for the round trip.
	// Zero (or negative) means no timeout is applied.
	Timeout time.Duration

	// MaxLatency causes MeasureLatency to return an error when the
	// measured RTT exceeds this threshold.
	MaxLatency time.Duration

	// ProxyPort is the local SOCKS5 proxy port used when DialContext is nil.
	ProxyPort uint16

	// DialContext provides the connection used to reach the test endpoint.
	// When set, ProxyPort is ignored.
	DialContext func(
		ctx context.Context,
		network string,
		address string,
	) (net.Conn, error)
}

type LatencyResult struct {
	RTT        time.Duration
	MaxLatency time.Duration
}

func (r LatencyResult) String() string {
	return r.RTT.String()
}

func (r LatencyResult) AboveMaximum() bool {
	return r.MaxLatency > 0 && r.RTT > r.MaxLatency
}

// UploadConfig controls a single upload measurement.
type UploadConfig struct {
	// Bytes is the number of bytes to upload. Must be > 0.
	Bytes int64

	// Timeout is the maximum time allowed for the transfer.
	Timeout time.Duration

	// MinSpeed causes MeasureUploadSpeed to return an error when the
	MinSpeed BitsPerSec

	// ProxyPort is the local SOCKS5 proxy port used when DialContext is nil.
	ProxyPort uint16

	// DialContext provides the connection used to reach the test endpoint.
	// When set, ProxyPort is ignored.
	DialContext func(
		ctx context.Context,
		network string,
		address string,
	) (net.Conn, error)
}
