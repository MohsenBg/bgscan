package dns

import (
	"context"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestParseTunnelEndpoint(t *testing.T) {
	tests := []struct {
		name       string
		resolverIP string
		listenPort uint16
		wantAddr   string
		wantErr    string
	}{
		{
			name:       "valid ipv4",
			resolverIP: "1.1.1.1",
			listenPort: 18080,
			wantAddr:   "1.1.1.1",
		},
		{
			name:       "valid ipv6",
			resolverIP: "2606:4700:4700::1111",
			listenPort: 18080,
			wantAddr:   "2606:4700:4700::1111",
		},
		{
			name:       "surrounding whitespace",
			resolverIP: " 1.1.1.1 ",
			listenPort: 18080,
			wantAddr:   "1.1.1.1",
		},
		{
			name:       "empty resolver",
			resolverIP: "",
			listenPort: 18080,
			wantErr:    "resolver IP is required",
		},
		{
			name:       "whitespace resolver",
			resolverIP: "   ",
			listenPort: 18080,
			wantErr:    "resolver IP is required",
		},
		{
			name:       "zero listen port",
			resolverIP: "1.1.1.1",
			listenPort: 0,
			wantErr:    "listen port must be greater than zero",
		},
		{
			name:       "unparsable resolver",
			resolverIP: "resolver.example.com",
			listenPort: 18080,
			wantErr:    "invalid resolver IP",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, err := parseTunnelEndpoint(tt.resolverIP, tt.listenPort)

			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("parseTunnelEndpoint() error = %v, want %q", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("parseTunnelEndpoint() error = %v", err)
			}

			if addr.String() != tt.wantAddr {
				t.Errorf("addr = %q, want %q", addr.String(), tt.wantAddr)
			}
		})
	}
}

// fakeEmbeddedClient implements embeddedClient for testing the boot
// wait loop without a vendor client.
type fakeEmbeddedClient struct {
	ready   atomic.Bool
	stopped atomic.Bool
	runErr  error
}

func (c *fakeEmbeddedClient) SessionReady() bool {
	return c.ready.Load()
}

func (c *fakeEmbeddedClient) Run(ctx context.Context) error {
	if c.runErr != nil {
		return c.runErr
	}

	<-ctx.Done()
	return nil
}

func (c *fakeEmbeddedClient) StopAsyncRuntime() {
	c.stopped.Store(true)
}

func TestBootEmbeddedSessionReady(t *testing.T) {
	app := &fakeEmbeddedClient{}

	// bootEmbedded polls every readyPollInterval; become ready shortly
	// after boot instead of immediately, so the loop is exercised.
	time.AfterFunc(50*time.Millisecond, func() { app.ready.Store(true) })

	handle, err := bootEmbedded(context.Background(), app)
	if err != nil {
		t.Fatalf("bootEmbedded() error = %v", err)
	}

	if handle == nil {
		t.Fatal("bootEmbedded() returned nil handle")
	}

	// Close must stop the runtime and be safe to call twice.
	if err := handle.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}
	if err := handle.Close(); err != nil {
		t.Errorf("second Close() error = %v, want nil", err)
	}

	if !app.stopped.Load() {
		t.Error("StopAsyncRuntime was not called")
	}
}

func TestBootEmbeddedRunError(t *testing.T) {
	app := &fakeEmbeddedClient{
		runErr: context.Canceled,
	}

	_, err := bootEmbedded(context.Background(), app)
	if err == nil || !strings.Contains(err.Error(), "client runtime exited") {
		t.Fatalf("bootEmbedded() error = %v, want runtime-exited error", err)
	}

	if !app.stopped.Load() {
		t.Error("StopAsyncRuntime was not called on failure")
	}
}

func TestBootEmbeddedRunExitsCleanly(t *testing.T) {
	// A nil runErr still means the runtime stopped before the session
	// was ready.
	app := &exitingFakeClient{}

	_, err := bootEmbedded(context.Background(), app)
	if err == nil || !strings.Contains(err.Error(), "exited before session ready") {
		t.Fatalf("bootEmbedded() error = %v, want exited-before-ready error", err)
	}
}

// exitingFakeClient returns nil from Run immediately.
type exitingFakeClient struct {
	stopped atomic.Bool
}

func (c *exitingFakeClient) SessionReady() bool { return false }

func (c *exitingFakeClient) Run(context.Context) error { return nil }

func (c *exitingFakeClient) StopAsyncRuntime() { c.stopped.Store(true) }

func TestBootEmbeddedContextCanceled(t *testing.T) {
	app := &fakeEmbeddedClient{}

	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(50*time.Millisecond, cancel)

	_, err := bootEmbedded(ctx, app)
	if err == nil || !strings.Contains(err.Error(), "context canceled while waiting for session") {
		t.Fatalf("bootEmbedded() error = %v, want context-canceled error", err)
	}

	if !app.stopped.Load() {
		t.Error("StopAsyncRuntime was not called on cancellation")
	}
}
