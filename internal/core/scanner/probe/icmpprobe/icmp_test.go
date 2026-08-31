package icmpprobe

import (
	"context"
	"errors"
	"net"
	"net/netip"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

// fakeClock is a manually-advanced clock for deterministic timeout testing.
type fakeClock struct {
	mu      sync.Mutex
	current time.Time
	timers  []*fakeTimer
}

type fakeTimer struct {
	deadline time.Time
	c        chan time.Time
	stopped  atomic.Bool
}

func (t *fakeTimer) Stop() bool {
	return t.stopped.CompareAndSwap(false, true)
}

func newFakeClock(t time.Time) *fakeClock {
	return &fakeClock{current: t}
}

func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.current
}

func (c *fakeClock) NewTimer(d time.Duration) *time.Timer {
	c.mu.Lock()
	deadline := c.current.Add(d)
	ft := &fakeTimer{deadline: deadline, c: make(chan time.Time, 1)}
	c.timers = append(c.timers, ft)
	c.mu.Unlock()

	realTimer := time.NewTimer(24 * time.Hour)
	go func() {
		<-realTimer.C
	}()

	realTimer.Stop()
	return time.NewTimer(d)
}

func (c *fakeClock) advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.current = c.current.Add(d)
}

// fakeSocket is a scriptable in-memory ICMP socket for testing.
// By default, it automatically enqueues a valid echo reply on every WriteTo.
// Tests can disable this via autoReply or force errors via writeErr.
type fakeSocket struct {
	mu        sync.Mutex
	written   [][]byte
	lastDest  net.Addr
	replies   chan []byte
	autoReply bool
	protocol  int
	writeErr  error
	closed    bool
}

func newFakeSocket(protocol int) *fakeSocket {
	return &fakeSocket{
		replies:   make(chan []byte, 16),
		autoReply: true,
		protocol:  protocol,
	}
}

func (f *fakeSocket) enqueueReply(id, seq int) {
	var replyType icmp.Type
	if f.protocol == icmpProtocol {
		replyType = ipv4.ICMPTypeEchoReply
	} else {
		replyType = ipv6.ICMPTypeEchoReply
	}

	msg := icmp.Message{
		Type: replyType,
		Code: 0,
		Body: &icmp.Echo{ID: id, Seq: seq},
	}
	data, err := msg.Marshal(nil)
	if err != nil {
		panic("fakeSocket.enqueueReply: marshal failed: " + err.Error())
	}
	f.replies <- data
}

func (f *fakeSocket) WriteTo(b []byte, addr net.Addr) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.closed {
		return 0, net.ErrClosed
	}
	if f.writeErr != nil {
		return 0, f.writeErr
	}

	cp := make([]byte, len(b))
	copy(cp, b)
	f.written = append(f.written, cp)
	f.lastDest = addr

	if f.autoReply {
		msg, err := icmp.ParseMessage(f.protocol, b)
		if err == nil {
			if echo, ok := msg.Body.(*icmp.Echo); ok {
				f.enqueueReply(echo.ID, echo.Seq)
			}
		}
	}

	return len(b), nil
}

func (f *fakeSocket) ReadFrom(b []byte) (int, net.Addr, error) {
	select {
	case pkt, ok := <-f.replies:
		if !ok {
			return 0, nil, errors.New("socket closed")
		}
		n := copy(b, pkt)
		// Echo replies originate from the address the request was sent to;
		// handlePacket verifies this against the waiter's target IP.
		return n, f.lastDest, nil
	case <-time.After(readTimeout):
		return 0, nil, &errTimeout{}
	}
}

func (f *fakeSocket) SetReadDeadline(_ time.Time) error { return nil }

func (f *fakeSocket) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.closed = true
	return nil
}

func (f *fakeSocket) sentPackets() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.written)
}

// errTimeout implements net.Error to simulate a network timeout.
type errTimeout struct{}

func (e *errTimeout) Error() string   { return "i/o timeout" }
func (e *errTimeout) Timeout() bool   { return true }
func (e *errTimeout) Temporary() bool { return true }

