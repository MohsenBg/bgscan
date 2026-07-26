package speedtest

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type fakeReadCloser struct {
	readFn  func([]byte) (int, error)
	closeFn func() error
}

func (f fakeReadCloser) Read(p []byte) (int, error) {
	return f.readFn(p)
}

func (f fakeReadCloser) Close() error {
	if f.closeFn != nil {
		return f.closeFn()
	}
	return nil
}

func restoreDownloadDeps() func() {
	oldFactory := downloadHTTPClientFactory
	oldURLFmt := downloadURLFmt
	oldNow := downloadNow

	return func() {
		downloadHTTPClientFactory = oldFactory
		downloadURLFmt = oldURLFmt
		downloadNow = oldNow
	}
}

func setDownloadNowSequence(t *testing.T, ts ...time.Time) {
	t.Helper()

	var i int
	downloadNow = func() time.Time {
		if i >= len(ts) {
			t.Fatalf("downloadNow called more than expected; calls=%d", i+1)
		}
		v := ts[i]
		i++
		return v
	}
}

func TestMeasureDownloadSpeed_SetupFailed(t *testing.T) {
	defer restoreDownloadDeps()()

	downloadHTTPClientFactory = func(port uint16) (*http.Client, error) {
		if port != 1080 {
			t.Fatalf("port = %d, want 1080", port)
		}
		return nil, errors.New("factory boom")
	}

	_, err := MeasureDownloadSpeed(context.Background(), DownloadConfig{
		Bytes:     1024,
		Timeout:   time.Second,
		ProxyPort: 1080,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "download probe setup failed") {
		t.Fatalf("error = %q, want to contain %q", err.Error(), "download probe setup failed")
	}
	if !strings.Contains(err.Error(), "factory boom") {
		t.Fatalf("error = %q, want to contain %q", err.Error(), "factory boom")
	}
}

func TestMeasureDownloadSpeed_RequestBuildFailed(t *testing.T) {
	defer restoreDownloadDeps()()

	downloadHTTPClientFactory = func(port uint16) (*http.Client, error) {
		return &http.Client{}, nil
	}
	downloadURLFmt = "://bad-url-%d"

	_, err := MeasureDownloadSpeed(context.Background(), DownloadConfig{
		Bytes:     1024,
		Timeout:   time.Second,
		ProxyPort: 1080,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "download probe request build failed") {
		t.Fatalf("error = %q, want to contain %q", err.Error(), "download probe request build failed")
	}
}

func TestMeasureDownloadSpeed_RequestFailed(t *testing.T) {
	defer restoreDownloadDeps()()

	downloadHTTPClientFactory = func(port uint16) (*http.Client, error) {
		if port != 1080 {
			t.Fatalf("port = %d, want 1080", port)
		}

		return &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if req.Method != http.MethodGet {
					t.Fatalf("method = %s, want %s", req.Method, http.MethodGet)
				}
				if req.URL.String() != "https://example.com/__down?bytes=2048" {
					t.Fatalf("url = %q, want %q", req.URL.String(), "https://example.com/__down?bytes=2048")
				}
				return nil, errors.New("network down")
			}),
		}, nil
	}
	downloadURLFmt = "https://example.com/__down?bytes=%d"

	_, err := MeasureDownloadSpeed(context.Background(), DownloadConfig{
		Bytes:     2048,
		Timeout:   time.Second,
		ProxyPort: 1080,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "download probe failed (proxy port 1080)") {
		t.Fatalf("error = %q, want to contain %q", err.Error(), "download probe failed (proxy port 1080)")
	}
	if !strings.Contains(err.Error(), "network down") {
		t.Fatalf("error = %q, want to contain %q", err.Error(), "network down")
	}
}

func TestMeasureDownloadSpeed_BodyReadFailed(t *testing.T) {
	defer restoreDownloadDeps()()

	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.Add(2 * time.Second)
	setDownloadNowSequence(t, start, end)

	downloadHTTPClientFactory = func(port uint16) (*http.Client, error) {
		return &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body: fakeReadCloser{
						readFn: func(p []byte) (int, error) {
							return 0, errors.New("read boom")
						},
					},
				}, nil
			}),
		}, nil
	}
	downloadURLFmt = "https://example.com/__down?bytes=%d"

	_, err := MeasureDownloadSpeed(context.Background(), DownloadConfig{
		Bytes:     512,
		Timeout:   time.Second,
		ProxyPort: 1080,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "download probe body read failed") {
		t.Fatalf("error = %q, want to contain %q", err.Error(), "download probe body read failed")
	}
	if !strings.Contains(err.Error(), "read boom") {
		t.Fatalf("error = %q, want to contain %q", err.Error(), "read boom")
	}
}

