package tcpprobe

import (
	"context"
	"errors"
	"net"
	"net/netip"
	"strconv"
	"strings"
	"testing"
	"time"

	"bgscan/internal/core/netutil"
	"bgscan/internal/core/result"
)

// mustParseAddr parses an IP address string, failing the test if invalid.
func mustParseAddr(t *testing.T, s string) netip.Addr {
	t.Helper()

	ip, err := netip.ParseAddr(s)
	if err != nil {
		t.Fatalf("ParseAddr(%q) error = %v", s, err)
	}

	return ip
}

// startTCPListener binds a TCP listener to a random available port on localhost.
func startTCPListener(t *testing.T) (net.Listener, uint16) {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen() error = %v", err)
	}

	port := uint16(ln.Addr().(*net.TCPAddr).Port)
	return ln, port
}

// getUnusedTCPPort finds and returns an available TCP port by briefly binding and closing a listener.
func getUnusedTCPPort(t *testing.T) uint16 {
	t.Helper()

	ln, port := startTCPListener(t)
	_ = ln.Close()
	return port
}

// acceptLoop continuously accepts and immediately closes incoming connections to keep the port open during tests.
func acceptLoop(t *testing.T, ln net.Listener, stop <-chan struct{}) {
	t.Helper()

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				select {
				case <-stop:
					return
				default:
				}
				return
			}
			_ = conn.Close()
		}
	}()
}

func TestNewTCPProbeValidPort(t *testing.T) {
	t.Parallel()

	p := NewTCPProbe("443", 2*time.Second, 3)

	tcp, ok := p.(*TCPProbe)
	if !ok {
		t.Fatalf("NewTCPProbe() type = %T, want *TCPProbe", p)
	}

	if tcp.port != 443 {
		t.Fatalf("port = %d, want %d", tcp.port, 443)
	}
	if tcp.timeout != 2*time.Second {
		t.Fatalf("timeout = %v, want %v", tcp.timeout, 2*time.Second)
	}
	if tcp.tries != 3 {
		t.Fatalf("tries = %d, want %d", tcp.tries, 3)
	}
	if tcp.dialer.Timeout != 2*time.Second {
		t.Fatalf("dialer.Timeout = %v, want %v", tcp.dialer.Timeout, 2*time.Second)
	}
}

func TestNewTCPProbeInvalidPortFallsBackTo80(t *testing.T) {
	t.Parallel()

	p := NewTCPProbe("not-a-port", 1500*time.Millisecond, 2)

	tcp, ok := p.(*TCPProbe)
	if !ok {
		t.Fatalf("NewTCPProbe() type = %T, want *TCPProbe", p)
	}

	if tcp.port != 80 {
		t.Fatalf("port = %d, want %d", tcp.port, 80)
	}
}

func TestTCPProbeSchema(t *testing.T) {
	t.Parallel()

	p := &TCPProbe{}
	got := p.Schema()

	if got.Name != Schema.Name {
		t.Fatalf("Schema().Name = %q, want %q", got.Name, Schema.Name)
	}
	if got.Directory != Schema.Directory {
		t.Fatalf("Schema().Directory = %q, want %q", got.Directory, Schema.Directory)
	}
	if len(got.Columns) != len(Schema.Columns) {
		t.Fatalf("len(Schema().Columns) = %d, want %d", len(got.Columns), len(Schema.Columns))
	}
	if got.Parser == nil {
		t.Fatal("Schema().Parser is nil")
	}
}

func TestTCPProbeInitNoOp(t *testing.T) {
	t.Parallel()

	p := &TCPProbe{}

	if err := p.Init(context.Background()); err != nil {
		t.Fatalf("Init() error = %v, want nil", err)
	}
}

func TestTCPProbeCloseNoOp(t *testing.T) {
	t.Parallel()

	p := &TCPProbe{}

	if err := p.Close(); err != nil {
		t.Fatalf("Close() error = %v, want nil", err)
	}
}

func TestTCPProbeRunContextCanceled(t *testing.T) {
	t.Parallel()

	p := &TCPProbe{
		port:    80,
		timeout: time.Second,
		tries:   3,
		dialer:  net.Dialer{Timeout: time.Second},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := p.Run(ctx, mustParseAddr(t, "127.0.0.1"))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Run() error = %v, want %v", err, context.Canceled)
	}
}

