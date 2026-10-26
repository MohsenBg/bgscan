package speedtest

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"testing"
	"time"
)

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

func TestNewHTTPClient_RequiresDialContextOrProxyPort(t *testing.T) {
	if _, err := newHTTPClient(httpClientConfig{}); err == nil {
		t.Fatal("expected an error when both DialContext and ProxyPort are unset")
	}
}

func TestNewHTTPClient_DialContextGetsConnectTimeout(t *testing.T) {
	var deadline time.Time
	var hasDeadline bool

	dialer := func(
		ctx context.Context,
		network string,
		address string,
	) (net.Conn, error) {
		deadline, hasDeadline = ctx.Deadline()
		return nil, nil
	}

	client, err := newHTTPClient(httpClientConfig{DialContext: dialer})
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

	if _, err := transport.DialContext(
		context.Background(),
		"tcp",
		"example.com:443",
	); err != nil {
		t.Fatalf("DialContext() error = %v", err)
	}

	if !hasDeadline {
		t.Fatal("DialContext context has no deadline")
	}

	remaining := time.Until(deadline)
	if remaining <= 0 || remaining > connectTimeout {
		t.Fatalf(
			"DialContext deadline is %v away, want within (0, %v]",
			remaining,
			connectTimeout,
		)
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

	if !transport.ForceAttemptHTTP2 {
		t.Error("ForceAttemptHTTP2 = false, want true")
	}

	if transport.ResponseHeaderTimeout != connectTimeout {
		t.Errorf(
			"ResponseHeaderTimeout = %v, want %v",
			transport.ResponseHeaderTimeout,
			connectTimeout,
		)
	}

	if transport.IdleConnTimeout != 90*time.Second {
		t.Errorf(
			"IdleConnTimeout = %v, want %v",
			transport.IdleConnTimeout,
			90*time.Second,
		)
	}

	if transport.ExpectContinueTimeout != time.Second {
		t.Errorf(
			"ExpectContinueTimeout = %v, want %v",
			transport.ExpectContinueTimeout,
			time.Second,
		)
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