func TestMeasureDownloadSpeed_Timeout(t *testing.T) {
	defer restoreDownloadDeps()()

	downloadHTTPClientFactory = func(port uint16) (*http.Client, error) {
		return &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader("abcdef")),
				}, nil
			}),
		}, nil
	}
	downloadURLFmt = "https://example.com/__down?bytes=%d"

	_, err := MeasureDownloadSpeed(context.Background(), DownloadConfig{
		Bytes:     6,
		Timeout:   0,
		ProxyPort: 1080,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "download probe timed out after") {
		t.Fatalf("error = %q, want timeout message", err.Error())
	}
}

func TestMeasureDownloadSpeed_NoData(t *testing.T) {
	defer restoreDownloadDeps()()

	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.Add(1 * time.Second)
	setDownloadNowSequence(t, start, end)

	downloadHTTPClientFactory = func(port uint16) (*http.Client, error) {
		return &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader("")),
				}, nil
			}),
		}, nil
	}
	downloadURLFmt = "https://example.com/__down?bytes=%d"

	_, err := MeasureDownloadSpeed(context.Background(), DownloadConfig{
		Bytes:     0,
		Timeout:   time.Second,
		ProxyPort: 1080,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "download probe returned no data") {
		t.Fatalf("error = %q, want to contain %q", err.Error(), "download probe returned no data")
	}
}

func TestMeasureDownloadSpeed_Success(t *testing.T) {
	defer restoreDownloadDeps()()

	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.Add(2 * time.Second)
	setDownloadNowSequence(t, start, end)

	downloadHTTPClientFactory = func(port uint16) (*http.Client, error) {
		if port != 1080 {
			t.Fatalf("port = %d, want 1080", port)
		}

		return &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if req.Method != http.MethodGet {
					t.Fatalf("method = %s, want %s", req.Method, http.MethodGet)
				}
				if req.URL.String() != "https://example.com/__down?bytes=4" {
					t.Fatalf("url = %q, want %q", req.URL.String(), "https://example.com/__down?bytes=4")
				}

				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader("abcd")),
				}, nil
			}),
		}, nil
	}
	downloadURLFmt = "https://example.com/__down?bytes=%d"

	result, err := MeasureDownloadSpeed(context.Background(), DownloadConfig{
		Bytes:     4,
		Timeout:   time.Second,
		ProxyPort: 1080,
	})
	if err != nil {
		t.Fatalf("MeasureDownloadSpeed() unexpected error = %v", err)
	}

	if got, want := result.Bytes, uint64(4); got != want {
		t.Fatalf("Bytes = %d, want %d", got, want)
	}

	if got, want := result.Speed, bitsPerSec(4, 2); got != want {
		t.Fatalf("Speed = %v, want %v", got, want)
	}

	if result.MinSpeed != 0 {
		t.Fatalf("MinSpeed = %v, want 0", result.MinSpeed)
	}
}

func TestMeasureDownloadSpeed_BelowMinimum(t *testing.T) {
	defer restoreDownloadDeps()()

	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.Add(2 * time.Second)
	setDownloadNowSequence(t, start, end)

	downloadHTTPClientFactory = func(port uint16) (*http.Client, error) {
		return &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader("abcd")),
				}, nil
			}),
		}, nil
	}
	downloadURLFmt = "https://example.com/__down?bytes=%d"

	gotSpeed := bitsPerSec(4, 2)

	result, err := MeasureDownloadSpeed(context.Background(), DownloadConfig{
		Bytes:     4,
		Timeout:   time.Second,
		MinSpeed:  gotSpeed + 1,
		ProxyPort: 1080,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "download speed") || !strings.Contains(err.Error(), "below minimum") {
		t.Fatalf("error = %q, want below-minimum message", err.Error())
	}
	if !result.BelowMinimum() {
		t.Fatalf("BelowMinimum() = false, want true; Speed=%v MinSpeed=%v", result.Speed, result.MinSpeed)
	}
	if result.Speed != gotSpeed {
		t.Fatalf("Speed = %v, want %v", result.Speed, gotSpeed)
	}
}
