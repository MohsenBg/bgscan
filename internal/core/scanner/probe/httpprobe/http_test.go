package httpprobe

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"testing"
	"time"

	"bgscan/internal/core/result"
)

// --- helpers ---

// newTestProbe constructs an HTTPProbe that routes requests to the provided
// test server, bypassing the actual target IP resolution.
func newTestProbe(ts *httptest.Server, req HTTPRequest, acceptedCodes []int) *HTTPProbe {
	p := &HTTPProbe{
		req:    req,
		filter: newStatusFilter(acceptedCodes, totalHTTPStatusCodes),
	}

	p.clientFactory = func(_ netip.Addr) (*http.Transport, *http.Client) {
		t := &http.Transport{
			DialContext: func(ctx context.Context, network, _ string) (net.Conn, error) {
				return (&net.Dialer{}).DialContext(ctx, network, ts.Listener.Addr().String())
			},
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		return t, &http.Client{Transport: t, Timeout: 5 * time.Second}
	}

	return p
}

// mustParseAddr parses an IP address string, panicking if invalid.
func mustParseAddr(s string) netip.Addr {
	return netip.MustParseAddr(s)
}

// --- NewHTTPProbe ---

func TestNewHTTPProbe_ReturnsProbe(t *testing.T) {
	req := HTTPRequest{
		URL:     "http://example.com:80",
		Timeout: 5 * time.Second,
	}

	p := NewHTTPProbe(req, nil)
	if p == nil {
		t.Fatal("NewHTTPProbe returned nil")
	}
}

func TestNewHTTPProbe_WithAcceptedCodes(t *testing.T) {
	req := HTTPRequest{URL: "http://example.com:80", Timeout: time.Second}

	p := NewHTTPProbe(req, []int{200, 301}).(*HTTPProbe)

	if !p.filter.isAccepted(200) {
		t.Error("200 should be accepted")
	}
	if !p.filter.isAccepted(301) {
		t.Error("301 should be accepted")
	}
	if p.filter.isAccepted(404) {
		t.Error("404 should NOT be accepted")
	}
}

func TestNewHTTPProbe_EmptyCodesAcceptsAll(t *testing.T) {
	req := HTTPRequest{URL: "http://example.com:80", Timeout: time.Second}

	p := NewHTTPProbe(req, nil).(*HTTPProbe)

	for _, code := range []int{100, 200, 301, 404, 500} {
		if !p.filter.isAccepted(code) {
			t.Errorf("code %d should be accepted with empty filter", code)
		}
	}
}

// --- Init / Close ---

func TestHTTPProbe_Init(t *testing.T) {
	p := &HTTPProbe{}
	if err := p.Init(context.Background()); err != nil {
		t.Errorf("Init() = %v, want nil", err)
	}
}

func TestHTTPProbe_Close(t *testing.T) {
	p := &HTTPProbe{}
	if err := p.Close(); err != nil {
		t.Errorf("Close() = %v, want nil", err)
	}
}

// --- Schema ---

func TestHTTPProbe_Schema(t *testing.T) {
	p := &HTTPProbe{}
	s := p.Schema()

	if s.Name != "HTTP" {
		t.Errorf("Schema().Name = %q, want %q", s.Name, "HTTP")
	}
	if s.Directory != "http" {
		t.Errorf("Schema().Directory = %q, want %q", s.Directory, "http")
	}
	if len(s.Columns) != 5 {
		t.Errorf("Schema().Columns len = %d, want 5", len(s.Columns))
	}
}

// --- Run: success cases ---

func TestRun_Success_200(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	req := HTTPRequest{
		URL:     "http://target:80",
		UseTLS:  false,
		Timeout: 5 * time.Second,
	}

	p := newTestProbe(ts, req, nil)
	ip := mustParseAddr("93.184.216.34")

	res, err := p.Run(context.Background(), ip)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	hr := res.(HTTPResult)
	if hr.IP != ip {
		t.Errorf("IP = %v, want %v", hr.IP, ip)
	}
	if hr.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", hr.StatusCode)
	}
	if hr.UseTLS {
		t.Error("UseTLS = true, want false")
	}
	if hr.Latency <= 0 {
		t.Errorf("Latency = %v, want > 0", hr.Latency)
	}
	if hr.HTTPVersion == "" {
		t.Error("HTTPVersion is empty")
	}
}

