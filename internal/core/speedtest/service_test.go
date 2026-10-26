package speedtest

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"
)

func newTestService(
	factory func(httpClientConfig) (*http.Client, error),
	now func() time.Time,
) *service {
	if now == nil {
		now = time.Now
	}

	return &service{
		downloadURLFmt: "https://example.com/__down?bytes=%d",
		uploadURL:      "https://example.com/__up",
		now:            now,
		clientFactory:  factory,
	}
}

func sequenceClock(t *testing.T, ts ...time.Time) func() time.Time {
	t.Helper()

	var i int
	return func() time.Time {
		if i >= len(ts) {
			t.Fatalf("clock called more than expected; calls=%d", i+1)
		}
		value := ts[i]
		i++
		return value
	}
}

func requireErrorContains(t *testing.T, err error, substrings ...string) {
	t.Helper()

	if err == nil {
		t.Fatal("expected an error, got nil")
	}

	for _, want := range substrings {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error = %q, want it to contain %q", err, want)
		}
	}
}

func TestMeasureDownloadSpeed_SetupFailed(t *testing.T) {
	svc := newTestService(func(cfg httpClientConfig) (*http.Client, error) {
		if cfg.ProxyPort != 1080 {
			t.Fatalf("ProxyPort = %d, want 1080", cfg.ProxyPort)
		}
		if cfg.DialContext != nil {
			t.Fatal("DialContext is not nil, want nil")
		}
		return nil, errors.New("factory boom")
	}, nil)

	_, err := svc.MeasureDownloadSpeed(context.Background(), DownloadConfig{
		Bytes:     1024,
		Timeout:   time.Second,
		ProxyPort: 1080,
	})
	requireErrorContains(t, err, "download probe setup failed", "factory boom")
}

func TestMeasureDownloadSpeed_PassesDialContext(t *testing.T) {
	dialContext := func(context.Context, string, string) (net.Conn, error) {
		return nil, errors.New("dial boom")
	}

	svc := newTestService(func(cfg httpClientConfig) (*http.Client, error) {
		if cfg.DialContext == nil {
			t.Fatal("DialContext is nil, want the configured dialer")
		}
		if cfg.ProxyPort != 0 {
			t.Fatalf("ProxyPort = %d, want 0", cfg.ProxyPort)
		}
		return nil, errors.New("factory boom")
	}, nil)

	_, err := svc.MeasureDownloadSpeed(context.Background(), DownloadConfig{
		Bytes:       1024,
		DialContext: dialContext,
	})
	requireErrorContains(t, err, "download probe setup failed", "factory boom")
}

func TestMeasureDownloadSpeed_RequestBuildFailed(t *testing.T) {
	svc := newTestService(func(httpClientConfig) (*http.Client, error) {
		return &http.Client{}, nil
	}, nil)
	svc.downloadURLFmt = "://bad-url-%d"

	_, err := svc.MeasureDownloadSpeed(context.Background(), DownloadConfig{
		Bytes:   1024,
		Timeout: time.Second,
	})
	requireErrorContains(t, err, "download probe request build failed")
}

func TestMeasureDownloadSpeed_RequestFailed(t *testing.T) {
	svc := newTestService(func(cfg httpClientConfig) (*http.Client, error) {
		if cfg.ProxyPort != 1080 {
			t.Fatalf("ProxyPort = %d, want 1080", cfg.ProxyPort)
		}
		return &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if req.Method != http.MethodGet {
					t.Fatalf("method = %s, want %s", req.Method, http.MethodGet)
				}
				if got, want := req.URL.String(), "https://example.com/__down?bytes=2048"; got != want {
					t.Fatalf("url = %q, want %q", got, want)
				}
				return nil, errors.New("network down")
			}),
		}, nil
	}, nil)

	_, err := svc.MeasureDownloadSpeed(context.Background(), DownloadConfig{
		Bytes:     2048,
		Timeout:   time.Second,
		ProxyPort: 1080,
	})
	requireErrorContains(t, err, "download probe failed", "network down")
}

