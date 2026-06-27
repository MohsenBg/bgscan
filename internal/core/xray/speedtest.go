package xray

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"bgscan/internal/logger"
)

const (
	cloudflareTraceURL = "https://speed.cloudflare.com/cdn-cgi/trace"
	cloudflareDownURL  = "https://speed.cloudflare.com/__down?bytes=%d"
	cloudflareUpURL    = "https://speed.cloudflare.com/__up"
)

func newHTTPClientWithSocks5(timeout time.Duration, port uint16) (*http.Client, error) {
	proxyURL, err := url.Parse(fmt.Sprintf("socks5://127.0.0.1:%d", port))
	if err != nil {
		return nil, fmt.Errorf("invalid proxy URL on port %d: %w", port, err)
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),

		DialContext: (&net.Dialer{
			Timeout:   timeout,
			KeepAlive: 30 * time.Second,
		}).DialContext,

		TLSHandshakeTimeout:   timeout,
		ResponseHeaderTimeout: timeout,
		ExpectContinueTimeout: timeout,
		IdleConnTimeout:       timeout,

		MaxIdleConns:        10,
		MaxIdleConnsPerHost: 10,
		MaxConnsPerHost:     10,

		ForceAttemptHTTP2: false,
	}

	return &http.Client{
		Transport: transport,
	}, nil
}

func MeasureLatency(ctx context.Context, timeout time.Duration, proxyPort uint16) (time.Duration, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	client, err := newHTTPClientWithSocks5(timeout, proxyPort)
	if err != nil {
		return 0, fmt.Errorf("latency probe setup failed: %w", err)
	}
	defer client.CloseIdleConnections()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cloudflareTraceURL, nil)
	if err != nil {
		return 0, fmt.Errorf("latency probe request build failed: %w", err)
	}

	start := time.Now()

	resp, err := client.Do(req)
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return 0, ctx.Err()
		}
		return 0, fmt.Errorf("latency probe failed (proxy port %d): %w", proxyPort, err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.CoreError("error closing response body: %v", err)
		}
	}()

	return time.Since(start), nil
}

func MeasureDownloadDuration(ctx context.Context, timeout time.Duration, bytesSize int64, proxyPort uint16) (time.Duration, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	client, err := newHTTPClientWithSocks5(timeout, proxyPort)
	if err != nil {
		return 0, fmt.Errorf("download probe setup failed: %w", err)
	}
	defer client.CloseIdleConnections()

	testURL := fmt.Sprintf(cloudflareDownURL, bytesSize)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
	if err != nil {
		return 0, fmt.Errorf("download probe request build failed: %w", err)
	}

	start := time.Now()

	resp, err := client.Do(req)
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return 0, ctx.Err()
		}
		return 0, fmt.Errorf("download probe failed (proxy port %d): %w", proxyPort, err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.CoreError("error closing response body: %v", err)
		}
	}()

	if _, err := io.Copy(io.Discard, resp.Body); err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return 0, ctx.Err()
		}
		return 0, fmt.Errorf("download probe body read failed: %w", err)
	}

	return time.Since(start), nil
}

func MeasureUploadDuration(ctx context.Context, timeout time.Duration, bytesSize int64, proxyPort uint16) (time.Duration, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	client, err := newHTTPClientWithSocks5(timeout, proxyPort)
	if err != nil {
		return 0, fmt.Errorf("upload probe setup failed: %w", err)
	}
	defer client.CloseIdleConnections()

	data := bytes.Repeat([]byte("A"), int(bytesSize))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cloudflareUpURL, bytes.NewReader(data))
	if err != nil {
		return 0, fmt.Errorf("upload probe request build failed: %w", err)
	}

	req.Header.Set("Content-Type", "application/octet-stream")

	start := time.Now()

	resp, err := client.Do(req)
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return 0, ctx.Err()
		}
		return 0, fmt.Errorf("upload probe failed (proxy port %d): %w", proxyPort, err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.CoreError("error closing response body: %v", err)
		}
	}()

	return time.Since(start), nil
}
