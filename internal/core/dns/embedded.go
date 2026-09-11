package dns

import (
	"context"
	"fmt"
	"net/netip"
	"strings"
	"sync/atomic"
	"time"
)

// readyPollInterval is how often the embedded tunnel wait loop polls
// for session readiness.
const readyPollInterval = 200 * time.Millisecond

// parseTunnelEndpoint validates the runtime tunnel endpoint args shared
// by the embedded services: resolverIP must parse, listenPort must be
// non-zero.
func parseTunnelEndpoint(resolverIP string, listenPort uint16) (netip.Addr, error) {
	if strings.TrimSpace(resolverIP) == "" {
		return netip.Addr{}, fmt.Errorf("resolver IP is required")
	}

	if listenPort == 0 {
		return netip.Addr{}, fmt.Errorf("listen port must be greater than zero")
	}

	addr, err := netip.ParseAddr(strings.TrimSpace(resolverIP))
	if err != nil {
		return netip.Addr{}, fmt.Errorf("invalid resolver IP %q: %w", resolverIP, err)
	}

	return addr, nil
}

// embeddedClient is the minimal surface both vendor clients expose for
// the shared boot wait loop.
type embeddedClient interface {
	SessionReady() bool
	Run(ctx context.Context) error
	StopAsyncRuntime()
}

// embeddedHandle is a running in-process tunnel client (local SOCKS5
// listener). Close stops it; it is safe to call twice.
type embeddedHandle[C embeddedClient] struct {
	cancel  context.CancelFunc
	app     C
	stopped atomic.Bool
}

// Close stops the client runtime. It is safe to call twice.
func (h *embeddedHandle[C]) Close() error {
	if h == nil || h.stopped.Swap(true) {
		return nil
	}

	h.cancel()
	h.app.StopAsyncRuntime()

	return nil
}

// bootEmbedded runs app until its session is ready, then returns a handle.
// The caller must Close the handle, otherwise goroutines and sockets leak.
func bootEmbedded[C embeddedClient](ctx context.Context, app C) (*embeddedHandle[C], error) {
	runCtx, cancel := context.WithCancel(context.Background())

	runErr := make(chan error, 1)
	go func() {
		runErr <- app.Run(runCtx)
	}()

	for {
		if app.SessionReady() {
			return &embeddedHandle[C]{
				cancel: cancel,
				app:    app,
			}, nil
		}

		select {
		case err := <-runErr:
			cancel()
			app.StopAsyncRuntime()

			if err != nil {
				return nil, fmt.Errorf("client runtime exited: %w", err)
			}
			return nil, fmt.Errorf("client runtime exited before session ready")

		case <-ctx.Done():
			cancel()
			app.StopAsyncRuntime()

			return nil, fmt.Errorf(
				"context canceled while waiting for session: %w",
				ctx.Err(),
			)

		case <-time.After(readyPollInterval):
		}
	}
}
