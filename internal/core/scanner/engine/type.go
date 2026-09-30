// Package engine provides the scan execution engine.
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
type ChainConfig struct {
	Mode             PipelineMode
	MaxBuffer        int
	BatchSize        int
	MaxIPsToTest     uint64
	Stages           []StageConfig
	MinProbeDuration time.Duration
	Pause            PauseController
	Shuffled         bool
	RateLimiter      *rate.Limiter
	MaxSuccessfulIPs uint64
	Log              *logger.Logger
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

	// OnError is invoked when a fatal error occurs.
	OnError func(error)
}

func (h ScanHooks) callOnError(err error) {
	if h.OnError != nil {
		h.OnError(err)
	}
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
