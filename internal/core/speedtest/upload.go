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

var (
	uploadHTTPClientFactory = newHTTPClient
	uploadURL               = cloudflareUpURL
	uploadNow               = time.Now
)

// UploadConfig controls a single upload measurement.
type UploadConfig struct {
	// Bytes is the number of bytes to upload.
	Bytes int64
	// Timeout is the maximum time allowed for the transfer.
	Timeout time.Duration
	// MinSpeed, when non-zero, causes MeasureUploadSpeed to return an error
	// if the measured throughput falls below this threshold.
	MinSpeed BitsPerSec
	// ProxyPort is the local SOCKS5 proxy port to route traffic through.
	ProxyPort uint16
}

// MeasureUploadSpeed uploads cfg.Bytes through the SOCKS5 proxy and
// returns a SpeedResult.
//
// Errors are returned for: connection failure, timeout, zero data
// sent, or measured speed below cfg.MinSpeed.
func MeasureUploadSpeed(ctx context.Context, cfg UploadConfig) (SpeedResult, error) {
	client, err := uploadHTTPClientFactory(cfg.ProxyPort)
	if err != nil {
		return SpeedResult{}, fmt.Errorf("upload probe setup failed: %w", err)
	}
	defer client.CloseIdleConnections()

	data := bytes.NewReader(make([]byte, cfg.Bytes))

	reqCtx := ctx
	var cancel context.CancelFunc
	if cfg.Timeout > 0 {
		reqCtx, cancel = context.WithTimeout(ctx, cfg.Timeout)
		defer cancel()
	}

	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, uploadURL, data)
	if err != nil {
		return SpeedResult{}, fmt.Errorf("upload probe request build failed: %w", err)
	}
	req.ContentLength = cfg.Bytes
	req.Header.Set("Content-Type", "application/octet-stream")

	start := uploadNow()
	resp, err := client.Do(req)
	elapsed := uploadNow().Sub(start)

	if err != nil {
		if errors.Is(reqCtx.Err(), context.DeadlineExceeded) {
			return SpeedResult{}, fmt.Errorf("upload probe timed out after %s: %w", cfg.Timeout, reqCtx.Err())
		}
		return SpeedResult{}, fmt.Errorf("upload probe failed (proxy port %d): %w", cfg.ProxyPort, err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.CoreError("error closing response body: %v", err)
		}
	}()

	if _, err := io.Copy(io.Discard, resp.Body); err != nil {
		if errors.Is(reqCtx.Err(), context.DeadlineExceeded) {
			return SpeedResult{}, fmt.Errorf("upload probe timed out after %s: %w", cfg.Timeout, reqCtx.Err())
		}
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
