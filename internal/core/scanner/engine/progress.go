package engine

import (
	"math"
	"time"
)

// ErrorStats tracks probe failures categorized by failure mode.
type ErrorStats struct {
	Timeout     uint64
	Unreachable uint64
	Refused     uint64
	BadResponse uint64
	Closed      uint64
	Other       uint64
}

// Total returns the cumulative count of all errors.
func (e ErrorStats) Total() uint64 {
	return e.Timeout + e.Unreachable + e.Refused + e.BadResponse + e.Closed + e.Other
}

// Progress represents an instantaneous snapshot of execution metrics.
type Progress struct {
	Total      uint64
	Processed  uint64
	Succeed    uint64
	Errors     ErrorStats
	Percent    float64
	Elapsed    time.Duration
	RatePerSec float64
	ETA        time.Duration
}

// maxETASeconds avoids int64 overflow when casting to time.Duration.
const maxETASeconds = float64(math.MaxInt64) / float64(time.Second)

func reportProgress(
	start time.Time,
	paused time.Duration,
	total uint64,
	processed uint64,
	succeed uint64,
	errs ErrorStats,
	cb func(p Progress),
) {
	if cb == nil {
		return
	}

	now := time.Now()
	elapsed := max(now.Sub(start)-paused, 0)

	var rate float64
	if elapsed > 0 {
		rate = float64(processed) / elapsed.Seconds()
	}

	var percent float64
	if total > 0 {
		percent = (float64(processed) / float64(total)) * 100.0
		if percent > 100.0 {
			percent = 100.0
		}
	}

	var eta time.Duration
	if rate > 0 && processed < total {
		etaSeconds := float64(total-processed) / rate
		if etaSeconds < maxETASeconds {
			eta = time.Duration(etaSeconds * float64(time.Second))
		} else {
			eta = time.Duration(math.MaxInt64)
		}
	}

	p := Progress{
		Total:      total,
		Processed:  processed,
		Succeed:    succeed,
		Percent:    percent,
		Elapsed:    elapsed,
		RatePerSec: rate,
		Errors:     errs,
		ETA:        eta,
	}
	cb(p)
}
