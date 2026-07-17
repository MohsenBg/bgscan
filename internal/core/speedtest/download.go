package speedtest

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"bgscan/internal/logger"
)

const cloudflareDownURL = "https://speed.cloudflare.com/__down?bytes=%d"

// DownloadConfig controls a single download measurement.
type DownloadConfig struct {
	// Bytes is the number of bytes to request from the test endpoint.
	Bytes int64
	// Timeout is the maximum time allowed for the body transfer.
	// Connection setup is excluded from this budget.
	Timeout time.Duration
	// MinSpeed, when non-zero, causes MeasureDownloadSpeed to return an error
	// if the measured throughput falls below this threshold.
	MinSpeed BitsPerSec
	// ProxyPort is the local SOCKS5 proxy port to route traffic through.
	ProxyPort uint16
}

// MeasureDownloadSpeed downloads cfg.Bytes through the SOCKS5 proxy and
// returns a SpeedResult. The timeout covers only the body transfer; connection
// setup and TLS handshake use connectTimeout and are excluded from the window.
//
// Errors are returned for: connection failure, transfer timeout, zero data
// received, or measured speed below cfg.MinSpeed.
func MeasureDownloadSpeed(ctx context.Context, cfg DownloadConfig) (SpeedResult, error) {
	client, err := newHTTPClient(cfg.ProxyPort)
	if err != nil {
		return SpeedResult{}, fmt.Errorf("download probe setup failed: %w", err)
	}
	defer client.CloseIdleConnections()

	testURL := fmt.Sprintf(cloudflareDownURL, cfg.Bytes)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
	if err != nil {
		return SpeedResult{}, fmt.Errorf("download probe request build failed: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return SpeedResult{}, ctx.Err()
		}
		return SpeedResult{}, fmt.Errorf("download probe failed (proxy port %d): %w", cfg.ProxyPort, err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.CoreError("error closing response body: %v", err)
		}
	}()

	transferCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()

	start := time.Now()
	n, err := io.Copy(io.Discard, &contextReader{ctx: transferCtx, r: resp.Body})
	elapsed := time.Since(start)

	if err != nil {
		if errors.Is(transferCtx.Err(), context.DeadlineExceeded) {
			return SpeedResult{}, fmt.Errorf("download probe timed out after %s: %w", cfg.Timeout, transferCtx.Err())
		}
		return SpeedResult{}, fmt.Errorf("download probe body read failed: %w", err)
	}

	if elapsed <= 0 || n == 0 {
		return SpeedResult{}, fmt.Errorf("download probe returned no data")
	}

	result := SpeedResult{
		Speed:    bitsPerSec(uint64(n), elapsed.Seconds()),
		Bytes:    uint64(n),
		MinSpeed: cfg.MinSpeed,
	}

	if result.BelowMinimum() {
		return result, fmt.Errorf("download speed %s is below minimum %s", result.Speed, cfg.MinSpeed)
	}

	return result, nil
}
