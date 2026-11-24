// Package probe defines the core interface for all active scanning primitives.
// Each probe package under this one checks one protocol or tunnel type
// and reports its outcome as a result.Result. The engine calls Init
// once, then Run for each target, possibly from several goroutines at
// once, then Close once. Implementations must be safe for concurrent
// Run calls.
package probe

import (
	"context"

	"github.com/MohsenBg/bgscan/internal/core/result"
	"github.com/MohsenBg/bgscan/internal/core/target"
)

// Probe defines the interface for all active scanning primitives.
type Probe interface {
	// Init is called once before any Run calls and lets probes allocate
	// resources such as sockets, background goroutines, caches, or protocol
	// state. Implementations that require no initialization may return nil.
	Init(ctx context.Context) error

	// Input returns the kind of target entry the probe accepts, such
	// as IP addresses only or hosts that may be domains.
	Input() target.Kind

	// Run probes one target row. It must honor ctx for cancellation, return
	// a populated Result on success, and return an error if the probe fails
	// or times out. Use probe.NormalizeErr to map low-level timeouts and
	// refusals into the shared sentinel errors.
	Run(ctx context.Context, entry target.Entry) (result.Result, error)

	// Schema returns the schema describing this probe's results.
	Schema() result.ResultSchema

	// Close releases probe-specific resources such as sockets, goroutines,
	// or file descriptors. It is called once at the end of the scanner
	// lifecycle.
	Close() error
}
