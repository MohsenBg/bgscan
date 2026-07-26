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

var (
	httpClientFactory = newHTTPClient
	latencyProbeURL   = cloudflareTraceURL
	now               = time.Now
)

type LatencyConfig struct {
	Timeout    time.Duration
	MaxLatency time.Duration
	ProxyPort  uint16
}

type LatencyResult struct {
	RTT        time.Duration
	MaxLatency time.Duration
}

func (r LatencyResult) String() string {
	return r.RTT.String()
}

func (r LatencyResult) AboveMaximum() bool {
	return r.MaxLatency > 0 && r.RTT > r.MaxLatency
}

func MeasureLatency(ctx context.Context, cfg LatencyConfig) (LatencyResult, error) {
	ctx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()

	client, err := httpClientFactory(cfg.ProxyPort)
	if err != nil {
		return LatencyResult{}, fmt.Errorf("latency probe setup failed: %w", err)
	}
	defer client.CloseIdleConnections()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, latencyProbeURL, nil)
	if err != nil {
		return LatencyResult{}, fmt.Errorf("latency probe request build failed: %w", err)
	}

	start := now()

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
