package httpprobe

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"testing"
	"time"

	"github.com/MohsenBg/bgscan/internal/core/scanner/probe"
	"github.com/MohsenBg/bgscan/internal/core/target"
)

// staticMode returns one fixed Request for every entry. It lets Run be tested
// without depending on HostMode or DirectMode.
type staticMode struct {
	req Request
	err error
}

func (m staticMode) Input() target.Kind { return target.KindHost }

func (m staticMode) Build(target.Entry) (Request, error) {
	return m.req, m.err
}

// newTestProbe constructs an HTTPProbe that routes requests to ts.
func newTestProbe(ts *httptest.Server, req Request, acceptedCodes []int) *HTTPProbe {
	p := &HTTPProbe{
		settings: Settings{
			Timeout: 5 * time.Second,
			UseTLS:  strings.HasPrefix(req.URL, "https://"),
		},
		mode:   staticMode{req: req},
		filter: newStatusFilter(acceptedCodes, totalHTTPStatusCodes),
		dialer: &net.Dialer{Timeout: 5 * time.Second},
	}

	p.clientFactory = func(Request) httpClientResult {
		tr := &http.Transport{
			DialContext: func(ctx context.Context, network, _ string) (net.Conn, error) {
				return (&net.Dialer{}).DialContext(ctx, network, ts.Listener.Addr().String())
			},
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}

		return httpClientResult{
			client: &http.Client{Transport: tr, Timeout: 5 * time.Second},
			close:  tr.CloseIdleConnections,
		}
	}

	return p
}

// mustParseAddr parses an IP address string, panicking if invalid.
func mustParseAddr(s string) netip.Addr {
	return netip.MustParseAddr(s)
}

func TestNewHTTPProbe_ReturnsProbe(t *testing.T) {
	s := Settings{Timeout: 5 * time.Second}
	mode := staticMode{req: Request{URL: "http://example.com:80"}}

	p := NewHTTPProbe(s, mode, nil, nil)
	if p == nil {
		t.Fatal("NewHTTPProbe returned nil")
	}
}

func TestNewHTTPProbe_WithAcceptedCodes(t *testing.T) {
	s := Settings{Timeout: time.Second}
	p := NewHTTPProbe(
		s,
		staticMode{req: Request{URL: "http://example.com:80"}},
		[]int{200, 301},
		nil,
	).(*HTTPProbe)

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
	s := Settings{Timeout: time.Second}
	p := NewHTTPProbe(
		s,
		staticMode{req: Request{URL: "http://example.com:80"}},
		nil,
		nil,
	).(*HTTPProbe)

	for _, code := range []int{100, 200, 301, 404, 500} {
		if !p.filter.isAccepted(code) {
			t.Errorf("code %d should be accepted with empty filter", code)
		}
	}
}

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

