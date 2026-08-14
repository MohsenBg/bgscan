package speedtest

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"bgscan/internal/logger"
)

const (
	CloudflareTraceHTTP  = "http://speed.cloudflare.com/cdn-cgi/trace"
	CloudflareTraceHTTPS = "https://speed.cloudflare.com/cdn-cgi/trace"

	GoogleGenerate204HTTP  = "http://www.google.com/generate_204"
	GoogleGenerate204HTTPS = "https://www.google.com/generate_204"
)

var (
	latencyHTTPClientFactory = newHTTPClient
	latencyNow               = time.Now
)

const defaultLatencyProbeURL = CloudflareTraceHTTPS

// LatencyConfig controls a single latency measurement.
type LatencyConfig struct {
	// URL is the HTTP endpoint used for latency measurement.
	// Empty uses CloudflareTraceHTTPS.
	URL string

	// Timeout is the maximum time allowed for the round trip.
	// Zero (or negative) means no timeout is applied.
	Timeout time.Duration

	// MaxLatency causes MeasureLatency to return an error when the
	// measured RTT exceeds this threshold.
	MaxLatency time.Duration

	// ProxyPort is the local SOCKS5 proxy port used when DialContext is nil.
	ProxyPort uint16

	// DialContext provides the connection used to reach the test endpoint.
	// When set, ProxyPort is ignored.
	DialContext func(
		ctx context.Context,
		network string,
		address string,
	) (net.Conn, error)
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

func measureLatency(ctx context.Context, cfg LatencyConfig) (LatencyResult, error) {
	reqCtx := ctx

	if cfg.Timeout > 0 {
		var cancel context.CancelFunc
		reqCtx, cancel = context.WithTimeout(ctx, cfg.Timeout)
		defer cancel()
	}

	client, err := latencyHTTPClientFactory(httpClientConfig{
		ProxyPort:   cfg.ProxyPort,
		DialContext: cfg.DialContext,
	})
	if err != nil {
		return LatencyResult{}, fmt.Errorf("latency probe setup failed: %w", err)
	}
	defer client.CloseIdleConnections()

	url := cfg.URL
	if url == "" {
		url = defaultLatencyProbeURL
	}

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, url, nil)
	if err != nil {
		return LatencyResult{}, fmt.Errorf("latency probe request build failed: %w", err)
	}

	start := latencyNow()

	resp, err := client.Do(req)
	if err != nil {
		if errors.Is(reqCtx.Err(), context.DeadlineExceeded) {
			return LatencyResult{}, fmt.Errorf(
				"latency probe timed out after %s: %w",
				cfg.Timeout,
				reqCtx.Err(),
			)
		}

		return LatencyResult{}, fmt.Errorf("latency probe failed: %w", err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.CoreError("error closing response body: %v", err)
		}
	}()

	// Cloudflare trace returns 200, Google generate_204 returns 204.
	if resp.StatusCode != http.StatusOK &&
		resp.StatusCode != http.StatusNoContent {
		return LatencyResult{}, fmt.Errorf(
			"latency probe unexpected status: %s",
			resp.Status,
		)
	}

	result := LatencyResult{
		RTT:        latencyNow().Sub(start),
		MaxLatency: cfg.MaxLatency,
	}

	if result.AboveMaximum() {
		return result, fmt.Errorf(
			"latency %s exceeds maximum %s",
			result.RTT,
			cfg.MaxLatency,
		)
	}

	return result, nil
}
