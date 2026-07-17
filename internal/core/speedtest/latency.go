package speedtest

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"bgscan/internal/logger"
)

const cloudflareTraceURL = "https://speed.cloudflare.com/cdn-cgi/trace"

// LatencyConfig controls a single latency measurement.
type LatencyConfig struct {
	// Timeout is the maximum time for the full round trip.
	Timeout time.Duration
	// MaxLatency, when non-zero, causes MeasureLatency to return an error
	// if the measured round-trip exceeds this threshold.
	MaxLatency time.Duration
	// ProxyPort is the local SOCKS5 proxy port to route traffic through.
	ProxyPort uint16
}

// LatencyResult holds the outcome of a single latency measurement.
type LatencyResult struct {
	RTT        time.Duration
	MaxLatency time.Duration // zero means no limit was set
}

// String returns a human-readable RTT, e.g. "42ms".
func (r LatencyResult) String() string {
	return r.RTT.String()
}

// AboveMaximum reports whether the RTT exceeded the configured maximum.
// Always false when MaxLatency is zero.
func (r LatencyResult) AboveMaximum() bool {
	return r.MaxLatency > 0 && r.RTT > r.MaxLatency
}

// MeasureLatency performs a single HTTP GET to Cloudflare's trace endpoint
// and returns a LatencyResult. Returns an error on timeout or if RTT exceeds
// cfg.MaxLatency.
func MeasureLatency(ctx context.Context, cfg LatencyConfig) (LatencyResult, error) {
	ctx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()

	client, err := newHTTPClient(cfg.ProxyPort)
	if err != nil {
		return LatencyResult{}, fmt.Errorf("latency probe setup failed: %w", err)
	}
	defer client.CloseIdleConnections()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cloudflareTraceURL, nil)
	if err != nil {
		return LatencyResult{}, fmt.Errorf("latency probe request build failed: %w", err)
	}

	start := time.Now()

	resp, err := client.Do(req)
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return LatencyResult{}, fmt.Errorf("latency probe timed out after %s: %w", cfg.Timeout, ctx.Err())
		}
		return LatencyResult{}, fmt.Errorf("latency probe failed (proxy port %d): %w", cfg.ProxyPort, err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.CoreError("error closing response body: %v", err)
		}
	}()

	result := LatencyResult{
		RTT:        time.Since(start),
		MaxLatency: cfg.MaxLatency,
	}

	if result.AboveMaximum() {
		return result, fmt.Errorf("latency %s exceeds maximum %s", result.RTT, cfg.MaxLatency)
	}

	return result, nil
}