func TestMeasureDownloadSpeed_BodyReadFailed(t *testing.T) {
	svc := newTestService(func(httpClientConfig) (*http.Client, error) {
		return &http.Client{
			Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body: fakeReadCloser{
						readFn: func([]byte) (int, error) {
							return 0, errors.New("read boom")
						},
					},
				}, nil
			}),
		}, nil
	}, nil)

	_, err := svc.MeasureDownloadSpeed(context.Background(), DownloadConfig{
		Bytes:     512,
		Timeout:   time.Second,
		ProxyPort: 1080,
	})
	requireErrorContains(t, err, "download probe body read failed", "read boom")
}

func TestMeasureDownloadSpeed_CanceledDuringSetup(t *testing.T) {
	svc := newTestService(func(httpClientConfig) (*http.Client, error) {
		return &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				<-req.Context().Done()
				return nil, req.Context().Err()
			}),
		}, nil
	}, nil)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(25 * time.Millisecond)
		cancel()
	}()
	defer cancel()

	_, err := svc.MeasureDownloadSpeed(ctx, DownloadConfig{
		Bytes:     1024,
		ProxyPort: 1080,
	})

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
}

func TestMeasureDownloadSpeed_TimeoutInterruptsStalledBody(t *testing.T) {
	body := newBlockingReadCloser()
	svc := newTestService(func(httpClientConfig) (*http.Client, error) {
		return &http.Client{
			Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       body,
				}, nil
			}),
		}, nil
	}, nil)

	start := time.Now()
	_, err := svc.MeasureDownloadSpeed(context.Background(), DownloadConfig{
		Bytes:     1024,
		Timeout:   25 * time.Millisecond,
		ProxyPort: 1080,
	})

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("error = %v, want context.DeadlineExceeded", err)
	}
	requireErrorContains(t, err, "download probe timed out after")

	if elapsed := time.Since(start); elapsed > 2*time.Second {
		t.Fatalf("timeout took %v; body read was not interrupted", elapsed)
	}
}

func TestMeasureDownloadSpeed_NoData(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	svc := newTestService(func(httpClientConfig) (*http.Client, error) {
		return &http.Client{
			Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader("")),
				}, nil
			}),
		}, nil
	}, sequenceClock(t, start, start.Add(time.Second)))

	_, err := svc.MeasureDownloadSpeed(context.Background(), DownloadConfig{
		Bytes:     1024,
		Timeout:   time.Second,
		ProxyPort: 1080,
	})
	requireErrorContains(t, err, "download probe returned no data")
}

func TestMeasureDownloadSpeed_Success(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	svc := newTestService(func(cfg httpClientConfig) (*http.Client, error) {
		if cfg.ProxyPort != 1080 {
			t.Fatalf("ProxyPort = %d, want 1080", cfg.ProxyPort)
		}
		if cfg.DialContext != nil {
			t.Fatal("DialContext is not nil, want nil")
		}
		return &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if got, want := req.URL.String(), "https://example.com/__down?bytes=4"; got != want {
					t.Fatalf("url = %q, want %q", got, want)
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader("abcd")),
				}, nil
			}),
		}, nil
	}, sequenceClock(t, start, start.Add(2*time.Second)))

	result, err := svc.MeasureDownloadSpeed(context.Background(), DownloadConfig{
		Bytes:     4,
		Timeout:   time.Second,
		ProxyPort: 1080,
	})
	if err != nil {
		t.Fatalf("MeasureDownloadSpeed() error = %v", err)
	}

	if got, want := result.Bytes, uint64(4); got != want {
		t.Fatalf("Bytes = %d, want %d", got, want)
	}
	if got, want := result.Speed, bitsPerSec(4, 2); got != want {
		t.Fatalf("Speed = %v, want %v", got, want)
	}
}

