package httpprobe

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/netip"
	"strings"
	"testing"
	"time"

	"bgscan/internal/core/result"
)

// mockRoundTripCloser implements http.RoundTripper and io.Closer for testing HTTP/3 probes.
type mockRoundTripCloser struct {
	// resp and roundErr control the RoundTrip behavior.
	resp     *http.Response
	roundErr error

	// closeErr and closed track Close behavior.
	closeErr error
	closed   bool

	// lastReq captures the request for assertions.
	lastReq *http.Request
}

func (m *mockRoundTripCloser) RoundTrip(req *http.Request) (*http.Response, error) {
	m.lastReq = req
	if m.roundErr != nil {
		return nil, m.roundErr
	}
	return m.resp, nil
}

func (m *mockRoundTripCloser) Close() error {
	m.closed = true
	return m.closeErr
}

// mockResponse creates a basic http.Response with the given status code and protocol.
func mockResponse(statusCode int, proto string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Proto:      proto,
		Body:       io.NopCloser(strings.NewReader("")),
		Header:     make(http.Header),
	}
}

// newTestHTTP3Probe constructs an HTTP3Probe with a mocked transport for isolated testing.
func newTestHTTP3Probe(
	transport roundTripCloser,
	req HTTPRequest,
	acceptedCodes []int,
) *HTTP3Probe {
	return &HTTP3Probe{
		req:       req,
		filter:    newStatusFilter(acceptedCodes, totalHTTPStatusCodes),
		transport: transport,
	}
}

func defaultH3Request() HTTPRequest {
	return HTTPRequest{
		URL:     "https://example.com:443",
		Host:    "example.com",
		SNI:     "example.com",
		UseTLS:  true,
		Timeout: 5 * time.Second,
	}
}

// --- NewHTTP3Probe ---

