package engine

import (
	"strings"

	"bgscan/internal/core/result"
	"bgscan/internal/core/scanner/probe"
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
	Mode PipelineMode

	// MaxBuffer is the channel buffer size between streaming stages.
	// Larger values reduce inter-stage blocking at the cost of memory.
	MaxBuffer int

	Stages []ScanConfig

	Pause *PauseController

	Shuffled bool
}

// ScanConfig defines settings and dependencies for a single scan stage.
type ScanConfig struct {
	Workers int
	Rate    int

	Probe  probe.Probe
	Writer *result.Writer
	Hooks  ScanHooks
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