func TestMeasureDownloadSpeed_BelowMinimum(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	svc := newTestService(func(httpClientConfig) (*http.Client, error) {
		return &http.Client{
			Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader("abcd")),
				}, nil
			}),
		}, nil
	}, sequenceClock(t, start, start.Add(2*time.Second)))

	wantSpeed := bitsPerSec(4, 2)
	result, err := svc.MeasureDownloadSpeed(context.Background(), DownloadConfig{
		Bytes:     4,
		Timeout:   time.Second,
		MinSpeed:  wantSpeed + 1,
		ProxyPort: 1080,
	})

	if !errors.Is(err, ErrBelowMinimum) {
		t.Fatalf("error = %v, want ErrBelowMinimum", err)
	}
	if !result.BelowMinimum() {
		t.Fatal("BelowMinimum() = false, want true")
	}
	if result.Speed != wantSpeed {
		t.Fatalf("Speed = %v, want %v", result.Speed, wantSpeed)
	}
}

func TestMeasureDownloadSpeed_RejectsNonPositiveBytes(t *testing.T) {
	for _, bytes := range []int64{0, -1} {
		t.Run("bytes", func(t *testing.T) {
			svc := newTestService(func(httpClientConfig) (*http.Client, error) {
				t.Fatal("client factory must not be called")
				return nil, nil
			}, nil)

			_, err := svc.MeasureDownloadSpeed(context.Background(), DownloadConfig{
				Bytes: bytes,
			})
			requireErrorContains(t, err, "requires Bytes > 0")
		})
	}
}

func TestMeasureUploadSpeed_SetupFailed(t *testing.T) {
	svc := newTestService(func(cfg httpClientConfig) (*http.Client, error) {
		if cfg.ProxyPort != 1080 {
			t.Fatalf("ProxyPort = %d, want 1080", cfg.ProxyPort)
		}
		return nil, errors.New("factory boom")
	}, nil)

	_, err := svc.MeasureUploadSpeed(context.Background(), UploadConfig{
		Bytes:     1024,
		Timeout:   time.Second,
		ProxyPort: 1080,
	})
	requireErrorContains(t, err, "upload probe setup failed", "factory boom")
}

func TestMeasureUploadSpeed_PassesDialContext(t *testing.T) {
	dialContext := func(context.Context, string, string) (net.Conn, error) {
		return nil, errors.New("dial boom")
	}

	svc := newTestService(func(cfg httpClientConfig) (*http.Client, error) {
		if cfg.DialContext == nil {
			t.Fatal("DialContext is nil, want the configured dialer")
		}
		if cfg.ProxyPort != 0 {
			t.Fatalf("ProxyPort = %d, want 0", cfg.ProxyPort)
		}
		return nil, errors.New("factory boom")
	}, nil)

	_, err := svc.MeasureUploadSpeed(context.Background(), UploadConfig{
		Bytes:       1024,
		DialContext: dialContext,
	})
	requireErrorContains(t, err, "upload probe setup failed", "factory boom")
}

func TestMeasureUploadSpeed_RequestBuildFailed(t *testing.T) {
	svc := newTestService(func(httpClientConfig) (*http.Client, error) {
		return &http.Client{}, nil
	}, nil)
	svc.uploadURL = "://bad-url"

	_, err := svc.MeasureUploadSpeed(context.Background(), UploadConfig{
		Bytes:     1024,
		Timeout:   time.Second,
		ProxyPort: 1080,
	})
	requireErrorContains(t, err, "upload probe request build failed")
}

