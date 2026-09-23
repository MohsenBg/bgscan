// Package probe defines the core interface for all active scanning primitives.
package probe

import (
	"context"
	"errors"
	"net/netip"

	"github.com/MohsenBg/bgscan/internal/core/result"
)

// ErrEnvironment marks a Run failure caused by the local scanning
// environment — a missing OS capability (e.g. no IPv6 route), an exhausted
// local port pool, or a broken local proxy/Xray instance — rather than the
// scanned target. Probes wrap it with %w on such failures; the engine
// counts consecutive ErrEnvironment failures and aborts the scan once the
// streak proves retrying more IPs cannot succeed.
var ErrEnvironment = errors.New("probe environment failure")

// Probe defines the interface for all active scanning primitives.
type Probe interface {
	// Init is called once before any Run calls and lets probes allocate
	// resources such as sockets, background goroutines, caches, or protocol
	// state. Implementations that require no initialization may return nil.
	Init(ctx context.Context) error

	// Run executes a probe against the provided IP address. It must honor
	// ctx for cancellation, return a populated Result on success, and
	// return an error if the probe fails or times out. Failures caused by
	// the local environment (not the target) must wrap ErrEnvironment.
	Run(ctx context.Context, ip netip.Addr) (result.Result, error)

	// Schema returns the schema describing this probe's results.
	Schema() result.ResultSchema

	// Close releases probe-specific resources such as sockets, goroutines,
	// or file descriptors. It is called once at the end of the scanner
	// lifecycle.
	Close() error
}
