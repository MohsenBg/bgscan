package scanner

import (
	"cmp"
	"slices"
	"sync"
	"time"

	"github.com/MohsenBg/bgscan/internal/core/result"
	"github.com/MohsenBg/bgscan/internal/core/scanner"
	"github.com/MohsenBg/bgscan/internal/core/scanner/engine"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/progress"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/sparkline"
	"github.com/MohsenBg/bgscan/internal/ui/components/tables/ipviewer"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

// StageStatus tracks the lifecycle of a single scan stage.
type StageStatus int

const (
	StatusWaiting StageStatus = iota + 1
	StatusScanning
	StatusEnded
	StatusError
)

// String returns a human-readable representation of the stage status.
func (s StageStatus) String() string {
	switch s {
	case StatusWaiting:
		return "Waiting"
	case StatusScanning:
		return "Scanning"
	case StatusEnded:
		return "Ended"
	case StatusError:
		return "Error"
	default:
		return "Unknown"
	}
}

type scanStage struct {
	id     int
	cfg    scanner.StageConfig
	maxIPs int

	table       ui.Component
	progressBar ui.Component

	// results is owned by the UI goroutine only.
	results []result.Result

	// batch is appended to by worker goroutines (engine hooks) and drained
	// by the UI goroutine in mergeBatch. Guarded by batchMu.
	batchMu sync.Mutex
	batch   []result.Result

	status   StageStatus
	progress engine.Progress
	program  ui.Program

	lastSparkUpdated time.Time
	sparkline        *sparkline.Model
}

// newScanStage initializes a new scan stage with safe capacity limits.
func newScanStage(env *ui.Env, program ui.Program, idx int, cfg scanner.StageConfig) *scanStage {
	maxIPs := int(env.Settings.MaxRow)
	if maxIPs < 0 {
		maxIPs = ui.DefaultMaxRow
	}

	return &scanStage{
		id:               idx,
		cfg:              cfg,
		maxIPs:           maxIPs,
		table:            createIPViewer(env, cfg.Probe.Schema()),
		progressBar:      progress.New(env),
		results:          make([]result.Result, 0, maxIPs),
		batch:            make([]result.Result, 0, 128),
		status:           StatusWaiting,
		sparkline:        sparkline.New(env),
		lastSparkUpdated: time.Now().Add(-1 * time.Second),
		program:          program,
	}
}

// ipViewer safely casts the generic table component to the specific IP viewer model.
func (s *scanStage) ipViewer() *ipviewer.Model {
	if viewer, ok := s.table.(*ipviewer.Model); ok {
		return viewer
	}
	return nil
}

func (s *scanStage) progressModel() *progress.Model {
	if progressBar, ok := s.progressBar.(*progress.Model); ok {
		return progressBar
	}
	return nil
}

// hooks returns the engine callbacks that feed this stage's state.
func (s *scanStage) hooks() engine.ScanHooks {
	return engine.ScanHooks{
		OnError:    s.onEngineError,
		OnProgress: s.onEngineProgress,
		OnSuccess:  s.onEngineSuccess,
		OnScanEnd:  s.onEngineScanEnd,
	}
}

// onEngineSuccess runs on a worker goroutine. It only appends to the batch;
// no message is sent, so no Update/View cycle is triggered per result.
func (s *scanStage) onEngineSuccess(r result.Result) {
	s.batchMu.Lock()
	s.batch = append(s.batch, r)
	s.batchMu.Unlock()
}

// onEngineProgress forwards progress to the UI goroutine.
func (s *scanStage) onEngineProgress(p engine.Progress) {
	s.program.Send(stageProgressMsg{s.id, p})
}

// onEngineError forwards the error to the UI goroutine.
func (s *scanStage) onEngineError(err error) {
	s.program.Send(stageErrorMsg{s.id, err})
}

// onEngineScanEnd notifies the UI goroutine that the stage completed.
func (s *scanStage) onEngineScanEnd() {
	s.program.Send(stageEndMsg{s.id})
}

// currentProgress returns the progress percent as 0..1.
func (s *scanStage) currentProgress() float64 {
	return s.progress.Percent / 100
}

// cmpResult compares two results for sorting.
// It sorts primarily by Score in descending order, and secondarily by Key in ascending order.
func cmpResult(a, b result.Result) int {
	return cmp.Or(
		cmp.Compare(b.Score(), a.Score()),
		cmp.Compare(a.Key(), b.Key()),
	)
}

// takeBatch swaps the pending batch for a fresh slice and returns the old
// one. The lock is held only for the swap. The new slice has its own backing
// array, so the returned slice is safe to modify without the lock.
func (s *scanStage) takeBatch() []result.Result {
	s.batchMu.Lock()
	out := s.batch
	s.batch = make([]result.Result, 0, 128)
	s.batchMu.Unlock()
	return out
}

// mergeBatch drains the pending batch into results. UI goroutine only.
func (s *scanStage) mergeBatch() {
	batch := s.takeBatch()
	if len(batch) == 0 {
		return
	}

	// Normalize first: parsing may change score or key.
	schema := s.cfg.Probe.Schema()
	for i, r := range batch {
		if norm, err := schema.Parser(r.ToRecord()); err == nil {
			batch[i] = norm
		}
	}

	// Sort only the small batch, not all results.
	slices.SortStableFunc(batch, cmpResult)
	if len(batch) > s.maxIPs {
		batch = batch[:s.maxIPs]
	}

	// If the table is full and the best new item is no better than the
	// worst current one, nothing can enter.
	if len(s.results) >= s.maxIPs &&
		cmpResult(batch[0], s.results[len(s.results)-1]) >= 0 {
		return
	}

	s.results = mergeSorted(s.results, batch, s.maxIPs)

	if viewer := s.ipViewer(); viewer != nil {
		viewer.SetRows(s.results)
	}
}

// mergeSorted merges two sorted slices into a new sorted slice of at most
// limit elements. On ties, elements from a come first.
func mergeSorted(a, b []result.Result, limit int) []result.Result {
	out := make([]result.Result, 0, min(len(a)+len(b), limit))
	i, j := 0, 0

	for len(out) < limit && (i < len(a) || j < len(b)) {
		switch {
		case j >= len(b):
			out = append(out, a[i])
			i++
		case i >= len(a):
			out = append(out, b[j])
			j++
		case cmpResult(b[j], a[i]) < 0:
			out = append(out, b[j])
			j++
		default:
			out = append(out, a[i])
			i++
		}
	}
	return out
}

// updateProgressBar pushes the current progress into the progress bar.
func (s *scanStage) updateProgressBar() tea.Cmd {
	switch s.status {
	case StatusScanning:
		return progress.UpdateProgressMsg{
			ID:       s.progressBar.ID(),
			Progress: s.currentProgress(),
		}.Cmd()

	case StatusEnded:
		return progress.UpdateProgressMsg{
			ID:       s.progressBar.ID(),
			Progress: 1,
		}.Cmd()
	}
	return nil
}

// update forwards the message to the stage's components.
func (s *scanStage) update(msg tea.Msg) tea.Cmd {
	var tCmd, pCmd tea.Cmd
	s.table, tCmd = s.table.Update(msg)
	s.progressBar, pCmd = s.progressBar.Update(msg)
	return tea.Batch(tCmd, pCmd)
}

func (s *scanStage) setProgress(p engine.Progress) {
	if s.status == StatusWaiting {
		s.status = StatusScanning
	}
	s.lastSparkUpdated = time.Now()
	s.sparkline.Push(p.RatePerSec)
	s.progress = p
}

// resize pushes width/height to the stage's table.
func (s *scanStage) resize(w, h int) {
	if viewer := s.ipViewer(); viewer != nil {
		viewer.SetWidth(w)
		viewer.SetHeight(h)
	}
	if pb := s.progressModel(); pb != nil {
		pb.SetWidth(w)
	}
}

func (s *scanStage) finished() bool {
	return s.status == StatusEnded || s.status == StatusError
}