func TestTCPProbeRunSuccess(t *testing.T) {
	t.Parallel()

	ln, port := startTCPListener(t)
	defer func() { _ = ln.Close() }()

	stop := make(chan struct{})
	defer close(stop)
	acceptLoop(t, ln, stop)

	p := &TCPProbe{
		port:    port,
		timeout: 2 * time.Second,
		tries:   3,
		dialer:  net.Dialer{Timeout: 2 * time.Second},
	}

	got, err := p.Run(context.Background(), mustParseAddr(t, "127.0.0.1"))
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	res, ok := got.(TCPResult)
	if !ok {
		t.Fatalf("Run() result type = %T, want %T", got, TCPResult{})
	}

	if res.IP != mustParseAddr(t, "127.0.0.1") {
		t.Fatalf("IP = %v, want %v", res.IP, mustParseAddr(t, "127.0.0.1"))
	}
	if res.Port != port {
		t.Fatalf("Port = %d, want %d", res.Port, port)
	}
	if res.Tries != 1 {
		t.Fatalf("Tries = %d, want %d", res.Tries, 1)
	}
	if res.Latency <= 0 {
		t.Fatalf("Latency = %v, want > 0", res.Latency)
	}
}

// TestTCPProbeRunConnectionRefusedFailsFast verifies that non-timeout errors
// (like connection refused) fail immediately without exhausting all retry attempts.
func TestTCPProbeRunConnectionRefusedFailsFast(t *testing.T) {
	t.Parallel()

	port := getUnusedTCPPort(t)

	p := &TCPProbe{
		port:    port,
		timeout: 2 * time.Second,
		tries:   5,
		dialer:  net.Dialer{Timeout: 2 * time.Second},
	}

	start := time.Now()
	_, err := p.Run(context.Background(), mustParseAddr(t, "127.0.0.1"))
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("Run() error = nil, want non-nil")
	}

	if strings.Contains(err.Error(), "tcp probe failed after") {
		t.Fatalf("Run() error = %q, want immediate underlying error", err.Error())
	}

	if elapsed > time.Second {
		t.Fatalf("Run() took %v, want fast failure", elapsed)
	}
}

func TestTCPProbeImplementsProbeInterface(t *testing.T) {
	t.Parallel()

	var _ interface {
		Schema() result.ResultSchema
		Init(context.Context) error
		Run(context.Context, netip.Addr) (result.Result, error)
		Close() error
	} = (*TCPProbe)(nil)
}

// timeoutNetError implements net.Error to simulate a network timeout.
type timeoutNetError struct{}

func (timeoutNetError) Error() string   { return "timeout" }
func (timeoutNetError) Timeout() bool   { return true }
func (timeoutNetError) Temporary() bool { return true }

// nonTimeoutNetError implements net.Error to simulate a non-timeout network error.
type nonTimeoutNetError struct{}

func (nonTimeoutNetError) Error() string   { return "non-timeout" }
func (nonTimeoutNetError) Timeout() bool   { return false }
func (nonTimeoutNetError) Temporary() bool { return true }

func TestIsTimeout(t *testing.T) {
	t.Parallel()

	if !netutil.IsTimeout(timeoutNetError{}) {
		t.Fatal("IsTimeout(timeoutNetError{}) = false, want true")
	}

	if netutil.IsTimeout(nonTimeoutNetError{}) {
		t.Fatal("IsTimeout(nonTimeoutNetError{}) = true, want false")
	}

	if netutil.IsTimeout(errors.New("plain error")) {
		t.Fatal("IsTimeout(plain error) = true, want false")
	}
}

func TestNewTCPProbePortOverflowFallsBackTo80(t *testing.T) {
	t.Parallel()

	p := NewTCPProbe(strconv.Itoa(70000), time.Second, 1)

	tcp, ok := p.(*TCPProbe)
	if !ok {
		t.Fatalf("NewTCPProbe() type = %T, want *TCPProbe", p)
	}

	if tcp.port != 80 {
		t.Fatalf("port = %d, want %d", tcp.port, 80)
	}
}