const (
	testID4 = 1000
	testID6 = 1001
)

type probeFixture struct {
	probe *ICMPProbe
	sock4 *fakeSocket
	sock6 *fakeSocket
	clk   *fakeClock
}

// newFixture builds an ICMPProbe wired to fake sockets.
// ipv6Available controls whether the IPv6 socket setup succeeds.
func newFixture(t *testing.T, opts Options, ipv6Available bool) probeFixture {
	t.Helper()

	sock4 := newFakeSocket(icmpProtocol)
	sock6 := newFakeSocket(icmp6Protocol)

	clk := newFakeClock(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))

	factory := func(_, _, addr string) (socket, string, int, error) {
		if addr == "::" {
			if !ipv6Available {
				return nil, "", 0, errors.New("no IPv6")
			}
			return sock6, "udp", testID6, nil
		}
		return sock4, "udp", testID4, nil
	}

	opts.Clock = clk
	opts.Factory = factory
	if opts.Timeout == 0 {
		opts.Timeout = time.Second
	}
	if opts.Tries == 0 {
		opts.Tries = 3
	}

	p, err := NewICMPProbe(opts)
	if err != nil {
		t.Fatalf("NewICMPProbe: %v", err)
	}
	if err := p.Init(t.Context()); err != nil {
		t.Fatalf("Init: %v", err)
	}

	t.Cleanup(func() { _ = p.Close() })

	return probeFixture{probe: p, sock4: sock4, sock6: sock6, clk: clk}
}

func ipv4Addr() netip.Addr { return netip.MustParseAddr("192.0.2.1") }
func ipv6Addr() netip.Addr { return netip.MustParseAddr("2001:db8::1") }

// ipAddrFrom builds the net.Addr form of ip, as handlePacket would receive
// from a raw socket's ReadFrom.
func ipAddrFrom(ip netip.Addr) net.Addr {
	return &net.IPAddr{IP: net.IP(ip.AsSlice())}
}

func TestNew_FailsWhenIPv4Unavailable(t *testing.T) {
	_, err := NewICMPProbe(Options{
		Timeout: time.Second,
		Tries:   1,
		Factory: func(_, _, _ string) (socket, string, int, error) {
			return nil, "", 0, errors.New("permission denied")
		},
	})
	if err == nil {
		t.Fatal("expected error when IPv4 socket cannot be opened")
	}
}

func TestNew_IPv6UnavailableIsNonFatal(t *testing.T) {
	sock4 := newFakeSocket(icmpProtocol)
	_, err := NewICMPProbe(Options{
		Timeout: time.Second,
		Tries:   1,
		Factory: func(_, _, addr string) (socket, string, int, error) {
			if addr == "::" {
				return nil, "", 0, errors.New("no IPv6")
			}
			return sock4, "udp", testID4, nil
		},
	})
	if err != nil {
		t.Fatalf("expected success when only IPv6 is unavailable, got: %v", err)
	}
}

func TestRun_IPv4_SuccessOnFirstTry(t *testing.T) {
	fx := newFixture(t, Options{Tries: 3}, true)

	res, err := fx.probe.Run(t.Context(), ipv4Addr())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	r, ok := res.(ICMPResult)
	if !ok {
		t.Fatalf("unexpected result type %T", res)
	}
	if r.IP != ipv4Addr() {
		t.Errorf("IP: got %v, want %v", r.IP, ipv4Addr())
	}
	if r.Tries != 1 {
		t.Errorf("Tries: got %d, want 1", r.Tries)
	}
	if r.Mode != "udp" {
		t.Errorf("Mode: got %q, want %q", r.Mode, "udp")
	}
	if fx.sock4.sentPackets() != 1 {
		t.Errorf("expected 1 packet sent, got %d", fx.sock4.sentPackets())
	}
}

