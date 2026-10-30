package app

import (
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/MohsenBg/bgscan/internal/logger"
	"github.com/MohsenBg/bgscan/internal/ui/shared/perf"
)

// frameStats is a temporary diagnostic for scroll lag. At Debug log level it
// writes one line per second, so the numbers show up in the UI log view:
//
//	[perf] keys=31/s frames=31/s view avg=3.1ms max=9.2ms
//
// Hold a scroll key, then read the line covering the moment you released it:
//
//   - keys/s far above the keyboard repeat rate means presses are queued and
//     the loop is behind. View is too slow to keep up, so presses arrive after
//     the key is up and the view keeps moving on its own.
//   - keys/s equal to frames/s and both around the repeat rate means nothing
//     is queued: input keeps up, and movement after release comes from
//     following new log lines instead of from the key.
//
// Update and View run on the same event loop goroutine, so no locking here.
type frameStats struct {
	since     time.Time
	keys      int
	frames    int
	viewTotal time.Duration
	viewMax   time.Duration
}

// perfTickMsg drives the once-per-second report.
type perfTickMsg struct{}

func perfTick() tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg { return perfTickMsg{} })
}

// recordKey counts a key press that reached Update.
func (s *frameStats) recordKey() {
	s.keys++
}

// recordFrame counts a rendered frame and how long it took.
func (s *frameStats) recordFrame(took time.Duration) {
	s.frames++
	s.viewTotal += took
	s.viewMax = max(s.viewMax, took)
}

// report resets the counters and writes one line for the window that ended.
func (s *frameStats) report(log *logger.Logger, now time.Time) {
	window := now.Sub(s.since)
	frames, keys, total, peak := s.frames, s.keys, s.viewTotal, s.viewMax

	s.since = now
	s.keys = 0
	s.frames = 0
	s.viewTotal = 0
	s.viewMax = 0

	// Logger.Debug already checks the level, so this costs nothing at Info.
	if log == nil || window < time.Second {
		return
	}

	avg := time.Duration(0)
	if frames > 0 {
		avg = total / time.Duration(frames)
	}

	log.Debug(
		"[perf] keys=%d/s frames=%d/s view avg=%s max=%s%s",
		int(float64(keys)/window.Seconds()),
		int(float64(frames)/window.Seconds()),
		avg.Round(10*time.Microsecond),
		peak.Round(10*time.Microsecond),
		perfReport(),
	)
}

// perfReport is the per-segment breakdown for the same window.
func perfReport() string {
	return perf.Report()
}
