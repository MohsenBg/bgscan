// transport_test.go
package speedtest

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestContextReader_NormalRead(t *testing.T) {
	data := "hello world"
	src := strings.NewReader(data)
	ctx := context.Background()

	cr := &contextReader{ctx: ctx, r: src}
	buf := make([]byte, 20)

	n, err := cr.Read(buf)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if n != len(data) {
		t.Errorf("expected to read %d bytes, got %d", len(data), n)
	}
	if string(buf[:n]) != data {
		t.Errorf("expected content %q, got %q", data, string(buf[:n]))
	}
}

func TestContextReader_CancelledContext(t *testing.T) {
	src := strings.NewReader("some data that won't be read")
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // instantly cancel

	cr := &contextReader{ctx: ctx, r: src}
	buf := make([]byte, 20)

	n, err := cr.Read(buf)
	if err == nil {
		t.Fatal("expected error due to cancelled context, got nil")
	}
	if err != context.Canceled {
		t.Errorf("expected context.Canceled error, got %v", err)
	}
	if n != 0 {
		t.Errorf("expected 0 bytes read, got %d", n)
	}
}

func TestNewHTTPClient_Config(t *testing.T) {
	port := uint16(1080)

	client, err := newHTTPClient(port)
	if err != nil {
		t.Fatalf("newHTTPClient() error = %v", err)
	}
	if client == nil {
		t.Fatal("newHTTPClient() returned nil client")
	}

	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("client.Transport type = %T, want *http.Transport", client.Transport)
	}

	if transport.Proxy == nil {
		t.Fatal("transport.Proxy is nil")
	}

	req := &http.Request{URL: mustParseURL(t, "https://example.com")}
	proxyURL, err := transport.Proxy(req)
	if err != nil {
		t.Fatalf("transport.Proxy() error = %v", err)
	}
	if proxyURL == nil {
		t.Fatal("transport.Proxy() returned nil")
	}

	wantProxy := "socks5://127.0.0.1:1080"
	if got := proxyURL.String(); got != wantProxy {
		t.Errorf("proxy URL = %q, want %q", got, wantProxy)
	}

	if transport.TLSHandshakeTimeout != connectTimeout {
		t.Errorf("TLSHandshakeTimeout = %v, want %v", transport.TLSHandshakeTimeout, connectTimeout)
	}

	if transport.ForceAttemptHTTP2 {
		t.Error("ForceAttemptHTTP2 = true, want false")
	}

	if transport.MaxIdleConns != 10 {
		t.Errorf("MaxIdleConns = %d, want 10", transport.MaxIdleConns)
	}

	if transport.MaxIdleConnsPerHost != 10 {
		t.Errorf("MaxIdleConnsPerHost = %d, want 10", transport.MaxIdleConnsPerHost)
	}

	if transport.MaxConnsPerHost != 10 {
		t.Errorf("MaxConnsPerHost = %d, want 10", transport.MaxConnsPerHost)
	}
}

func mustParseURL(t *testing.T, raw string) *url.URL {
	t.Helper()
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("url.Parse(%q) error = %v", raw, err)
	}
	return u
}
