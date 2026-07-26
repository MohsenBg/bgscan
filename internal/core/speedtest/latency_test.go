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

func restoreLatencyDeps() func() {
	oldFactory := httpClientFactory
	oldURL := latencyProbeURL

	return func() {
		httpClientFactory = oldFactory
		latencyProbeURL = oldURL
	}
}

func TestLatencyResult_String(t *testing.T) {
	r := LatencyResult{RTT: 42 * time.Millisecond}
	if got, want := r.String(), "42ms"; got != want {
		t.Fatalf("String() = %q, want %q", got, want)
	}
}

func TestLatencyResult_AboveMaximum(t *testing.T) {
	tests := []struct {
		name string
		r    LatencyResult
		want bool
	}{
		{
			name: "zero max latency",
			r: LatencyResult{
				RTT:        100 * time.Millisecond,
				MaxLatency: 0,
			},
			want: false,
		},
		{
			name: "below maximum",
			r: LatencyResult{
				RTT:        50 * time.Millisecond,
				MaxLatency: 100 * time.Millisecond,
			},
			want: false,
		},
		{
			name: "equal maximum",
			r: LatencyResult{
				RTT:        100 * time.Millisecond,
				MaxLatency: 100 * time.Millisecond,
			},
			want: false,
		},
		{
			name: "above maximum",
			r: LatencyResult{
				RTT:        101 * time.Millisecond,
				MaxLatency: 100 * time.Millisecond,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.AboveMaximum(); got != tt.want {
				t.Fatalf("AboveMaximum() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMeasureLatency_SetupFailed(t *testing.T) {
	defer restoreLatencyDeps()()

	httpClientFactory = func(port uint16) (*http.Client, error) {
		if port != 1080 {
			t.Fatalf("port = %d, want 1080", port)
		}
		return nil, errors.New("factory boom")
	}

	_, err := MeasureLatency(context.Background(), LatencyConfig{
		Timeout:   time.Second,
		ProxyPort: 1080,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "latency probe setup failed") {
		t.Fatalf("error = %q, want to contain %q", err.Error(), "latency probe setup failed")
	}
	if !strings.Contains(err.Error(), "factory boom") {
		t.Fatalf("error = %q, want to contain %q", err.Error(), "factory boom")
	}
}

func TestMeasureLatency_RequestBuildFailed(t *testing.T) {
	defer restoreLatencyDeps()()

	httpClientFactory = func(port uint16) (*http.Client, error) {
		return &http.Client{}, nil
	}
	latencyProbeURL = "://bad-url"

	_, err := MeasureLatency(context.Background(), LatencyConfig{
		Timeout:   time.Second,
		ProxyPort: 1080,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "latency probe request build failed") {
		t.Fatalf("error = %q, want to contain %q", err.Error(), "latency probe request build failed")
	}
}

func TestMeasureLatency_RequestFailed(t *testing.T) {
	defer restoreLatencyDeps()()

	httpClientFactory = func(port uint16) (*http.Client, error) {
		if port != 1080 {
			t.Fatalf("port = %d, want 1080", port)
		}

		return &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if req.Method != http.MethodGet {
					t.Fatalf("method = %s, want %s", req.Method, http.MethodGet)
				}
				if req.URL.String() != "https://example.com/trace" {
					t.Fatalf("url = %q, want %q", req.URL.String(), "https://example.com/trace")
				}
				return nil, errors.New("network down")
			}),
		}, nil
	}
	latencyProbeURL = "https://example.com/trace"

	_, err := MeasureLatency(context.Background(), LatencyConfig{
		Timeout:   time.Second,
		ProxyPort: 1080,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "latency probe failed (proxy port 1080)") {
		t.Fatalf("error = %q, want to contain %q", err.Error(), "latency probe failed (proxy port 1080)")
	}
	if !strings.Contains(err.Error(), "network down") {
		t.Fatalf("error = %q, want to contain %q", err.Error(), "network down")
	}
}

func TestMeasureLatency_Timeout(t *testing.T) {
	defer restoreLatencyDeps()()

	httpClientFactory = func(port uint16) (*http.Client, error) {
		return &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				<-req.Context().Done()
				return nil, req.Context().Err()
			}),
		}, nil
	}
	latencyProbeURL = "https://example.com/trace"

	_, err := MeasureLatency(context.Background(), LatencyConfig{
		Timeout:   20 * time.Millisecond,
		ProxyPort: 1080,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "latency probe timed out after") {
		t.Fatalf("error = %q, want timeout message", err.Error())
	}
}

func TestMeasureLatency_Success(t *testing.T) {
	defer restoreLatencyDeps()()

	httpClientFactory = func(port uint16) (*http.Client, error) {
		if port != 1080 {
			t.Fatalf("port = %d, want 1080", port)
		}

		return &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if req.Method != http.MethodGet {
					t.Fatalf("method = %s, want %s", req.Method, http.MethodGet)
				}
				if req.URL.String() != "https://example.com/trace" {
					t.Fatalf("url = %q, want %q", req.URL.String(), "https://example.com/trace")
				}

				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader("ok")),
				}, nil
			}),
		}, nil
	}
	latencyProbeURL = "https://example.com/trace"

	result, err := MeasureLatency(context.Background(), LatencyConfig{
		Timeout:   time.Second,
		ProxyPort: 1080,
	})
	if err != nil {
		t.Fatalf("MeasureLatency() unexpected error = %v", err)
	}

	if result.RTT < 0 {
		t.Fatalf("RTT = %v, want non-negative", result.RTT)
	}
	if result.MaxLatency != 0 {
		t.Fatalf("MaxLatency = %v, want 0", result.MaxLatency)
	}
}

func TestMeasureLatency_AboveMaximum(t *testing.T) {
	defer restoreLatencyDeps()()

	httpClientFactory = func(port uint16) (*http.Client, error) {
		return &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				time.Sleep(20 * time.Millisecond)

				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader("ok")),
				}, nil
			}),
		}, nil
	}
	latencyProbeURL = "https://example.com/trace"

	result, err := MeasureLatency(context.Background(), LatencyConfig{
		Timeout:    time.Second,
		MaxLatency: 1 * time.Millisecond,
		ProxyPort:  1080,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "exceeds maximum") {
		t.Fatalf("error = %q, want exceeds maximum message", err.Error())
	}
	if !result.AboveMaximum() {
		t.Fatalf("AboveMaximum() = false, want true; RTT=%v MaxLatency=%v", result.RTT, result.MaxLatency)
	}
}
