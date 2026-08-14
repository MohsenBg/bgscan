package speedtest

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestContextReader_NormalRead(t *testing.T) {
	data := "hello world"
	src := strings.NewReader(data)

	cr := &contextReader{
		ctx: context.Background(),
		r:   src,
	}

	buf := make([]byte, 20)

	n, err := cr.Read(buf)
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}

	if n != len(data) {
		t.Fatalf("Read() bytes = %d, want %d", n, len(data))
	}

	if got := string(buf[:n]); got != data {
		t.Fatalf("Read() data = %q, want %q", got, data)
	}
}

func TestContextReader_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	cr := &contextReader{
		ctx: ctx,
		r:   strings.NewReader("some data that won't be read"),
	}

	buf := make([]byte, 20)

	n, err := cr.Read(buf)
	if err != context.Canceled {
		t.Fatalf("Read() error = %v, want %v", err, context.Canceled)
	}

	if n != 0 {
		t.Fatalf("Read() bytes = %d, want 0", n)
	}
}

func TestNewHTTPClient_SOCKS5(t *testing.T) {
	client, err := newHTTPClient(httpClientConfig{
		ProxyPort: 1080,
	})
	if err != nil {
		t.Fatalf("newHTTPClient() error = %v", err)
	}

	if client == nil {
		t.Fatal("newHTTPClient() returned nil")
	}

	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatalf(
			"client.Transport type = %T, want *http.Transport",
			client.Transport,
		)
	}

	if transport.Proxy == nil {
		t.Fatal("transport.Proxy is nil")
	}

	req := &http.Request{
		URL: mustParseURL(t, "https://example.com"),
	}

	proxyURL, err := transport.Proxy(req)
	if err != nil {
		t.Fatalf("transport.Proxy() error = %v", err)
	}

	if proxyURL == nil {
		t.Fatal("transport.Proxy() returned nil")
	}

	want := "socks5://127.0.0.1:1080"
	if got := proxyURL.String(); got != want {
		t.Errorf("proxy URL = %q, want %q", got, want)
	}

	assertHTTPTransportConfig(t, transport)
}

func TestNewHTTPClient_DialContext(t *testing.T) {
	dialer := func(
		ctx context.Context,
		network string,
		address string,
	) (net.Conn, error) {
		return nil, nil
	}

	client, err := newHTTPClient(httpClientConfig{
		DialContext: dialer,
	})
	if err != nil {
		t.Fatalf("newHTTPClient() error = %v", err)
	}

	if client == nil {
		t.Fatal("newHTTPClient() returned nil")
	}

	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatalf(
			"client.Transport type = %T, want *http.Transport",
			client.Transport,
		)
	}

	if transport.Proxy != nil {
		t.Fatal("transport.Proxy is not nil, want nil")
	}

	if transport.DialContext == nil {
		t.Fatal("transport.DialContext is nil")
	}

	assertHTTPTransportConfig(t, transport)
}

func TestNewHTTPClient_DialContextTakesPrecedence(t *testing.T) {
	dialer := func(
		ctx context.Context,
		network string,
		address string,
	) (net.Conn, error) {
		return nil, nil
	}

	client, err := newHTTPClient(httpClientConfig{
		ProxyPort:   1080,
		DialContext: dialer,
	})
	if err != nil {
		t.Fatalf("newHTTPClient() error = %v", err)
	}

	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatalf(
			"client.Transport type = %T, want *http.Transport",
			client.Transport,
		)
	}

	if transport.Proxy != nil {
		t.Fatal("transport.Proxy is not nil, want nil")
	}

	if transport.DialContext == nil {
		t.Fatal("transport.DialContext is nil")
	}
}

func assertHTTPTransportConfig(t *testing.T, transport *http.Transport) {
	t.Helper()

	if transport.TLSHandshakeTimeout != connectTimeout {
		t.Errorf(
			"TLSHandshakeTimeout = %v, want %v",
			transport.TLSHandshakeTimeout,
			connectTimeout,
		)
	}

	if transport.ForceAttemptHTTP2 {
		t.Error("ForceAttemptHTTP2 = true, want false")
	}

	if transport.MaxIdleConns != 10 {
		t.Errorf(
			"MaxIdleConns = %d, want 10",
			transport.MaxIdleConns,
		)
	}

	if transport.MaxIdleConnsPerHost != 10 {
		t.Errorf(
			"MaxIdleConnsPerHost = %d, want 10",
			transport.MaxIdleConnsPerHost,
		)
	}

	if transport.MaxConnsPerHost != 10 {
		t.Errorf(
			"MaxConnsPerHost = %d, want 10",
			transport.MaxConnsPerHost,
		)
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