func TestMeasureUploadSpeed_RequestFailed(t *testing.T) {
	svc := newTestService(func(httpClientConfig) (*http.Client, error) {
		return &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if req.Method != http.MethodPost {
					t.Fatalf("method = %s, want %s", req.Method, http.MethodPost)
				}
				if got, want := req.URL.String(), "https://example.com/__up"; got != want {
					t.Fatalf("url = %q, want %q", got, want)
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
	}, nil)

	_, err := svc.MeasureUploadSpeed(context.Background(), UploadConfig{
		Bytes:     2048,
		Timeout:   time.Second,
		ProxyPort: 1080,
	})
	requireErrorContains(t, err, "upload probe failed", "network down")
}

func TestMeasureUploadSpeed_TimeoutInterruptsRequest(t *testing.T) {
	svc := newTestService(func(httpClientConfig) (*http.Client, error) {
		return &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				<-req.Context().Done()
				return nil, req.Context().Err()
			}),
		}, nil
	}, nil)

	_, err := svc.MeasureUploadSpeed(context.Background(), UploadConfig{
		Bytes:     1024,
		Timeout:   25 * time.Millisecond,
		ProxyPort: 1080,
	})

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("error = %v, want context.DeadlineExceeded", err)
	}
	requireErrorContains(t, err, "upload probe timed out after")
}

func TestMeasureUploadSpeed_ResponseReadFailed(t *testing.T) {
	svc := newTestService(func(httpClientConfig) (*http.Client, error) {
		return &http.Client{
			Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body: fakeReadCloser{
						readFn: func([]byte) (int, error) {
							return 0, errors.New("read boom")
						},
					},
				}, nil
			}),
		}, nil
	}, nil)

	_, err := svc.MeasureUploadSpeed(context.Background(), UploadConfig{
		Bytes:     1024,
		Timeout:   time.Second,
		ProxyPort: 1080,
	})
	requireErrorContains(t, err, "upload probe response read failed", "read boom")
}

func TestMeasureUploadSpeed_TimeoutInterruptsStalledBody(t *testing.T) {
	body := newBlockingReadCloser()
	svc := newTestService(func(httpClientConfig) (*http.Client, error) {
		return &http.Client{
			Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       body,
				}, nil
			}),
		}, nil
	}, nil)

	_, err := svc.MeasureUploadSpeed(context.Background(), UploadConfig{
		Bytes:     1024,
		Timeout:   25 * time.Millisecond,
		ProxyPort: 1080,
	})

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("error = %v, want context.DeadlineExceeded", err)
	}
	requireErrorContains(t, err, "upload probe timed out after")
}

func TestMeasureUploadSpeed_Success(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	svc := newTestService(func(cfg httpClientConfig) (*http.Client, error) {
		if cfg.ProxyPort != 1080 {
			t.Fatalf("ProxyPort = %d, want 1080", cfg.ProxyPort)
		}
		return &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				body, err := io.ReadAll(req.Body)
				if err != nil {
					t.Fatalf("reading request body: %v", err)
				}
				if !bytes.Equal(body, make([]byte, 4)) {
					t.Fatal("uploaded body contains unexpected data")
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader("ok")),
				}, nil
			}),
		}, nil
	}, sequenceClock(t, start, start.Add(2*time.Second)))

	result, err := svc.MeasureUploadSpeed(context.Background(), UploadConfig{
		Bytes:     4,
		Timeout:   time.Second,
		ProxyPort: 1080,
	})
	if err != nil {
		t.Fatalf("MeasureUploadSpeed() error = %v", err)
	}

	if got, want := result.Bytes, uint64(4); got != want {
		t.Fatalf("Bytes = %d, want %d", got, want)
	}
	if got, want := result.Speed, bitsPerSec(4, 2); got != want {
		t.Fatalf("Speed = %v, want %v", got, want)
	}
}

