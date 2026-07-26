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

func restoreUploadDeps() func() {
	oldFactory := uploadHTTPClientFactory
	oldURL := uploadURL
	oldNow := uploadNow

	return func() {
		uploadHTTPClientFactory = oldFactory
		uploadURL = oldURL
		uploadNow = oldNow
	}
}

func setUploadNowSequence(t *testing.T, ts ...time.Time) {
	t.Helper()

	var i int
	uploadNow = func() time.Time {
		if i >= len(ts) {
			t.Fatalf("uploadNow called more than expected; calls=%d", i+1)
		}
		v := ts[i]
		i++
		return v
	}
}

func TestMeasureUploadSpeed_SetupFailed(t *testing.T) {
	defer restoreUploadDeps()()

	uploadHTTPClientFactory = func(port uint16) (*http.Client, error) {
		if port != 1080 {
			t.Fatalf("port = %d, want 1080", port)
		}
		return nil, errors.New("factory boom")
	}

	_, err := MeasureUploadSpeed(context.Background(), UploadConfig{
		Bytes:     1024,
		Timeout:   time.Second,
		ProxyPort: 1080,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "upload probe setup failed") {
		t.Fatalf("error = %q, want setup failure", err.Error())
	}
	if !strings.Contains(err.Error(), "factory boom") {
		t.Fatalf("error = %q, want wrapped factory error", err.Error())
	}
}

func TestMeasureUploadSpeed_RequestBuildFailed(t *testing.T) {
	defer restoreUploadDeps()()

	uploadHTTPClientFactory = func(port uint16) (*http.Client, error) {
		return &http.Client{}, nil
	}
	uploadURL = "://bad-url"

	_, err := MeasureUploadSpeed(context.Background(), UploadConfig{
		Bytes:     1024,
		Timeout:   time.Second,
		ProxyPort: 1080,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "upload probe request build failed") {
		t.Fatalf("error = %q, want request build failure", err.Error())
	}
}

func TestMeasureUploadSpeed_RequestFailed(t *testing.T) {
	defer restoreUploadDeps()()

	uploadHTTPClientFactory = func(port uint16) (*http.Client, error) {
		if port != 1080 {
			t.Fatalf("port = %d, want 1080", port)
		}

		return &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if req.Method != http.MethodPost {
					t.Fatalf("method = %s, want %s", req.Method, http.MethodPost)
				}
				if req.URL.String() != "https://example.com/__up" {
					t.Fatalf("url = %q, want %q", req.URL.String(), "https://example.com/__up")
				}
				if got, want := req.ContentLength, int64(2048); got != want {
					t.Fatalf("ContentLength = %d, want %d", got, want)
				}
				if got := req.Header.Get("Content-Type"); got != "application/octet-stream" {
					t.Fatalf("Content-Type = %q, want application/octet-stream", got)
				}
				return nil, errors.New("network down")
			}),
		}, nil
	}
	uploadURL = "https://example.com/__up"

	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.Add(time.Second)
	setUploadNowSequence(t, start, end)

	_, err := MeasureUploadSpeed(context.Background(), UploadConfig{
		Bytes:     2048,
		Timeout:   time.Second,
		ProxyPort: 1080,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "upload probe failed (proxy port 1080)") {
		t.Fatalf("error = %q, want request failure", err.Error())
	}
	if !strings.Contains(err.Error(), "network down") {
		t.Fatalf("error = %q, want wrapped transport error", err.Error())
	}
}

func TestMeasureUploadSpeed_TimeoutDuringDo(t *testing.T) {
	defer restoreUploadDeps()()

	uploadHTTPClientFactory = func(port uint16) (*http.Client, error) {
		return &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				<-req.Context().Done()
				return nil, req.Context().Err()
			}),
		}, nil
	}
	uploadURL = "https://example.com/__up"

	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.Add(2 * time.Second)
	setUploadNowSequence(t, start, end)

	_, err := MeasureUploadSpeed(context.Background(), UploadConfig{
		Bytes:     1024,
		Timeout:   time.Nanosecond,
		ProxyPort: 1080,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "upload probe timed out after") {
		t.Fatalf("error = %q, want timeout message", err.Error())
	}
}

