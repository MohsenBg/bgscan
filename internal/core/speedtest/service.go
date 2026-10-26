package speedtest

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Sentinel errors returned by Service methods. Use errors.Is to test.
var (
	// ErrBelowMinimum is wrapped when a measurement succeeds but falls
	// short of the configured minimum speed.
	ErrBelowMinimum = errors.New("speedtest: below minimum speed")

	// ErrAboveMaximum is wrapped when a measurement succeeds but exceeds
	// the configured maximum latency.
	ErrAboveMaximum = errors.New("speedtest: above maximum latency")
)

type Service interface {
	MeasureLatency(context.Context, LatencyConfig) (LatencyResult, error)
	MeasureDownloadSpeed(context.Context, DownloadConfig) (SpeedResult, error)
	MeasureUploadSpeed(context.Context, UploadConfig) (SpeedResult, error)
}

const (
	CloudflareTraceHTTP    = "http://speed.cloudflare.com/cdn-cgi/trace"
	CloudflareTraceHTTPS   = "https://speed.cloudflare.com/cdn-cgi/trace"
	GoogleGenerate204HTTP  = "http://www.google.com/generate_204"
	GoogleGenerate204HTTPS = "https://www.google.com/generate_204"
	CloudflareDownloadURL  = "https://speed.cloudflare.com/__down?bytes=%d"
	CloudflareUploadURL    = "https://speed.cloudflare.com/__up"
)

type service struct {
	downloadURLFmt string
	uploadURL      string
	now            func() time.Time
	clientFactory  func(cfg httpClientConfig) (*http.Client, error)
}

func NewService() Service {
	return &service{
		downloadURLFmt: CloudflareDownloadURL,
		uploadURL:      CloudflareUploadURL,
		now:            time.Now,
		clientFactory:  newHTTPClient,
	}
}

func (s *service) MeasureLatency(
	ctx context.Context,
	cfg LatencyConfig,
) (LatencyResult, error) {
	reqCtx := ctx

	if cfg.Timeout > 0 {
		var cancel context.CancelFunc
		reqCtx, cancel = context.WithTimeout(ctx, cfg.Timeout)
		defer cancel()
	}

	client, err := s.clientFactory(httpClientConfig{
		ProxyPort:   cfg.ProxyPort,
		DialContext: cfg.DialContext,
	})
	if err != nil {
		return LatencyResult{}, fmt.Errorf("latency probe setup failed: %w", err)
	}
	defer client.CloseIdleConnections()

	url := cfg.URL
	if url == "" {
		url = GoogleGenerate204HTTP
	}

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, url, nil)
	if err != nil {
		return LatencyResult{}, fmt.Errorf("latency probe request build failed: %w", err)
	}

	start := s.now()

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
	defer func() { _ = resp.Body.Close() }()

	// Cloudflare trace returns 200; Google generate_204 returns 204.
	if resp.StatusCode != http.StatusOK &&
		resp.StatusCode != http.StatusNoContent {
		return LatencyResult{}, fmt.Errorf(
			"latency probe unexpected status: %s",
			resp.Status,
		)
	}

	result := LatencyResult{
		RTT:        s.now().Sub(start),
		MaxLatency: cfg.MaxLatency,
	}

	if result.AboveMaximum() {
		// Do NOT wrap context.DeadlineExceeded: no deadline fired, the
		// probe simply completed too slowly. Callers that branch on
		// DeadlineExceeded must not confuse the two.
		return result, fmt.Errorf(
			"latency %s exceeds maximum %s: %w",
			result.RTT,
			cfg.MaxLatency,
			ErrAboveMaximum,
		)
	}

	return result, nil
}

func (s *service) MeasureDownloadSpeed(
	ctx context.Context,
	cfg DownloadConfig,
) (SpeedResult, error) {
	if cfg.Bytes <= 0 {
		return SpeedResult{}, fmt.Errorf("download probe requires Bytes > 0")
	}

	client, err := s.clientFactory(httpClientConfig{
		ProxyPort:   cfg.ProxyPort,
		DialContext: cfg.DialContext,
	})
	if err != nil {
		return SpeedResult{}, fmt.Errorf("download probe setup failed: %w", err)
	}
	defer client.CloseIdleConnections()

	testURL := fmt.Sprintf(s.downloadURLFmt, cfg.Bytes)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
	if err != nil {
		return SpeedResult{}, fmt.Errorf("download probe request build failed: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return SpeedResult{}, ctx.Err()
		}
		return SpeedResult{}, fmt.Errorf("download probe failed: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return SpeedResult{}, fmt.Errorf("download probe unexpected status: %s", resp.Status)
	}

	// Transfer timeout starts after headers arrive; connection setup is
	// excluded from the budget.
	transferCtx := ctx
	if cfg.Timeout > 0 {
		var cancel context.CancelFunc
		transferCtx, cancel = context.WithTimeout(ctx, cfg.Timeout)
		defer cancel()
	}

	stopClose := context.AfterFunc(transferCtx, func() {
		_ = resp.Body.Close()
	})
	defer stopClose()

	start := s.now()
	n, err := io.Copy(io.Discard, resp.Body)
	elapsed := s.now().Sub(start)

	if err != nil {
		if ctx.Err() != nil {
			return SpeedResult{}, ctx.Err()
		}
		if cfg.Timeout > 0 && transferCtx.Err() == context.DeadlineExceeded {
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
			"download speed %s is below minimum %s: %w",
			result.Speed,
			cfg.MinSpeed,
			ErrBelowMinimum,
		)
	}

	return result, nil
}

func (s *service) MeasureUploadSpeed(
	ctx context.Context,
	cfg UploadConfig,
) (SpeedResult, error) {
	if cfg.Bytes <= 0 {
		return SpeedResult{}, fmt.Errorf("upload probe requires Bytes > 0")
	}

	client, err := s.clientFactory(httpClientConfig{
		ProxyPort:   cfg.ProxyPort,
		DialContext: cfg.DialContext,
	})
	if err != nil {
		return SpeedResult{}, fmt.Errorf("upload probe setup failed: %w", err)
	}
	defer client.CloseIdleConnections()

	data := bytes.NewReader(make([]byte, cfg.Bytes))

	reqCtx := ctx
	if cfg.Timeout > 0 {
		var cancel context.CancelFunc
		reqCtx, cancel = context.WithTimeout(ctx, cfg.Timeout)
		defer cancel()
	}

	req, err := http.NewRequestWithContext(
		reqCtx,
		http.MethodPost,
		s.uploadURL,
		data,
	)
	if err != nil {
		return SpeedResult{}, fmt.Errorf("upload probe request build failed: %w", err)
	}

	req.ContentLength = cfg.Bytes
	req.Header.Set("Content-Type", "application/octet-stream")

	start := s.now()
	resp, err := client.Do(req)
	elapsed := s.now().Sub(start)

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
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return SpeedResult{}, fmt.Errorf("upload probe unexpected status: %s", resp.Status)
	}

	stopClose := context.AfterFunc(reqCtx, func() {
		_ = resp.Body.Close()
	})
	defer stopClose()

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
			"upload speed %s is below minimum %s: %w",
			result.Speed,
			cfg.MinSpeed,
			ErrBelowMinimum,
		)
	}

	return result, nil
}
