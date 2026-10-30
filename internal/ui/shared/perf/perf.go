// Package perf is a temporary diagnostic. It measures named segments of a
// render and reports how much time each one took.
//
// Everything here runs on the Bubble Tea event loop goroutine, so the tallies
// need no locking. Delete this package once the render is fast enough.
package perf

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type stat struct {
	total time.Duration
	peak  time.Duration
	count int
}

var stats = map[string]*stat{}

// Begin starts measuring a segment. Call the returned function to stop.
//
//	defer perf.Begin("body")()
func Begin(name string) func() {
	start := time.Now()
	return func() { Record(name, time.Since(start)) }
}

// Record adds one measurement.
func Record(name string, took time.Duration) {
	s := stats[name]
	if s == nil {
		s = &stat{}
		stats[name] = s
	}

	s.total += took
	s.count++
	s.peak = max(s.peak, took)
}

// Report formats the averages, slowest total first, and starts a new window.
func Report() string {
	if len(stats) == 0 {
		return ""
	}

	names := make([]string, 0, len(stats))
	for name := range stats {
		names = append(names, name)
	}
	sort.Slice(names, func(i, j int) bool { return stats[names[i]].total > stats[names[j]].total })

	var b strings.Builder
	for _, name := range names {
		s := stats[name]
		avg := s.total / time.Duration(s.count)

		fmt.Fprintf(&b, " %s=%.2fms(max %.2f)/%d",
			name,
			float64(avg.Microseconds())/1000,
			float64(s.peak.Microseconds())/1000,
			s.count,
		)
	}

	stats = map[string]*stat{}
	return b.String()
}
