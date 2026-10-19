// Package meter assigns quality tiers to latency and throughput metrics
// displayed in result tables.
package meter

import (
	"strconv"
	"strings"
	"time"

	"github.com/MohsenBg/bgscan/internal/core/result"
	"github.com/MohsenBg/bgscan/internal/core/speedtest"
)

// Tier represents a metric quality level.
type Tier uint8

const (
	// TierPlain leaves unrecognized cells at their default color.
	TierPlain Tier = iota
	// TierFail represents missing or invalid measurements.
	TierFail
	TierExcellent
	TierGood
	TierFair
	TierPoor
	// TierOver represents values beyond the configured quality bounds.
	TierOver
)

// Resolve returns the quality tier for a metric cell.
// Non-metric columns and unknown probes return TierPlain.
func Resolve(schema string, ctype result.ColumnType, cell string) Tier {
	schema = strings.ToLower(schema)

	switch ctype {
	case result.ColumnDuration:
		return resolveDuration(schema, cell)
	case result.ColumnBandwidth:
		return resolveBandwidth(schema, cell)
	default:
		return TierPlain
	}
}

func resolveDuration(schema, cell string) Tier {
	bands, ok := latency[strings.ToLower(schema)]
	if !ok {
		return TierPlain
	}

	d, ok := Duration(cell)
	if !ok {
		return TierFail
	}

	switch {
	case d <= bands[0]:
		return TierExcellent
	case d <= bands[1]:
		return TierGood
	case d <= bands[2]:
		return TierFair
	case d <= bands[3]:
		return TierPoor
	default:
		return TierOver
	}
}

func resolveBandwidth(schema, cell string) Tier {
	bands, ok := bandwidth[strings.ToLower(schema)]
	if !ok {
		return TierPlain
	}

	v, err := speedtest.ParseBitsPerSec(cell)
	if err != nil || v == 0 {
		return TierFail
	}

	switch {
	case v >= bands[0]:
		return TierExcellent
	case v >= bands[1]:
		return TierGood
	case v >= bands[2]:
		return TierFair
	case v >= bands[3]:
		return TierPoor
	default:
		return TierOver
	}
}

// latency defines inclusive upper bounds for each quality tier:
// {excellent, good, fair, poor}. Values are tuned for real-world,
// often congested or filtered networks rather than ideal links.
var latency = map[string][4]time.Duration{
	"icmp":        {60 * time.Millisecond, 120 * time.Millisecond, 250 * time.Millisecond, 600 * time.Millisecond},
	"tcp":         {100 * time.Millisecond, 200 * time.Millisecond, 400 * time.Millisecond, 1000 * time.Millisecond},
	"dnsresolver": {100 * time.Millisecond, 200 * time.Millisecond, 500 * time.Millisecond, 1200 * time.Millisecond},
	"http":        {500 * time.Millisecond, 1200 * time.Millisecond, 2500 * time.Millisecond, 5000 * time.Millisecond},
	"vaydns":      {2000 * time.Millisecond, 4000 * time.Millisecond, 6000 * time.Millisecond, 8000 * time.Millisecond},
	"dnstt":       {2000 * time.Millisecond, 4000 * time.Millisecond, 6000 * time.Millisecond, 8000 * time.Millisecond},
	"stormdns":    {2000 * time.Millisecond, 4000 * time.Millisecond, 6000 * time.Millisecond, 8000 * time.Millisecond},
	"masterdns":   {2000 * time.Millisecond, 4000 * time.Millisecond, 6000 * time.Millisecond, 8000 * time.Millisecond},
	"slipstream":  {2000 * time.Millisecond, 4000 * time.Millisecond, 6000 * time.Millisecond, 8000 * time.Millisecond},
	"thefeed":     {400 * time.Millisecond, 1000 * time.Millisecond, 2000 * time.Millisecond, 4500 * time.Millisecond},
	"xray":        {1000 * time.Millisecond, 2500 * time.Millisecond, 5000 * time.Millisecond, 9000 * time.Millisecond},
}

// bandwidth defines descending minimum throughput thresholds:
// {excellent, good, fair, poor}.
var bandwidth = map[string][4]speedtest.BitsPerSec{
	// Download and upload through the proxied connection.
	"xray": {10 * speedtest.Mbps, 4 * speedtest.Mbps, 1 * speedtest.Mbps, 250 * speedtest.Kbps},
}

// Duration parses the duration format produced by result.FormatDuration,
func Duration(cell string) (time.Duration, bool) {
	cell = strings.TrimSpace(cell)
	if cell == "" || cell == "-" {
		return 0, false
	}

	var unit time.Duration

	switch {
	case strings.HasSuffix(cell, "µs"):
		unit = time.Microsecond
		cell = strings.TrimSuffix(cell, "µs")
	case strings.HasSuffix(cell, "ms"):
		unit = time.Millisecond
		cell = strings.TrimSuffix(cell, "ms")
	case strings.HasSuffix(cell, "s"):
		unit = time.Second
		cell = strings.TrimSuffix(cell, "s")
	default:
		return 0, false
	}

	value, err := strconv.ParseFloat(strings.TrimSpace(cell), 64)
	if err != nil {
		return 0, false
	}

	return time.Duration(value * float64(unit)), true
}