func TestRun_IPv6_SuccessOnFirstTry(t *testing.T) {
	fx := newFixture(t, Options{Tries: 3}, true)

	res, err := fx.probe.Run(t.Context(), ipv6Addr())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	r := res.(ICMPResult)
	if r.IP != ipv6Addr() {
		t.Errorf("IP: got %v, want %v", r.IP, ipv6Addr())
	}
	if r.Mode != "udp" {
		t.Errorf("Mode: got %q, want %q", r.Mode, "udp")
	}
	if fx.sock6.sentPackets() != 1 {
		t.Errorf("expected 1 packet sent on sock6, got %d", fx.sock6.sentPackets())
	}
}

func TestRun_IPv6_UnavailableReturnsError(t *testing.T) {
	fx := newFixture(t, Options{Tries: 1}, false)

	_, err := fx.probe.Run(t.Context(), ipv6Addr())
	if err == nil {
		t.Fatal("expected error for IPv6 target when IPv6 unavailable")
	}
	want := "IPv6 is not available on this system"
	if err.Error() != want {
		t.Errorf("error: got %q, want %q", err.Error(), want)
	}
}

func TestRun_ExhaustsRetriesReturnsLastError(t *testing.T) {
	fx := newFixture(t, Options{Tries: 3, Timeout: 5 * time.Millisecond}, true)

	fx.sock4.mu.Lock()
	fx.sock4.autoReply = false
	fx.sock4.mu.Unlock()

	_, err := fx.probe.Run(t.Context(), ipv4Addr())
	if err == nil {
		t.Fatal("expected error after exhausting retries")
	}
	if err.Error() != "timeout" {
		t.Errorf("error: got %q, want %q", err.Error(), "timeout")
	}
	if fx.sock4.sentPackets() != 3 {
		t.Errorf("expected 3 packets sent (one per try), got %d", fx.sock4.sentPackets())
	}
}

func TestRun_SucceedsOnSecondTry(t *testing.T) {
	fx := newFixture(t, Options{Tries: 3, Timeout: 5 * time.Millisecond}, true)

	var sent atomic.Int32
	sock4 := fx.sock4
	sock4.mu.Lock()
	sock4.autoReply = false
	sock4.mu.Unlock()

	go func() {
		for {
			sock4.mu.Lock()
			n := int32(len(sock4.written))
			sock4.mu.Unlock()

			if n > sent.Load() {
				pkt := sock4.written[n-1]
				msg, err := icmp.ParseMessage(icmpProtocol, pkt)
				if err == nil {
					echo := msg.Body.(*icmp.Echo)
					if n >= 2 {
						sock4.enqueueReply(echo.ID, echo.Seq)
					}
				}
				sent.Store(n)
			}
			time.Sleep(time.Millisecond)
		}
	}()

	res, err := fx.probe.Run(t.Context(), ipv4Addr())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	r := res.(ICMPResult)
	if r.Tries != 2 {
		t.Errorf("Tries: got %d, want 2", r.Tries)
	}
}

func TestRun_CancelledContextBeforeStart(t *testing.T) {
	fx := newFixture(t, Options{Tries: 3}, true)

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	_, err := fx.probe.Run(ctx, ipv4Addr())
	if !errors.Is(err, context.Canceled) {
		t.Errorf("error: got %v, want context.Canceled", err)
	}
}

func TestRun_CancelledContextDuringPing(t *testing.T) {
	fx := newFixture(t, Options{Tries: 3}, true)

	fx.sock4.mu.Lock()
	fx.sock4.autoReply = false
	fx.sock4.mu.Unlock()

	ctx, cancel := context.WithCancel(t.Context())

	go func() {
		for fx.sock4.sentPackets() == 0 {
			time.Sleep(time.Millisecond)
		}
		cancel()
	}()

	_, err := fx.probe.Run(ctx, ipv4Addr())
	if !errors.Is(err, context.Canceled) {
		t.Errorf("error: got %v, want context.Canceled", err)
	}
}

