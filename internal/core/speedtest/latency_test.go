package speedtest

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"
)

func restoreLatencyDeps() func() {
	oldFactory := latencyHTTPClientFactory

	return func() {
		latencyHTTPClientFactory = oldFactory
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
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.AboveMaximum(); got != tt.want {
				t.Fatalf("AboveMaximum() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMeasureLatency_SetupFailed(t *testing.T) {
	defer restoreLatencyDeps()()

	latencyHTTPClientFactory = func(cfg httpClientConfig) (*http.Client, error) {
		if cfg.ProxyPort != 1080 {
			t.Fatalf("ProxyPort = %d, want 1080", cfg.ProxyPort)
		}

		return nil, errors.New("factory boom")
	}

	_, err := measureLatency(context.Background(), LatencyConfig{
		Timeout:   time.Second,
		ProxyPort: 1080,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "latency probe setup failed") {
		t.Fatalf("error = %q", err.Error())
	}
}

func TestMeasureLatency_DialContext(t *testing.T) {
	defer restoreLatencyDeps()()

	dialContext := func(
		ctx context.Context,
		network string,
		address string,
	) (net.Conn, error) {
		return nil, errors.New("dial boom")
	}

	latencyHTTPClientFactory = func(cfg httpClientConfig) (*http.Client, error) {
		if cfg.DialContext == nil {
			t.Fatal("DialContext is nil")
		}

		return nil, errors.New("factory boom")
	}

	_, err := measureLatency(context.Background(), LatencyConfig{
		Timeout:     time.Second,
		DialContext: dialContext,
	})

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMeasureLatency_RequestBuildFailed(t *testing.T) {
	defer restoreLatencyDeps()()

	latencyHTTPClientFactory = func(cfg httpClientConfig) (*http.Client, error) {
		return &http.Client{}, nil
	}

	_, err := measureLatency(context.Background(), LatencyConfig{
		Timeout: time.Second,
		URL:     "://bad-url",
	})

	if err == nil {
		t.Fatal("expected error")
	}

	if !strings.Contains(err.Error(), "latency probe request build failed") {
		t.Fatalf("error = %q", err.Error())
	}
}

func TestMeasureLatency_RequestFailed(t *testing.T) {
	defer restoreLatencyDeps()()

	latencyHTTPClientFactory = func(cfg httpClientConfig) (*http.Client, error) {
		return &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if req.URL.String() != "https://example.com/trace" {
					t.Fatalf("url = %q", req.URL.String())
				}

				return nil, errors.New("network down")
			}),
		}, nil
	}

	_, err := measureLatency(context.Background(), LatencyConfig{
		Timeout:   time.Second,
		ProxyPort: 1080,
		URL:       "https://example.com/trace",
	})

	if err == nil {
		t.Fatal("expected error")
	}

	if !strings.Contains(err.Error(), "network down") {
		t.Fatalf("error = %q", err.Error())
	}
}

func TestMeasureLatency_Timeout(t *testing.T) {
	defer restoreLatencyDeps()()

	latencyHTTPClientFactory = func(cfg httpClientConfig) (*http.Client, error) {
		return &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				<-req.Context().Done()
				return nil, req.Context().Err()
			}),
		}, nil
	}

	_, err := measureLatency(context.Background(), LatencyConfig{
		Timeout: 20 * time.Millisecond,
		URL:     "https://example.com/trace",
	})

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMeasureLatency_Success(t *testing.T) {
	defer restoreLatencyDeps()()

	latencyHTTPClientFactory = func(cfg httpClientConfig) (*http.Client, error) {
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

	result, err := measureLatency(context.Background(), LatencyConfig{
		Timeout: time.Second,
		URL:     "https://example.com/trace",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.RTT < 0 {
		t.Fatalf("RTT = %v", result.RTT)
	}
}

func TestMeasureLatency_Google204(t *testing.T) {
	defer restoreLatencyDeps()()

	latencyHTTPClientFactory = func(cfg httpClientConfig) (*http.Client, error) {
		return &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusNoContent,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader("")),
				}, nil
			}),
		}, nil
	}

	_, err := measureLatency(context.Background(), LatencyConfig{
		URL: GoogleGenerate204HTTPS,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