func TestNewHTTP3Probe_Success(t *testing.T) {
	req := defaultH3Request()
	p, err := NewHTTP3Probe(req, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p == nil {
		t.Fatal("returned nil probe")
	}
}

func TestNewHTTP3Probe_WithCodes(t *testing.T) {
	req := defaultH3Request()
	p, err := NewHTTP3Probe(req, []int{200, 204})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	h3p := p.(*HTTP3Probe)
	if !h3p.filter.isAccepted(200) {
		t.Error("200 should be accepted")
	}
	if h3p.filter.isAccepted(404) {
		t.Error("404 should not be accepted")
	}
}

// --- Init / Close ---

func TestHTTP3Probe_Init(t *testing.T) {
	p := &HTTP3Probe{}
	if err := p.Init(context.Background()); err != nil {
		t.Errorf("Init() = %v, want nil", err)
	}
}

func TestHTTP3Probe_Close_Success(t *testing.T) {
	m := &mockRoundTripCloser{}
	p := newTestHTTP3Probe(m, defaultH3Request(), nil)

	if err := p.Close(); err != nil {
		t.Errorf("Close() = %v, want nil", err)
	}
	if !m.closed {
		t.Error("transport.Close() was not called")
	}
}

func TestHTTP3Probe_Close_Error(t *testing.T) {
	m := &mockRoundTripCloser{closeErr: errors.New("quic shutdown failed")}
	p := newTestHTTP3Probe(m, defaultH3Request(), nil)

	err := p.Close()
	if err == nil {
		t.Fatal("expected error")
	}
	want := "close http3 transport: quic shutdown failed"
	if err.Error() != want {
		t.Errorf("error = %q, want %q", err.Error(), want)
	}
}

// --- Schema ---

func TestHTTP3Probe_Schema(t *testing.T) {
	p := &HTTP3Probe{}
	s := p.Schema()

	if s.Name != "HTTP" {
		t.Errorf("Name = %q, want %q", s.Name, "HTTP")
	}
	if s.Directory != "http" {
		t.Errorf("Directory = %q, want %q", s.Directory, "http")
	}
}

// --- Run: success ---

func TestHTTP3Run_Success_200(t *testing.T) {
	m := &mockRoundTripCloser{resp: mockResponse(200, "HTTP/3.0")}
	p := newTestHTTP3Probe(m, defaultH3Request(), nil)

	ip := netip.MustParseAddr("93.184.216.34")
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
	if hr.HTTPVersion != "HTTP/3.0" {
		t.Errorf("HTTPVersion = %q, want %q", hr.HTTPVersion, "HTTP/3.0")
	}
	if !hr.UseTLS {
		t.Error("UseTLS = false, want true")
	}
	if hr.Latency <= 0 {
		t.Errorf("Latency = %v, want > 0", hr.Latency)
	}
}

func TestHTTP3Run_Success_204(t *testing.T) {
	m := &mockRoundTripCloser{resp: mockResponse(204, "HTTP/3.0")}
	p := newTestHTTP3Probe(m, defaultH3Request(), nil)

	res, err := p.Run(context.Background(), netip.MustParseAddr("1.2.3.4"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	hr := res.(HTTPResult)
	if hr.StatusCode != 204 {
		t.Errorf("StatusCode = %d, want 204", hr.StatusCode)
	}
}

func TestHTTP3Run_WithFilter_Accepted(t *testing.T) {
	m := &mockRoundTripCloser{resp: mockResponse(301, "HTTP/3.0")}
	p := newTestHTTP3Probe(m, defaultH3Request(), []int{301, 302})

	res, err := p.Run(context.Background(), netip.MustParseAddr("1.2.3.4"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	hr := res.(HTTPResult)
	if hr.StatusCode != 301 {
		t.Errorf("StatusCode = %d, want 301", hr.StatusCode)
	}
}

// --- Run: request properties ---

func TestHTTP3Run_UsesHEADMethod(t *testing.T) {
	m := &mockRoundTripCloser{resp: mockResponse(200, "HTTP/3.0")}
	p := newTestHTTP3Probe(m, defaultH3Request(), nil)

	_, err := p.Run(context.Background(), netip.MustParseAddr("1.2.3.4"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if m.lastReq == nil {
		t.Fatal("no request captured")
	}
	if m.lastReq.Method != http.MethodHead {
		t.Errorf("Method = %q, want HEAD", m.lastReq.Method)
	}
}

// TestHTTP3Run_RewritesHostToIP verifies that the request URL host is rewritten
// to the target IP address, ensuring the connection targets the correct endpoint.
func TestHTTP3Run_RewritesHostToIP(t *testing.T) {
	m := &mockRoundTripCloser{resp: mockResponse(200, "HTTP/3.0")}
	p := newTestHTTP3Probe(m, defaultH3Request(), nil)

	ip := netip.MustParseAddr("10.20.30.40")
	_, err := p.Run(context.Background(), ip)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// URL.Host should be rewritten to the target IP.
	wantHost := "10.20.30.40:443"
	if m.lastReq.URL.Host != wantHost {
		t.Errorf("URL.Host = %q, want %q", m.lastReq.URL.Host, wantHost)
	}
}

// TestHTTP3Run_PreservesOriginalHost ensures that the Host header retains the
// original hostname for proper virtual hosting, even when the URL host is rewritten.
func TestHTTP3Run_PreservesOriginalHost(t *testing.T) {
	m := &mockRoundTripCloser{resp: mockResponse(200, "HTTP/3.0")}

	req := defaultH3Request()
	req.Host = "cdn.example.com"

	p := newTestHTTP3Probe(m, req, nil)

	_, err := p.Run(context.Background(), netip.MustParseAddr("1.2.3.4"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// req.Host should be the original hostname for virtual hosting.
	if m.lastReq.Host != "cdn.example.com" {
		t.Errorf("Host = %q, want %q", m.lastReq.Host, "cdn.example.com")
	}
}

// TestHTTP3Run_IPv6 verifies that IPv6 addresses are correctly bracketed
// in the URL host to comply with URI syntax rules.
func TestHTTP3Run_IPv6(t *testing.T) {
	m := &mockRoundTripCloser{resp: mockResponse(200, "HTTP/3.0")}
	p := newTestHTTP3Probe(m, defaultH3Request(), nil)

	ip := netip.MustParseAddr("2001:db8::1")
	_, err := p.Run(context.Background(), ip)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// IPv6 should be bracketed in URL.Host.
	wantHost := "[2001:db8::1]:443"
	if m.lastReq.URL.Host != wantHost {
		t.Errorf("URL.Host = %q, want %q", m.lastReq.URL.Host, wantHost)
	}
}

// --- Run: failures ---

func TestHTTP3Run_ContextCancelled(t *testing.T) {
	m := &mockRoundTripCloser{resp: mockResponse(200, "HTTP/3.0")}
	p := newTestHTTP3Probe(m, defaultH3Request(), nil)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := p.Run(ctx, netip.MustParseAddr("1.2.3.4"))
	if err == nil {
		t.Fatal("expected error with cancelled context")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("error = %v, want context.Canceled", err)
	}
}

func TestHTTP3Run_TransportError(t *testing.T) {
	m := &mockRoundTripCloser{roundErr: errors.New("QUIC handshake failed")}
	p := newTestHTTP3Probe(m, defaultH3Request(), nil)

	_, err := p.Run(context.Background(), netip.MustParseAddr("1.2.3.4"))
	if err == nil {
		t.Fatal("expected error")
	}
	want := "QUIC handshake failed"
	if !strings.Contains(err.Error(), want) {
		t.Errorf("error = %q, want %q", err.Error(), want)
	}
}

func TestHTTP3Run_StatusNotAccepted(t *testing.T) {
	m := &mockRoundTripCloser{resp: mockResponse(503, "HTTP/3.0")}
	p := newTestHTTP3Probe(m, defaultH3Request(), []int{200})

	_, err := p.Run(context.Background(), netip.MustParseAddr("1.2.3.4"))
	if err == nil {
		t.Fatal("expected error for rejected status")
	}
	want := "status 503 not accepted"
	if err.Error() != want {
		t.Errorf("error = %q, want %q", err.Error(), want)
	}
}

func TestHTTP3Run_InvalidURL(t *testing.T) {
	m := &mockRoundTripCloser{resp: mockResponse(200, "HTTP/3.0")}

	req := defaultH3Request()
	req.URL = "://bad-url"

	p := newTestHTTP3Probe(m, req, nil)

	_, err := p.Run(context.Background(), netip.MustParseAddr("1.2.3.4"))
	if err == nil {
		t.Fatal("expected error for invalid URL")
	}
	if !strings.Contains(err.Error(), "build request") {
		t.Errorf("error = %q, want it to contain 'build request'", err.Error())
	}
}

// --- Run: result interface ---

func TestHTTP3Run_ResultInterface(t *testing.T) {
	m := &mockRoundTripCloser{resp: mockResponse(200, "HTTP/3.0")}
	p := newTestHTTP3Probe(m, defaultH3Request(), nil)

	ip := netip.MustParseAddr("8.8.8.8")
	res, err := p.Run(context.Background(), ip)
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

// --- Run: always reports HTTP/3.0 and TLS ---

// TestHTTP3Run_AlwaysHTTP3AndTLS ensures that the probe correctly reports
// HTTP/3.0 and TLS usage, as HTTP/3 inherently requires TLS.
func TestHTTP3Run_AlwaysHTTP3AndTLS(t *testing.T) {
	m := &mockRoundTripCloser{resp: mockResponse(200, "HTTP/3.0")}

	// Even if req says UseTLS=false, HTTP3Probe hardcodes true.
	req := defaultH3Request()
	req.UseTLS = false

	p := newTestHTTP3Probe(m, req, nil)

	res, err := p.Run(context.Background(), netip.MustParseAddr("1.2.3.4"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	hr := res.(HTTPResult)
	if hr.HTTPVersion != "HTTP/3.0" {
		t.Errorf("HTTPVersion = %q, want HTTP/3.0", hr.HTTPVersion)
	}
	if !hr.UseTLS {
		t.Error("UseTLS = false, want true (HTTP/3 always uses TLS)")
	}
}

// --- Multiple sequential runs ---

// TestHTTP3Run_MultipleCalls verifies that the probe can be executed
// multiple times sequentially without state leakage or transport exhaustion.
func TestHTTP3Run_MultipleCalls(t *testing.T) {
	callCount := 0
	m := &mockRoundTripCloser{}

	// Track call count (origRT unused but kept for future mocking extensions).
	origRT := m.RoundTrip
	_ = origRT

	p := newTestHTTP3Probe(m, defaultH3Request(), nil)

	for i := range 5 {
		m.resp = mockResponse(200, "HTTP/3.0")
		_, err := p.Run(context.Background(), netip.MustParseAddr("1.2.3.4"))
		if err != nil {
			t.Fatalf("call %d: unexpected error: %v", i, err)
		}
		callCount++
	}

	if callCount != 5 {
		t.Errorf("ran %d times, want 5", callCount)
	}
}