func TestRun_ProbeClosedDuringPing(t *testing.T) {
	fx := newFixture(t, Options{Tries: 3}, true)

	fx.sock4.mu.Lock()
	fx.sock4.autoReply = false
	fx.sock4.mu.Unlock()

	errCh := make(chan error, 1)
	go func() {
		_, err := fx.probe.Run(t.Context(), ipv4Addr())
		errCh <- err
	}()

	for fx.sock4.sentPackets() == 0 {
		time.Sleep(time.Millisecond)
	}
	_ = fx.probe.Close()

	err := <-errCh
	if err == nil {
		t.Fatal("expected error after probe closed")
	}
}

func TestRun_LatencyReflectsClockAdvance(t *testing.T) {
	const advance = 42 * time.Millisecond

	fx := newFixture(t, Options{Tries: 1}, true)

	fx.sock4.mu.Lock()
	fx.sock4.autoReply = false
	fx.sock4.mu.Unlock()

	resultCh := make(chan ICMPResult, 1)
	go func() {
		res, err := fx.probe.Run(t.Context(), ipv4Addr())
		if err != nil {
			return
		}
		resultCh <- res.(ICMPResult)
	}()

	for fx.sock4.sentPackets() == 0 {
		time.Sleep(time.Millisecond)
	}

	fx.clk.advance(advance)

	pkt := fx.sock4.written[0]
	msg, _ := icmp.ParseMessage(icmpProtocol, pkt)
	echo := msg.Body.(*icmp.Echo)
	fx.sock4.enqueueReply(echo.ID, echo.Seq)

	select {
	case r := <-resultCh:
		if r.Latency != advance {
			t.Errorf("Latency: got %v, want %v", r.Latency, advance)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for result")
	}
}

func TestHandlePacket_IPv4_SignalsWaiter(t *testing.T) {
	fx := newFixture(t, Options{Tries: 1}, true)
	p := fx.probe

	const id, seq = 999, 42
	key := makeKey(icmpProtocol, id, seq)
	ch := make(chan struct{}, 1)
	p.waiters.Store(key, &waiter{ch: ch, addr: ipv4Addr()})
	defer p.waiters.Delete(key)

	pkt, _ := (&icmp.Message{
		Type: ipv4.ICMPTypeEchoReply,
		Code: 0,
		Body: &icmp.Echo{ID: id, Seq: seq},
	}).Marshal(nil)

	p.handlePacket(pkt, icmpProtocol, ipAddrFrom(ipv4Addr()))

	select {
	case <-ch:
	default:
		t.Fatal("waiter channel was not signalled")
	}
}

func TestHandlePacket_IPv6_SignalsWaiter(t *testing.T) {
	fx := newFixture(t, Options{Tries: 1}, true)
	p := fx.probe

	const id, seq = 888, 7
	key := makeKey(icmp6Protocol, id, seq)
	ch := make(chan struct{}, 1)
	p.waiters.Store(key, &waiter{ch: ch, addr: ipv6Addr()})
	defer p.waiters.Delete(key)

	pkt, _ := (&icmp.Message{
		Type: ipv6.ICMPTypeEchoReply,
		Code: 0,
		Body: &icmp.Echo{ID: id, Seq: seq},
	}).Marshal(nil)

	p.handlePacket(pkt, icmp6Protocol, ipAddrFrom(ipv6Addr()))

	select {
	case <-ch:
	default:
		t.Fatal("waiter channel was not signalled")
	}
}

func TestHandlePacket_WrongType_Ignored(t *testing.T) {
	fx := newFixture(t, Options{Tries: 1}, true)
	p := fx.probe

	const id, seq = 777, 5
	key := makeKey(icmpProtocol, id, seq)
	ch := make(chan struct{}, 1)
	p.waiters.Store(key, &waiter{ch: ch, addr: ipv4Addr()})
	defer p.waiters.Delete(key)

	pkt, _ := (&icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{ID: id, Seq: seq},
	}).Marshal(nil)

	p.handlePacket(pkt, icmpProtocol, ipAddrFrom(ipv4Addr()))

	select {
	case <-ch:
		t.Fatal("waiter should not have been signalled for a non-reply packet")
	default:
	}
}