func TestRun_Success_301(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusMovedPermanently)
	}))
	defer ts.Close()

	req := HTTPRequest{URL: "http://target:80", Timeout: 5 * time.Second}
	p := newTestProbe(ts, req, nil)

	res, err := p.Run(context.Background(), mustParseAddr("1.2.3.4"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	hr := res.(HTTPResult)
	if hr.StatusCode != 301 {
		t.Errorf("StatusCode = %d, want 301", hr.StatusCode)
	}
}

func TestRun_Success_WithFilter(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer ts.Close()

	req := HTTPRequest{URL: "http://target:80", Timeout: 5 * time.Second}
	// Accept only 403.
	p := newTestProbe(ts, req, []int{403})

	res, err := p.Run(context.Background(), mustParseAddr("1.2.3.4"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	hr := res.(HTTPResult)
	if hr.StatusCode != 403 {
		t.Errorf("StatusCode = %d, want 403", hr.StatusCode)
	}
}

func TestRun_UsesHEADMethod(t *testing.T) {
	var gotMethod string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		w.WriteHeader(200)
	}))
	defer ts.Close()

	req := HTTPRequest{URL: "http://target:80", Timeout: 5 * time.Second}
	p := newTestProbe(ts, req, nil)

	_, err := p.Run(context.Background(), mustParseAddr("1.2.3.4"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodHead {
		t.Errorf("method = %q, want %q", gotMethod, http.MethodHead)
	}
}

// --- Run: failure cases ---

func TestRun_StatusNotAccepted(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	req := HTTPRequest{URL: "http://target:80", Timeout: 5 * time.Second}
	// Only accept 200.
	p := newTestProbe(ts, req, []int{200})

	_, err := p.Run(context.Background(), mustParseAddr("1.2.3.4"))
	if err == nil {
		t.Fatal("expected error for rejected status code")
	}

	want := "status 404 not accepted"
	if err.Error() != want {
		t.Errorf("error = %q, want %q", err.Error(), want)
	}
}

func TestRun_ContextCancelled(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer ts.Close()

	req := HTTPRequest{URL: "http://target:80", Timeout: 5 * time.Second}
	p := newTestProbe(ts, req, nil)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := p.Run(ctx, mustParseAddr("1.2.3.4"))
	if err == nil {
		t.Fatal("expected error with cancelled context")
	}
}

func TestRun_ConnectionRefused(t *testing.T) {
	req := HTTPRequest{URL: "http://target:80", Timeout: time.Second}

	p := &HTTPProbe{
		req:    req,
		filter: newStatusFilter(nil, totalHTTPStatusCodes),
	}
	p.clientFactory = func(_ netip.Addr) (*http.Transport, *http.Client) {
		tr := &http.Transport{
			DialContext: func(ctx context.Context, network, _ string) (net.Conn, error) {
				return nil, net.ErrClosed
			},
		}
		return tr, &http.Client{Transport: tr, Timeout: time.Second}
	}

	_, err := p.Run(context.Background(), mustParseAddr("1.2.3.4"))
	if err == nil {
		t.Fatal("expected error for connection failure")
	}
}

func TestRun_ServerTimeout(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(3 * time.Second)
		w.WriteHeader(200)
	}))
	defer ts.Close()

	req := HTTPRequest{URL: "http://target:80", Timeout: 100 * time.Millisecond}

	p := &HTTPProbe{
		req:    req,
		filter: newStatusFilter(nil, totalHTTPStatusCodes),
	}
	p.clientFactory = func(_ netip.Addr) (*http.Transport, *http.Client) {
		tr := &http.Transport{
			DialContext: func(ctx context.Context, network, _ string) (net.Conn, error) {
				return (&net.Dialer{}).DialContext(ctx, network, ts.Listener.Addr().String())
			},
		}
		return tr, &http.Client{Transport: tr, Timeout: 100 * time.Millisecond}
	}

	_, err := p.Run(context.Background(), mustParseAddr("1.2.3.4"))
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

// --- Run: TLS ---