func TestMeasureUploadSpeed_ResponseReadFailed(t *testing.T) {
	defer restoreUploadDeps()()

	uploadHTTPClientFactory = func(port uint16) (*http.Client, error) {
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
	uploadURL = "https://example.com/__up"

	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.Add(time.Second)
	setUploadNowSequence(t, start, end)

	_, err := MeasureUploadSpeed(context.Background(), UploadConfig{
		Bytes:     1024,
		Timeout:   time.Second,
		ProxyPort: 1080,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "upload probe response read failed") {
		t.Fatalf("error = %q, want response read failure", err.Error())
	}
	if !strings.Contains(err.Error(), "read boom") {
		t.Fatalf("error = %q, want wrapped read error", err.Error())
	}
}

func TestMeasureUploadSpeed_NoData(t *testing.T) {
	defer restoreUploadDeps()()

	uploadHTTPClientFactory = func(port uint16) (*http.Client, error) {
		return &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader("ok")),
				}, nil
			}),
		}, nil
	}
	uploadURL = "https://example.com/__up"

	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.Add(time.Second)
	setUploadNowSequence(t, start, end)

	_, err := MeasureUploadSpeed(context.Background(), UploadConfig{
		Bytes:     0,
		Timeout:   time.Second,
		ProxyPort: 1080,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "upload probe sent no data") {
		t.Fatalf("error = %q, want no data failure", err.Error())
	}
}

func TestMeasureUploadSpeed_Success(t *testing.T) {
	defer restoreUploadDeps()()

	uploadHTTPClientFactory = func(port uint16) (*http.Client, error) {
		if port != 1080 {
			t.Fatalf("port = %d, want 1080", port)
		}

		return &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if req.Method != http.MethodPost {
					t.Fatalf("method = %s, want %s", req.Method, http.MethodPost)
				}
				if req.URL.String() != "https://example.com/__up" {
					t.Fatalf("url = %q, want %q", req.URL.String(), "https://example.com/__up")
				}
				if got, want := req.ContentLength, int64(4); got != want {
					t.Fatalf("ContentLength = %d, want %d", got, want)
				}

				body, err := io.ReadAll(req.Body)
				if err != nil {
					t.Fatalf("reading request body: %v", err)
				}
				if got, want := len(body), 4; got != want {
					t.Fatalf("uploaded bytes = %d, want %d", got, want)
				}

				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader("ok")),
				}, nil
			}),
		}, nil
	}
	uploadURL = "https://example.com/__up"

	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.Add(2 * time.Second)
	setUploadNowSequence(t, start, end)

	result, err := MeasureUploadSpeed(context.Background(), UploadConfig{
		Bytes:     4,
		Timeout:   time.Second,
		ProxyPort: 1080,
	})
	if err != nil {
		t.Fatalf("MeasureUploadSpeed() unexpected error = %v", err)
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

func TestMeasureUploadSpeed_BelowMinimum(t *testing.T) {
	defer restoreUploadDeps()()

	uploadHTTPClientFactory = func(port uint16) (*http.Client, error) {
		return &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader("ok")),
				}, nil
			}),
		}, nil
	}
	uploadURL = "https://example.com/__up"

	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.Add(2 * time.Second)
	setUploadNowSequence(t, start, end)

	gotSpeed := bitsPerSec(4, 2)

	result, err := MeasureUploadSpeed(context.Background(), UploadConfig{
		Bytes:     4,
		Timeout:   time.Second,
		MinSpeed:  gotSpeed + 1,
		ProxyPort: 1080,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "upload speed") || !strings.Contains(err.Error(), "below minimum") {
		t.Fatalf("error = %q, want below minimum failure", err.Error())
	}
	if !result.BelowMinimum() {
		t.Fatalf("BelowMinimum() = false, want true")
	}
	if result.Speed != gotSpeed {
		t.Fatalf("Speed = %v, want %v", result.Speed, gotSpeed)
	}
}