func TestHandlePacket_UnknownKey_NoSignal(t *testing.T) {
	fx := newFixture(t, Options{Tries: 1}, true)
	p := fx.probe

	pkt, _ := (&icmp.Message{
		Type: ipv4.ICMPTypeEchoReply,
		Code: 0,
		Body: &icmp.Echo{ID: 1, Seq: 1},
	}).Marshal(nil)

	p.handlePacket(pkt, icmpProtocol, ipAddrFrom(ipv4Addr()))
}

func TestHandlePacket_Garbage_Ignored(t *testing.T) {
	fx := newFixture(t, Options{Tries: 1}, true)
	p := fx.probe
	p.handlePacket([]byte{0xde, 0xad, 0xbe, 0xef}, icmpProtocol, ipAddrFrom(ipv4Addr()))
}

func TestHandlePacket_AddrMismatch_Dropped(t *testing.T) {
	fx := newFixture(t, Options{Tries: 1}, true)
	p := fx.probe

	const id, seq = 555, 9
	key := makeKey(icmpProtocol, id, seq)
	ch := make(chan struct{}, 1)
	p.waiters.Store(key, &waiter{ch: ch, addr: ipv4Addr()})
	defer p.waiters.Delete(key)

	pkt, _ := (&icmp.Message{
		Type: ipv4.ICMPTypeEchoReply,
		Code: 0,
		Body: &icmp.Echo{ID: id, Seq: seq},
	}).Marshal(nil)

	// Reply carries the registered (protocol, id, seq) key but originates
	// from a different host — the source-validation guard must drop it so
	// a colliding key can never signal the wrong waiter.
	p.handlePacket(pkt, icmpProtocol, ipAddrFrom(netip.MustParseAddr("192.0.2.2")))

	select {
	case <-ch:
		t.Fatal("waiter was signalled by a reply from a different source address")
	default:
	}
}

func TestMakeKey_Uniqueness(t *testing.T) {
	cases := [][3]int{
		{icmpProtocol, 1, 1},
		{icmpProtocol, 1, 2},
		{icmpProtocol, 2, 1},
		{icmpProtocol, 0xffff, 0xffff},
		{icmp6Protocol, 1, 1},
		{icmp6Protocol, 1, 2},
		{icmp6Protocol, 2, 1},
		{icmp6Protocol, 0xffff, 0xffff},
	}
	seen := map[uint64]bool{}
	for _, c := range cases {
		k := makeKey(c[0], c[1], c[2])
		if seen[k] {
			t.Errorf("collision for protocol=%d id=%d seq=%d", c[0], c[1], c[2])
		}
		seen[k] = true
	}

	// Identical id/seq across protocols must never share a key: an IPv6 reply
	// must not resolve to an IPv4 waiter (and vice versa).
	if makeKey(icmpProtocol, 1, 1) == makeKey(icmp6Protocol, 1, 1) {
		t.Error("IPv4 and IPv6 keys with identical id/seq must differ")
	}
}

func TestAddrMatches(t *testing.T) {
	v4 := ipv4Addr()
	v4other := netip.MustParseAddr("192.0.2.2")
	v6 := ipv6Addr()
	v4mapped := netip.MustParseAddr("::ffff:192.0.2.1")

	tests := []struct {
		name   string
		target netip.Addr
		from   net.Addr
		want   bool
	}{
		{"v4 IPAddr match", v4, &net.IPAddr{IP: net.IP(v4.AsSlice())}, true},
		{"v4 UDPAddr match", v4, &net.UDPAddr{IP: net.IP(v4.AsSlice())}, true},
		{"v4 matches IPv4-mapped form", v4, &net.IPAddr{IP: net.IP(v4mapped.AsSlice())}, true},
		{"v4 mismatch", v4, &net.IPAddr{IP: net.IP(v4other.AsSlice())}, false},
		{"v4 target vs v6 sender", v4, &net.IPAddr{IP: net.IP(v6.AsSlice())}, false},
		{"v6 match", v6, &net.IPAddr{IP: net.IP(v6.AsSlice())}, true},
		{"v6 target vs v4 sender", v6, &net.IPAddr{IP: net.IP(v4.AsSlice())}, false},
		{"unsupported addr type", v4, &net.TCPAddr{IP: net.IP(v4.AsSlice())}, false},
		{"nil addr", v4, nil, false},
		{"IPAddr with nil IP", v4, &net.IPAddr{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := addrMatches(tt.target, tt.from); got != tt.want {
				t.Errorf("addrMatches(%v, %T) = %v, want %v", tt.target, tt.from, got, tt.want)
			}
		})
	}
}