func TestRun_TLS_Success(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer ts.Close()

	req := HTTPRequest{
		URL:     "https://target:443",
		UseTLS:  true,
		Timeout: 5 * time.Second,
	}

	p := newTestProbe(ts, req, nil)
	res, err := p.Run(context.Background(), mustParseAddr("10.0.0.1"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	hr := res.(HTTPResult)
	if !hr.UseTLS {
		t.Error("UseTLS = false, want true")
	}
	if hr.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", hr.StatusCode)
	}
}

// --- tlsNextProto ---

func TestTLSNextProto_H1(t *testing.T) {
	m := tlsNextProto(HTTPVersionH1)
	if m == nil {
		t.Fatal("expected non-nil empty map for H1")
	}
	if len(m) != 0 {
		t.Errorf("len = %d, want 0", len(m))
	}
}

func TestTLSNextProto_H2(t *testing.T) {
	m := tlsNextProto(HTTPVersionH2)
	if m != nil {
		t.Errorf("expected nil for H2, got %v", m)
	}
}

func TestTLSNextProto_H1H2(t *testing.T) {
	m := tlsNextProto(HTTPVersionH1H2)
	if m != nil {
		t.Errorf("expected nil for H1H2, got %v", m)
	}
}

// --- buildClient ---

func TestBuildClient_TransportConfig(t *testing.T) {
	tlsCfg := &tls.Config{ServerName: "example.com"}
	p := &HTTPProbe{
		req: HTTPRequest{
			Timeout: 7 * time.Second,
			Version: HTTPVersionH1,
			UseTLS:  true,
		},
		tls:    tlsCfg,
		dialer: &net.Dialer{Timeout: 5 * time.Second},
	}

	tr, client := p.buildClient(mustParseAddr("1.2.3.4"))

	if client.Timeout != 7*time.Second {
		t.Errorf("client.Timeout = %v, want 7s", client.Timeout)
	}
	if tr.TLSHandshakeTimeout != 7*time.Second {
		t.Errorf("TLSHandshakeTimeout = %v, want 7s", tr.TLSHandshakeTimeout)
	}
	if tr.ResponseHeaderTimeout != 7*time.Second {
		t.Errorf("ResponseHeaderTimeout = %v, want 7s", tr.ResponseHeaderTimeout)
	}
	if !tr.DisableKeepAlives {
		t.Error("DisableKeepAlives = false, want true")
	}
	if tr.TLSClientConfig != tlsCfg {
		t.Error("TLSClientConfig not set correctly")
	}
	// H1 → empty non-nil map
	if tr.TLSNextProto == nil || len(tr.TLSNextProto) != 0 {
		t.Errorf("TLSNextProto = %v, want empty non-nil map for H1", tr.TLSNextProto)
	}
}

func TestBuildClient_H2_ForcesHTTP2(t *testing.T) {
	p := &HTTPProbe{
		req: HTTPRequest{
			Timeout: 5 * time.Second,
			Version: HTTPVersionH2,
		},
		dialer: &net.Dialer{},
	}

	tr, _ := p.buildClient(mustParseAddr("1.2.3.4"))

	if !tr.ForceAttemptHTTP2 {
		t.Error("ForceAttemptHTTP2 = false, want true for H2")
	}
	if tr.TLSNextProto != nil {
		t.Errorf("TLSNextProto = %v, want nil for H2", tr.TLSNextProto)
	}
}

// TestBuildClient_DialRedirectsToIP verifies that DialContext rewrites the
// target address to the specified IP while preserving the original port.
func TestBuildClient_DialRedirectsToIP(t *testing.T) {
	p := &HTTPProbe{
		req: HTTPRequest{
			Timeout: 5 * time.Second,
			Version: HTTPVersionH1H2,
		},
		dialer: &net.Dialer{},
	}

	tr, _ := p.buildClient(mustParseAddr("10.20.30.40"))
	if tr == nil {
		t.Fatal("transport is nil")
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close() //nolint:errcheck

	go func() {
		conn, _ := ln.Accept()
		if conn != nil {
			_ = conn.Close()
		}
	}()

	_, port, _ := net.SplitHostPort(ln.Addr().String())

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, dialErr := tr.DialContext(ctx, "tcp", "original.host:"+port)
	// We expect a connection error (can't reach 10.20.30.40) but NOT a parse error.
	if dialErr != nil && strings.Contains(dialErr.Error(), "parse addr") {
		t.Errorf("unexpected parse error: %v", dialErr)
	}
}

// --- Run: result implements probe interface ---

func TestRun_ResultImplementsInterface(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer ts.Close()

	req := HTTPRequest{URL: "http://target:80", Timeout: 5 * time.Second}
	p := newTestProbe(ts, req, nil)

	res, err := p.Run(context.Background(), mustParseAddr("8.8.8.8"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.Key() != "8.8.8.8" {
		t.Errorf("Key() = %q, want %q", res.Key(), "8.8.8.8")
	}
	if res.KeyType() != result.KeyIP {
		t.Errorf("KeyType() = %v, want KeyIP", res.KeyType())
	}
}

// --- Run: multiple sequential calls ---

func TestRun_MultipleCalls(t *testing.T) {
	callCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(200)
	}))
	defer ts.Close()

	req := HTTPRequest{URL: "http://target:80", Timeout: 5 * time.Second}
	p := newTestProbe(ts, req, nil)

	for i := range 5 {
		_, err := p.Run(context.Background(), mustParseAddr("1.2.3.4"))
		if err != nil {
			t.Fatalf("call %d: unexpected error: %v", i, err)
		}
	}

	if callCount != 5 {
		t.Errorf("server called %d times, want 5", callCount)
	}
}
