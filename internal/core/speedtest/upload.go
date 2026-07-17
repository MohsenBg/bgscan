package speedtest

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"bgscan/internal/logger"
)

const cloudflareUpURL = "https://speed.cloudflare.com/__up"

// UploadConfig controls a single upload measurement.
type UploadConfig struct {
	// Bytes is the number of bytes to upload.
	Bytes int64
	// Timeout is the maximum time allowed for the transfer, starting at
	// first byte sent.
	Timeout time.Duration
	// MinSpeed, when non-zero, causes MeasureUploadSpeed to return an error
	// if the measured throughput falls below this threshold.
	MinSpeed BitsPerSec
	// ProxyPort is the local SOCKS5 proxy port to route traffic through.
	ProxyPort uint16
}

// MeasureUploadSpeed uploads cfg.Bytes through the SOCKS5 proxy and
// returns a SpeedResult. The timeout covers only the transfer window,
// starting from first byte sent. Connection setup is excluded.
//
// Errors are returned for: connection failure, transfer timeout, zero data
// sent, or measured speed below cfg.MinSpeed.
func MeasureUploadSpeed(ctx context.Context, cfg UploadConfig) (SpeedResult, error) {
	client, err := newHTTPClient(cfg.ProxyPort)
	if err != nil {
		return SpeedResult{}, fmt.Errorf("upload probe setup failed: %w", err)
	}
	defer client.CloseIdleConnections()

	data := bytes.NewReader(make([]byte, cfg.Bytes))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cloudflareUpURL, data)
	if err != nil {
		return SpeedResult{}, fmt.Errorf("upload probe request build failed: %w", err)
	}
	req.ContentLength = cfg.Bytes
	req.Header.Set("Content-Type", "application/octet-stream")

	// Measure the request duration.
	start := time.Now()

	resp, err := client.Do(req)

	elapsed := time.Since(start)

	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return SpeedResult{}, fmt.Errorf("upload probe timed out after %s: %w", cfg.Timeout, ctx.Err())
		}
		return SpeedResult{}, fmt.Errorf("upload probe failed (proxy port %d): %w", cfg.ProxyPort, err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.CoreError("error closing response body: %v", err)
		}
	}()

	// Drain the response so the request fully completes.
	if _, err := io.Copy(io.Discard, resp.Body); err != nil {
		return SpeedResult{}, fmt.Errorf("upload probe response read failed: %w", err)
	}

	if elapsed <= 0 || cfg.Bytes == 0 {
		return SpeedResult{}, fmt.Errorf("upload probe sent no data")
	}

	result := SpeedResult{
		Speed:    bitsPerSec(uint64(cfg.Bytes), elapsed.Seconds()),
		Bytes:    uint64(cfg.Bytes),
		MinSpeed: cfg.MinSpeed,
	}

	if result.BelowMinimum() {
		return result, fmt.Errorf(
			"upload speed %s is below minimum %s",
			result.Speed,
			cfg.MinSpeed,
		)
	}

	return result, nil
}
