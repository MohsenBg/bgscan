package engine

import (
	"strings"
	"time"

	"bgscan/internal/core/result"
	"bgscan/internal/core/scanner/probe"

	"golang.org/x/time/rate"
)

// PipelineMode defines how data streams and flows across multi-stage scans.
type PipelineMode string

const (
	// ModeSequential runs stages one after another.
	// Stage N+1 starts only after Stage N completely finishes and writes to disk.
	ModeSequential PipelineMode = "sequential"

	// ModeStreaming runs all stages concurrently using independent worker pools.
	// Successful IPs from Stage N are pushed instantly into a memory channel for Stage N+1.
	ModeStreaming PipelineMode = "streaming"

	// ModeBatch chunks incoming IPs into fixed-size arrays.
	// A batch must traverse all stages sequentially before the next batch is fetched.
	ModeBatch PipelineMode = "batch"
)

// ChainConfig controls the execution strategy for a multi-stage scan sequence.
type ChainConfig struct {
	// Mode selects the pipeline execution strategy (sequential, streaming, or batch).
	Mode PipelineMode

	// MaxBuffer is the channel buffer size between streaming stages.
	// Larger values reduce inter-stage blocking at the cost of memory.
	MaxBuffer int

	// BatchSize is the number of IPs grouped together when Mode is ModeBatch.
	BatchSize int

	// MaxIPsToTest caps the total number of IPs processed across all stages.
	MaxIPsToTest uint64

	// Stages defines the ordered list of scan stages to execute.
	Stages []StageConfig

	// MinProbeDuration enforces a minimum duration for each probe, useful for
	// normalizing timing-based side channels.
	MinProbeDuration time.Duration

	// Pause allows external control to pause/resume the scan chain.
	Pause PauseController

	// Shuffled randomizes IP order before scanning when true.
	Shuffled bool

	// RateLimiter throttles the rate of outgoing probes, if set.
	RateLimiter *rate.Limiter
}

// ScanConfig controls the execution of a single, standalone scan.
type ScanConfig struct {
	// Workers is the number of concurrent probe workers.
	Workers int

	// MaxIPsToTest caps the total number of IPs processed.
	MaxIPsToTest uint64

	// MinProbeDuration enforces a minimum duration for each probe.
	MinProbeDuration time.Duration

	// ProgressInterval sets how often OnProgress hooks fire.
	ProgressInterval time.Duration

	// Probe is the protocol-specific probe implementation to run against each IP.
	Probe probe.Probe

	// Writer persists successful scan results.
	Writer result.Writer

	// Hooks provides optional lifecycle callbacks.
	Hooks ScanHooks

	// Pause allows external control to pause/resume the scan.
	Pause PauseController

	// Shuffled randomizes IP order before scanning when true.
	Shuffled bool

	// RateLimiter throttles the rate of outgoing probes, if set.
	RateLimiter *rate.Limiter
}

// StageConfig defines settings and dependencies for a single scan stage.
type StageConfig struct {
	// Workers is the number of concurrent probe workers for this stage.
	Workers int

	// ProgressInterval sets how often OnProgress hooks fire for this stage.
	ProgressInterval time.Duration

	// Probe is the protocol-specific probe implementation for this stage.
	Probe probe.Probe

	// Writer persists successful results for this stage.
	Writer result.Writer

	// Hooks provides optional lifecycle callbacks for this stage.
	Hooks ScanHooks
}

// ScanHooks provides optional lifecycle callbacks for the scanning engine.
// All fields are optional — nil means the hook is disabled.
type ScanHooks struct {
	// OnProgress is called periodically with a scan progress snapshot.
	OnProgress func(Progress)

	// OnSuccess is called for each successfully scanned IP.
	OnSuccess func(result.Result)

	// OnScanEnd is called once after the entire scan finishes.
	OnScanEnd func()

	// OnError is called when a non-fatal engine error occurs.
	OnError func(error)
}

// callOnError safely invokes OnError if it has been provided.
func (h ScanHooks) callOnError(err error) {
	if h.OnError != nil {
		h.OnError(err)
	}
}

// callOnSuccess safely invokes OnSuccess if it has been provided.
func (h ScanHooks) callOnSuccess(r result.Result) {
	if h.OnSuccess != nil {
		h.OnSuccess(r)
	}
}

// callOnScanEnd safely invokes OnScanEnd if it has been provided.
func (h ScanHooks) callOnScanEnd() {
	if h.OnScanEnd != nil {
		h.OnScanEnd()
	}
}

// ParsePipelineMode converts an incoming configuration string into a valid PipelineMode.
// It gracefully defaults to ModeSequential if the input is empty or unrecognized.
func ParsePipelineMode(s string) PipelineMode {
	s = strings.TrimSpace(strings.ToLower(s))

	switch s {
	case "sequential", "simple":
		return ModeSequential
	case "streaming", "parallel":
		return ModeStreaming
	case "batch", "pipeline":
		return ModeBatch
	default:
		return ModeSequential
	}
}
