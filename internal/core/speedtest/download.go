package speedtest

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/MohsenBg/bgscan/internal/logger"
)

const cloudflareDownURL = "https://speed.cloudflare.com/__down?bytes=%d"

var (
	downloadHTTPClientFactory = newHTTPClient
	downloadURLFmt            = cloudflareDownURL
	downloadNow               = time.Now
)

// DownloadConfig controls a single download measurement.
type DownloadConfig struct {
	// Bytes is the number of bytes to request from the test endpoint.
	// Must be > 0.
	Bytes int64

	// Timeout is the maximum time allowed for the body transfer.
	// Connection setup is excluded from this budget.
	// Zero (or negative) means no transfer timeout is applied.
	Timeout time.Duration

	// MinSpeed causes MeasureDownloadSpeed to return an error when the
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

// MeasureDownloadSpeed downloads cfg.Bytes and measures the throughput.
//
// Connection setup and TLS handshake use connectTimeout and are excluded
// from the transfer measurement window.
func measureDownloadSpeed(ctx context.Context, cfg DownloadConfig) (SpeedResult, error) {
	if cfg.Bytes < 0 {
		return SpeedResult{}, fmt.Errorf("download probe requires Bytes > 0")
	}

	client, err := downloadHTTPClientFactory(httpClientConfig{
		ProxyPort:   cfg.ProxyPort,
		DialContext: cfg.DialContext,
	})
	if err != nil {
		return SpeedResult{}, fmt.Errorf("download probe setup failed: %w", err)
	}
	defer client.CloseIdleConnections()

	testURL := fmt.Sprintf(downloadURLFmt, cfg.Bytes)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
	if err != nil {
		return SpeedResult{}, fmt.Errorf("download probe request build failed: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return SpeedResult{}, ctx.Err()
		}

		return SpeedResult{}, fmt.Errorf("download probe failed: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.CoreError("error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return SpeedResult{}, fmt.Errorf("download probe unexpected status: %s", resp.Status)
	}

	// A zero or negative Timeout means "no transfer deadline" — don't call
	// context.WithTimeout with a non-positive duration, since that expires
	// the context immediately rather than disabling the deadline.
	transferCtx := ctx
	if cfg.Timeout > 0 {
		var cancel context.CancelFunc
		transferCtx, cancel = context.WithTimeout(ctx, cfg.Timeout)
		defer cancel()
	}

	start := downloadNow()

	n, err := io.Copy(
		io.Discard,
		&contextReader{
			ctx: transferCtx,
			r:   resp.Body,
		},
	)

	elapsed := downloadNow().Sub(start)

	if err != nil {
		if errors.Is(transferCtx.Err(), context.DeadlineExceeded) {
			return SpeedResult{}, fmt.Errorf(
				"download probe timed out after %s: %w",
				cfg.Timeout,
				transferCtx.Err(),
			)
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
		return result, fmt.Errorf(
			"download speed %s is below minimum %s",
			result.Speed,
			cfg.MinSpeed,
		)
	}

	return result, nil
}