func TestMeasureUploadSpeed_BelowMinimum(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	svc := newTestService(func(httpClientConfig) (*http.Client, error) {
		return &http.Client{
			Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader("ok")),
				}, nil
			}),
		}, nil
	}, sequenceClock(t, start, start.Add(2*time.Second)))

	wantSpeed := bitsPerSec(4, 2)
	result, err := svc.MeasureUploadSpeed(context.Background(), UploadConfig{
		Bytes:     4,
		Timeout:   time.Second,
		MinSpeed:  wantSpeed + 1,
		ProxyPort: 1080,
	})

	if !errors.Is(err, ErrBelowMinimum) {
		t.Fatalf("error = %v, want ErrBelowMinimum", err)
	}
	if !result.BelowMinimum() {
		t.Fatal("BelowMinimum() = false, want true")
	}
	if result.Speed != wantSpeed {
		t.Fatalf("Speed = %v, want %v", result.Speed, wantSpeed)
	}
}

func TestMeasureUploadSpeed_RejectsNonPositiveBytes(t *testing.T) {
	for _, bytes := range []int64{0, -1} {
		t.Run("bytes", func(t *testing.T) {
			svc := newTestService(func(httpClientConfig) (*http.Client, error) {
				t.Fatal("client factory must not be called")
				return nil, nil
			}, nil)

			_, err := svc.MeasureUploadSpeed(context.Background(), UploadConfig{
				Bytes: bytes,
			})
			requireErrorContains(t, err, "requires Bytes > 0")
		})
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
		{name: "no maximum", r: LatencyResult{RTT: 100 * time.Millisecond}, want: false},
		{name: "below", r: LatencyResult{RTT: 50 * time.Millisecond, MaxLatency: 100 * time.Millisecond}, want: false},
		{name: "equal", r: LatencyResult{RTT: 100 * time.Millisecond, MaxLatency: 100 * time.Millisecond}, want: false},
		{name: "above", r: LatencyResult{RTT: 101 * time.Millisecond, MaxLatency: 100 * time.Millisecond}, want: true},
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
	svc := newTestService(func(cfg httpClientConfig) (*http.Client, error) {
		if cfg.ProxyPort != 1080 {
			t.Fatalf("ProxyPort = %d, want 1080", cfg.ProxyPort)
		}
		return nil, errors.New("factory boom")
	}, nil)

	_, err := svc.MeasureLatency(context.Background(), LatencyConfig{
		Timeout:   time.Second,
		ProxyPort: 1080,
	})
	requireErrorContains(t, err, "latency probe setup failed", "factory boom")
}

func TestMeasureLatency_RequestBuildFailed(t *testing.T) {
	svc := newTestService(func(httpClientConfig) (*http.Client, error) {
		return &http.Client{}, nil
	}, nil)

	_, err := svc.MeasureLatency(context.Background(), LatencyConfig{
		Timeout: time.Second,
		URL:     "://bad-url",
	})
	requireErrorContains(t, err, "latency probe request build failed")
}

func TestMeasureLatency_RequestFailed(t *testing.T) {
	svc := newTestService(func(httpClientConfig) (*http.Client, error) {
		return &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if got, want := req.URL.String(), "https://example.com/trace"; got != want {
					t.Fatalf("url = %q, want %q", got, want)
				}
				return nil, errors.New("network down")
			}),
		}, nil
	}, nil)

	_, err := svc.MeasureLatency(context.Background(), LatencyConfig{
		Timeout:   time.Second,
		ProxyPort: 1080,
		URL:       "https://example.com/trace",
	})
	requireErrorContains(t, err, "latency probe failed", "network down")
}

func TestMeasureLatency_Timeout(t *testing.T) {
	svc := newTestService(func(httpClientConfig) (*http.Client, error) {
		return &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				<-req.Context().Done()
				return nil, req.Context().Err()
			}),
		}, nil
	}, nil)

	_, err := svc.MeasureLatency(context.Background(), LatencyConfig{
		Timeout: 25 * time.Millisecond,
		URL:     "https://example.com/trace",
	})

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("error = %v, want context.DeadlineExceeded", err)
	}
	requireErrorContains(t, err, "latency probe timed out after")
}