func TestDestination_UDPMode(t *testing.T) {
	ip := netip.MustParseAddr("10.0.0.1")
	addr := destination(ip, "udp")
	if _, ok := addr.(*net.UDPAddr); !ok {
		t.Errorf("expected *net.UDPAddr, got %T", addr)
	}
}

func TestDestination_RawMode(t *testing.T) {
	ip := netip.MustParseAddr("10.0.0.1")
	addr := destination(ip, "raw")
	if _, ok := addr.(*net.IPAddr); !ok {
		t.Errorf("expected *net.IPAddr, got %T", addr)
	}
}

func TestDestination_IPv4MappedIPv6_Unmapped(t *testing.T) {
	ip := netip.MustParseAddr("::ffff:192.0.2.1")
	addr := destination(ip, "raw")
	ipAddr := addr.(*net.IPAddr)
	if len(ipAddr.IP) != 4 {
		t.Errorf("expected 4-byte IP after Unmap, got %d bytes", len(ipAddr.IP))
	}
}

func TestClose_Idempotent(t *testing.T) {
	fx := newFixture(t, Options{Tries: 1}, true)

	if err := fx.probe.Close(); err != nil {
		t.Fatalf("first Close: %v", err)
	}
	if err := fx.probe.Close(); err != nil {
		t.Fatalf("second Close: %v", err)
	}
}

func TestClose_ReturnsSocketErrors(t *testing.T) {
	sock4 := newFakeSocket(icmpProtocol)
	errClose := errors.New("close failed")

	errSock := &errOnCloseSocket{fakeSocket: sock4, closeErr: errClose}

	p, err := NewICMPProbe(Options{
		Timeout: time.Second,
		Tries:   1,
		Factory: func(_, _, addr string) (socket, string, int, error) {
			if addr == "::" {
				return nil, "", 0, errors.New("no IPv6")
			}
			return errSock, "udp", testID4, nil
		},
	})
	if err != nil {
		t.Fatalf("NewICMPProbe: %v", err)
	}

	if err := p.Close(); !errors.Is(err, errClose) {
		t.Errorf("Close: got %v, want to contain %v", err, errClose)
	}
}

// errOnCloseSocket wraps fakeSocket to return a fixed error from Close.
type errOnCloseSocket struct {
	*fakeSocket
	closeErr error
}

func (s *errOnCloseSocket) Close() error { return s.closeErr }

func TestInit_Idempotent(t *testing.T) {
	fx := newFixture(t, Options{Tries: 1}, true)
	ctx := t.Context()

	for i := range 5 {
		if err := fx.probe.Init(ctx); err != nil {
			t.Fatalf("Init call %d: %v", i+1, err)
		}
	}
}

func TestSequenceNumber_Wraps16Bit(t *testing.T) {
	fx := newFixture(t, Options{Tries: 1}, true)
	p := fx.probe

	p.seq.Store(0xfffe)

	res, err := p.Run(t.Context(), ipv4Addr())
	if err != nil {
		t.Fatalf("Run at seq 0xffff: %v", err)
	}
	_ = res

	res, err = p.Run(t.Context(), ipv4Addr())
	if err != nil {
		t.Fatalf("Run at seq 0: %v", err)
	}
	_ = res
}
