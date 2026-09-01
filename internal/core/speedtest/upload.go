package speedtest

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/MohsenBg/bgscan/internal/logger"
)

const cloudflareUpURL = "https://speed.cloudflare.com/__up"

var (
	uploadHTTPClientFactory = newHTTPClient
	uploadURL               = cloudflareUpURL
	uploadNow               = time.Now
)

// UploadConfig controls a single upload measurement.
type UploadConfig struct {
	// Bytes is the number of bytes to upload. Must be > 0.
	Bytes int64

	// Timeout is the maximum time allowed for the transfer.
	// Zero (or negative) means no timeout is applied.
	Timeout time.Duration

	// MinSpeed causes MeasureUploadSpeed to return an error when the
	// measured throughput falls below this threshold.
	MinSpeed BitsPerSec

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

// MeasureUploadSpeed uploads cfg.Bytes and measures the throughput.
//
// Unlike MeasureDownloadSpeed, connection setup and TLS handshake are
// included in the measurement window here, since the upload body starts
// streaming as part of the same request/response round trip and can't be
// cleanly separated from connection setup without a second connection.
func measureUploadSpeed(ctx context.Context, cfg UploadConfig) (SpeedResult, error) {
	if cfg.Bytes < 0 {
		return SpeedResult{}, fmt.Errorf("upload probe requires Bytes > 0")
	}

	client, err := uploadHTTPClientFactory(httpClientConfig{
		ProxyPort:   cfg.ProxyPort,
		DialContext: cfg.DialContext,
	})
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

	req, err := http.NewRequestWithContext(
		reqCtx,
		http.MethodPost,
		uploadURL,
		data,
	)
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
			return SpeedResult{}, fmt.Errorf(
				"upload probe timed out after %s: %w",
				cfg.Timeout,
				reqCtx.Err(),
			)
		}

		return SpeedResult{}, fmt.Errorf("upload probe failed: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.CoreError("error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return SpeedResult{}, fmt.Errorf("upload probe unexpected status: %s", resp.Status)
	}

	if _, err := io.Copy(io.Discard, resp.Body); err != nil {
		if errors.Is(reqCtx.Err(), context.DeadlineExceeded) {
			return SpeedResult{}, fmt.Errorf(
				"upload probe timed out after %s: %w",
				cfg.Timeout,
				reqCtx.Err(),
			)
		}

		return SpeedResult{}, fmt.Errorf(
			"upload probe response read failed: %w",
			err,
		)
	}

	if elapsed <= 0 {
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