func TestMeasureLatency_Success(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	svc := newTestService(func(httpClientConfig) (*http.Client, error) {
		return &http.Client{
			Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader("ok")),
				}, nil
			}),
		}, nil
	}, sequenceClock(t, start, start.Add(42*time.Millisecond)))

	result, err := svc.MeasureLatency(context.Background(), LatencyConfig{
		Timeout: time.Second,
		URL:     "https://example.com/trace",
	})
	if err != nil {
		t.Fatalf("MeasureLatency() error = %v", err)
	}

	if got, want := result.RTT, 42*time.Millisecond; got != want {
		t.Fatalf("RTT = %v, want %v", got, want)
	}
}

func TestMeasureLatency_AcceptsNoContent(t *testing.T) {
	svc := newTestService(func(httpClientConfig) (*http.Client, error) {
		return &http.Client{
			Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusNoContent,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader("")),
				}, nil
			}),
		}, nil
	}, nil)

	if _, err := svc.MeasureLatency(context.Background(), LatencyConfig{
		Timeout: time.Second,
		URL:     GoogleGenerate204HTTPS,
	}); err != nil {
		t.Fatalf("MeasureLatency() error = %v", err)
	}
}

func TestMeasureLatency_UnexpectedStatus(t *testing.T) {
	svc := newTestService(func(httpClientConfig) (*http.Client, error) {
		return &http.Client{
			Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusInternalServerError,
					Status:     "500 Internal Server Error",
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader("")),
				}, nil
			}),
		}, nil
	}, nil)

	_, err := svc.MeasureLatency(context.Background(), LatencyConfig{
		Timeout: time.Second,
		URL:     "https://example.com/trace",
	})
	requireErrorContains(t, err, "latency probe unexpected status")
}

func TestMeasureLatency_AboveMaximum(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	svc := newTestService(func(httpClientConfig) (*http.Client, error) {
		return &http.Client{
			Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader("ok")),
				}, nil
			}),
		}, nil
	}, sequenceClock(t, start, start.Add(100*time.Millisecond)))

	result, err := svc.MeasureLatency(context.Background(), LatencyConfig{
		Timeout:    time.Second,
		MaxLatency: 50 * time.Millisecond,
		URL:        "https://example.com/trace",
	})

	if !errors.Is(err, ErrAboveMaximum) {
		t.Fatalf("error = %v, want ErrAboveMaximum", err)
	}
	if errors.Is(err, context.DeadlineExceeded) {
		t.Fatal("ErrAboveMaximum must not wrap context.DeadlineExceeded")
	}
	if got, want := result.RTT, 100*time.Millisecond; got != want {
		t.Fatalf("RTT = %v, want %v", got, want)
	}
}

func TestMeasureLatency_DefaultURL(t *testing.T) {
	var gotURL string
	svc := newTestService(func(httpClientConfig) (*http.Client, error) {
		return &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				gotURL = req.URL.String()
				return &http.Response{
					StatusCode: http.StatusNoContent,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader("")),
				}, nil
			}),
		}, nil
	}, nil)

	if _, err := svc.MeasureLatency(context.Background(), LatencyConfig{
		Timeout: time.Second,
	}); err != nil {
		t.Fatalf("MeasureLatency() error = %v", err)
	}

	if gotURL != GoogleGenerate204HTTP {
		t.Fatalf("URL = %q, want %q", gotURL, GoogleGenerate204HTTP)
	}
}

func TestMeasureLatency_CustomDialContext(t *testing.T) {
	var called bool
	dialContext := func(context.Context, string, string) (net.Conn, error) {
		called = true
		return nil, errors.New("dial boom")
	}

	svc := newTestService(newHTTPClient, nil)

	_, err := svc.MeasureLatency(context.Background(), LatencyConfig{
		Timeout:     time.Second,
		URL:         "https://example.com/trace",
		DialContext: dialContext,
	})
	if !called {
		t.Fatal("DialContext was not called")
	}
	requireErrorContains(t, err, "latency probe failed", "dial boom")
}