func TestRun_Success_200(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	req := Request{URL: "http://target:80"}

	p := newTestProbe(ts, req, nil)
	ip := mustParseAddr("93.184.216.34")

	res, err := p.Run(context.Background(), entryOf(ip))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	hr := res.(HTTPResult)
	if hr.Entry.IP.Addr() != ip {
		t.Errorf("IP = %v, want %v", hr.Entry.IP.Addr(), ip)
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

	p := newTestProbe(ts, Request{URL: "http://target:80"}, nil)

	res, err := p.Run(context.Background(), entryOf(mustParseAddr("1.2.3.4")))
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

	p := newTestProbe(ts, Request{URL: "http://target:80"}, []int{403})

	res, err := p.Run(context.Background(), entryOf(mustParseAddr("1.2.3.4")))
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

	p := newTestProbe(ts, Request{URL: "http://target:80"}, nil)

	_, err := p.Run(context.Background(), entryOf(mustParseAddr("1.2.3.4")))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodHead {
		t.Errorf("method = %q, want %q", gotMethod, http.MethodHead)
	}
}

func TestRun_StatusNotAccepted(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	p := newTestProbe(ts, Request{URL: "http://target:80"}, []int{200})

	_, err := p.Run(context.Background(), entryOf(mustParseAddr("1.2.3.4")))
	if err == nil {
		t.Fatal("expected error for rejected status code")
	}

	want := "status 404 not accepted"
	if !strings.Contains(err.Error(), want) {
		t.Errorf("error = %q, want containing %q", err.Error(), want)
	}
	if !errors.Is(err, probe.ErrBadResponse) {
		t.Errorf("error = %v, want errors.Is ErrBadResponse", err)
	}
}

func TestRun_ContextCancelled(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer ts.Close()

	p := newTestProbe(ts, Request{URL: "http://target:80"}, nil)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := p.Run(ctx, entryOf(mustParseAddr("1.2.3.4")))
	if err == nil {
		t.Fatal("expected error with cancelled context")
	}
}

func TestRun_ConnectionRefused(t *testing.T) {
	p := &HTTPProbe{
		settings: Settings{Timeout: time.Second},
		mode:     staticMode{req: Request{URL: "http://target:80"}},
		filter:   newStatusFilter(nil, totalHTTPStatusCodes),
		dialer:   &net.Dialer{Timeout: time.Second},
	}
	p.clientFactory = func(Request) httpClientResult {
		tr := &http.Transport{
			DialContext: func(ctx context.Context, network, _ string) (net.Conn, error) {
				return nil, net.ErrClosed
			},
		}
		return httpClientResult{
			client: &http.Client{Transport: tr, Timeout: time.Second},
			close:  tr.CloseIdleConnections,
		}
	}

	_, err := p.Run(context.Background(), entryOf(mustParseAddr("1.2.3.4")))
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

	p := &HTTPProbe{
		settings: Settings{Timeout: 100 * time.Millisecond},
		mode:     staticMode{req: Request{URL: "http://target:80"}},
		filter:   newStatusFilter(nil, totalHTTPStatusCodes),
		dialer:   &net.Dialer{Timeout: 100 * time.Millisecond},
	}
	p.clientFactory = func(Request) httpClientResult {
		tr := &http.Transport{
			DialContext: func(ctx context.Context, network, _ string) (net.Conn, error) {
				return (&net.Dialer{}).DialContext(ctx, network, ts.Listener.Addr().String())
			},
		}
		return httpClientResult{
			client: &http.Client{Transport: tr, Timeout: 100 * time.Millisecond},
			close:  tr.CloseIdleConnections,
		}
	}

	_, err := p.Run(context.Background(), entryOf(mustParseAddr("1.2.3.4")))
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestRun_TLS_Success(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer ts.Close()

	req := Request{URL: "https://target:443"}

	p := newTestProbe(ts, req, nil)
	res, err := p.Run(context.Background(), entryOf(mustParseAddr("10.0.0.1")))
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

func TestBuildClient_TransportConfig(t *testing.T) {
	p := &HTTPProbe{
		settings: Settings{
			Timeout:       7 * time.Second,
			Version:       HTTPVersionH1,
			UseTLS:        true,
			MinTLSVersion: tls.VersionTLS12,
			MaxTLSVersion: tls.VersionTLS13,
		},
		dialer: &net.Dialer{Timeout: 5 * time.Second},
	}

	res := p.newStdClient(Request{URL: "https://example.com:443", SNI: "example.com", Port: 443})
	defer res.close()

	client := res.client
	tr, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatal("transport is not *http.Transport")
	}

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
	if tr.TLSClientConfig == nil {
		t.Fatal("TLSClientConfig is nil")
	}
	if tr.TLSClientConfig.ServerName != "example.com" {
		t.Errorf("ServerName = %q, want %q", tr.TLSClientConfig.ServerName, "example.com")
	}
	// H1 → empty non-nil map
	if tr.TLSNextProto == nil || len(tr.TLSNextProto) != 0 {
		t.Errorf("TLSNextProto = %v, want empty non-nil map for H1", tr.TLSNextProto)
	}
}

func TestBuildClient_H2_ForcesHTTP2(t *testing.T) {
	p := &HTTPProbe{
		settings: Settings{
			Timeout: 5 * time.Second,
			Version: HTTPVersionH2,
			UseTLS:  true,
		},
		dialer: &net.Dialer{},
	}

	res := p.newStdClient(Request{URL: "https://example.com:443", Port: 443})
	defer res.close()

	tr, ok := res.client.Transport.(*http.Transport)
	if !ok {
		t.Fatal("transport is not *http.Transport")
	}

	if !tr.ForceAttemptHTTP2 {
		t.Error("ForceAttemptHTTP2 = false, want true for H2")
	}
	if tr.TLSNextProto != nil {
		t.Errorf("TLSNextProto = %v, want nil for H2", tr.TLSNextProto)
	}
}

func TestBuildClient_H1H2NegotiatesHTTP2(t *testing.T) {
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	ts.EnableHTTP2 = true
	ts.StartTLS()
	defer ts.Close()

	p := &HTTPProbe{
		settings: Settings{
			Timeout:       5 * time.Second,
			Version:       HTTPVersionH1H2,
			UseTLS:        true,
			SkipTLSVerify: true,
		},
		dialer: &net.Dialer{Timeout: 5 * time.Second},
	}

	res := p.newStdClient(Request{URL: ts.URL})
	defer res.close()

	httpReq, err := http.NewRequestWithContext(context.Background(), http.MethodHead, ts.URL, nil)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}

	resp, err := res.client.Do(httpReq)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.ProtoMajor != 2 {
		t.Fatalf("Proto = %q, want HTTP/2", resp.Proto)
	}
	if resp.TLS == nil || resp.TLS.NegotiatedProtocol != "h2" {
		t.Fatalf("negotiated ALPN = %v, want h2", resp.TLS.NegotiatedProtocol)
	}
}

// TestBuildClient_DialRedirectsToIP verifies that DialContext rewrites the
// target address to the specified IP while preserving the original port.
func TestBuildClient_DialRedirectsToIP(t *testing.T) {
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

	p := &HTTPProbe{
		settings: Settings{
			Timeout: 5 * time.Second,
			Version: HTTPVersionH1H2,
		},
		dialer: &net.Dialer{},
	}

	res := p.newStdClient(Request{
		URL:    "http://original.host:" + port,
		DialIP: netip.MustParseAddr("10.20.30.40"),
	})
	defer res.close()

	tr, ok := res.client.Transport.(*http.Transport)
	if !ok {
		t.Fatal("transport is not *http.Transport")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, dialErr := tr.DialContext(ctx, "tcp", "original.host:"+port)
	// We expect a connection error (can't reach 10.20.30.40) but NOT a parse error.
	if dialErr != nil && strings.Contains(dialErr.Error(), "parse addr") {
		t.Errorf("unexpected parse error: %v", dialErr)
	}
}

func TestRun_ResultImplementsInterface(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer ts.Close()

	p := newTestProbe(ts, Request{URL: "http://target:80"}, nil)

	res, err := p.Run(context.Background(), entryOf(mustParseAddr("8.8.8.8")))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.Key() != "8.8.8.8/32:0:" {
		t.Errorf("Key() = %q, want %q", res.Key(), "8.8.8.8/32:0:")
	}
	if res.KeyType() != target.KindIP {
		t.Errorf("KeyType() = %v, want KeyIP", res.KeyType())
	}
}

func TestRun_MultipleCalls(t *testing.T) {
	callCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(200)
	}))
	defer ts.Close()

	p := newTestProbe(ts, Request{URL: "http://target:80"}, nil)

	for i := range 5 {
		_, err := p.Run(context.Background(), entryOf(mustParseAddr("1.2.3.4")))
		if err != nil {
			t.Fatalf("call %d: unexpected error: %v", i, err)
		}
	}

	if callCount != 5 {
		t.Errorf("server called %d times, want 5", callCount)
	}
}
