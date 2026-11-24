// Package engine provides the scan execution engine.
// It runs one stage or a chain of stages over a target file, with a
// worker pool per stage, optional pause support, rate limiting, and
// progress hooks. ChainConfig picks the pipeline mode: stages run one
// after another, stream targets to each other, or move in batches.
package engine

import (
	"strings"
	"time"

	"github.com/MohsenBg/bgscan/internal/core/result"
	"github.com/MohsenBg/bgscan/internal/core/scanner/probe"
	"github.com/MohsenBg/bgscan/internal/logger"

	"golang.org/x/time/rate"
)

// PipelineMode defines the data flow strategy across stages.
type PipelineMode string

const (
	// ModeSequential executes stages sequentially, writing each stage to disk before starting the next.
	ModeSequential PipelineMode = "sequential"

	// ModeStreaming executes stages concurrently, streaming successes directly between stages via channels.
	ModeStreaming PipelineMode = "streaming"

	// ModeBatch routes discrete batches of targets through all stages sequentially.
	ModeBatch PipelineMode = "batch"
)

// ChainConfig configures a multi-stage scan pipeline.
// MaxBuffer is the channel buffer size between stages in streaming
// mode, and BatchSize is the number of targets per batch in batch
// mode. Values at or below zero for either field use a default size.
// MaxSuccessfulIPs stops the scan once the final stage has that many
// successes, and zero means no limit. When IgnoreWriterError is false,
// a result writer error stops the scan. When it is true, the scan
// keeps running despite writer failures.
type ChainConfig struct {
	Mode              PipelineMode
	MaxBuffer         int
	BatchSize         int
	MaxIPsToTest      uint64
	Stages            []StageConfig
	MinProbeDuration  time.Duration
	Pause             PauseController
	Shuffled          bool
	RateLimiter       *rate.Limiter
	MaxSuccessfulIPs  uint64
	Log               *logger.Logger
	IgnoreWriterError bool
}

// StageConfig defines the execution parameters for a single scan stage.
type StageConfig struct {
	Workers          int
	ProgressInterval time.Duration
	Probe            probe.Probe
	Writer           result.Writer
	Hooks            ScanHooks
}

// ScanHooks defines lifecycle callbacks for stage execution.
type ScanHooks struct {
	// OnProgress is invoked periodically with progress metrics.
	// Implementations must be safe for concurrent execution.
	OnProgress func(Progress)

	// OnSuccess is invoked for each probe success.
	OnSuccess func(result.Result)

	// OnScanEnd is invoked when the stage completes.
	OnScanEnd func()
}

func (h ScanHooks) callOnSuccess(r result.Result) {
	if h.OnSuccess != nil {
		h.OnSuccess(r)
	}
}

func (h ScanHooks) callOnScanEnd() {
	if h.OnScanEnd != nil {
		h.OnScanEnd()
	}
}

// ParsePipelineMode returns the PipelineMode matching s, defaulting to ModeSequential.
func ParsePipelineMode(s string) PipelineMode {
	switch strings.TrimSpace(strings.ToLower(s)) {
	case "sequential":
		return ModeSequential
	case "streaming":
		return ModeStreaming
	case "batch":
		return ModeBatch
	default:
		return ModeSequential
	}
}
