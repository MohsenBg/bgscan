package icmpprobe

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"

	"bgscan/internal/core/netutil"
	"bgscan/internal/core/result"
)

const (
	icmpProtocol  = 1  // IP protocol number for ICMPv4.
	icmp6Protocol = 58 // IP protocol number for ICMPv6.
	maxPacket     = 4096
	readTimeout   = 200 * time.Millisecond // Short timeout ensures the reader loop checks for shutdown frequently.
	payload       = ""                     // Empty payload minimizes packet size for scanning workloads.
)

// socket abstracts an ICMP packet connection, primarily to allow mocking in tests.
type socket interface {
	WriteTo(b []byte, addr net.Addr) (int, error)
	ReadFrom(b []byte) (int, net.Addr, error)
	SetReadDeadline(t time.Time) error
	Close() error
}

// clock abstracts time operations for testing.
type clock interface {
	Now() time.Time
	NewTimer(d time.Duration) *time.Timer
}

type realClock struct{}

func (realClock) Now() time.Time                       { return time.Now() }
func (realClock) NewTimer(d time.Duration) *time.Timer { return time.NewTimer(d) }

// socketFactory abstracts socket creation for testing.
type socketFactory func(privileged, unprivileged, addr string) (socket, string, int, error)

// ICMPProbe measures reachability and latency for IPv4 and IPv6 targets using ICMP echo requests.
// It maintains shared ICMP sockets and dedicated reader goroutines to demultiplex echo replies.
// IPv6 support is best-effort; if unavailable, IPv6 targets return an error.
type ICMPProbe struct {
	conn4 socket
	mode4 string
	id4   int

	conn6 socket // nil if unavailable
	mode6 string
	id6   int

	seq     atomic.Uint32
	timeout time.Duration
	tries   uint16
	clock   clock

	waiters   sync.Map
	done      chan struct{}
	closeOnce sync.Once
	startOnce sync.Once
}

// Options configures the behavior of an ICMPProbe.
type Options struct {
	Timeout time.Duration // Per-ping timeout.
	Tries   uint16        // Maximum number of ping attempts.
	Clock   clock         // Optional clock interface for testing.
	Factory socketFactory // Optional socket factory for testing.
}

// NewICMPProbe creates a new ICMPProbe. If opts.Clock or opts.Factory are nil, real implementations are used.
func NewICMPProbe(opts Options) (*ICMPProbe, error) {
	if opts.Clock == nil {
		opts.Clock = realClock{}
	}
	if opts.Factory == nil {
		opts.Factory = defaultFactory
	}

	conn4, mode4, id4, err := opts.Factory("ip4:icmp", "udp4", "0.0.0.0")
	if err != nil {
		return nil, err
	}

	conn6, mode6, id6, _ := opts.Factory("ip6:ipv6-icmp", "udp6", "::")

	return &ICMPProbe{
		conn4:   conn4,
		mode4:   mode4,
		id4:     id4,
		conn6:   conn6,
		mode6:   mode6,
		id6:     id6,
		timeout: opts.Timeout,
		tries:   opts.Tries,
		clock:   opts.Clock,
		done:    make(chan struct{}),
	}, nil
}

// Schema returns the result schema for ICMP probes.
func (p *ICMPProbe) Schema() result.ResultSchema {
	return Schema
}

// Init implements probe.Probe. It starts the background reader goroutines on first invocation.
func (p *ICMPProbe) Init(_ context.Context) error {
	p.startOnce.Do(func() {
		go p.reader(p.conn4, icmpProtocol)
		if p.conn6 != nil {
			go p.reader(p.conn6, icmp6Protocol)
		}
	})
	return nil
}

// reader consumes incoming ICMP packets from a socket and demultiplexes replies to waiting Ping callers.
func (p *ICMPProbe) reader(conn socket, protocol int) {
	buf := make([]byte, maxPacket)

	for {
		select {
		case <-p.done:
			return
		default:
		}

		_ = conn.SetReadDeadline(p.clock.Now().Add(readTimeout))

		n, _, err := conn.ReadFrom(buf)
		if err != nil {
			if netutil.IsTimeout(err) {
				continue
			}
			return
		}

		p.handlePacket(buf[:n], protocol)
	}
}

// handlePacket parses an incoming ICMP packet and signals the corresponding waiter if it matches an active Echo Reply.
func (p *ICMPProbe) handlePacket(packet []byte, protocol int) {
	msg, err := icmp.ParseMessage(protocol, packet)
	if err != nil {
		return
	}

	switch protocol {
	case icmpProtocol:
		if msg.Type != ipv4.ICMPTypeEchoReply {
			return
		}
	case icmp6Protocol:
		if msg.Type != ipv6.ICMPTypeEchoReply {
			return
		}
	default:
		return
	}

	body, ok := msg.Body.(*icmp.Echo)
	if !ok {
		return
	}

	key := makeKey(body.ID, body.Seq)

	if ch, ok := p.waiters.Load(key); ok {
		select {
		case ch.(chan struct{}) <- struct{}{}:
		default:
		}
	}
}

// makeKey generates a unique 64-bit identifier from an ICMP ID and sequence number.
func makeKey(id, seq int) uint64 {
	return uint64(id)<<32 | uint64(seq)
}

// Ping sends a single ICMP echo request to the target IP and waits for a reply or timeout.
func (p *ICMPProbe) Ping(ctx context.Context, ip netip.Addr, timeout time.Duration) error {
	var (
		conn  socket
		id    int
		mode  string
		proto int
	)

	if ip.Is4() {
		conn = p.conn4
		id = p.id4
		mode = p.mode4
		proto = icmpProtocol
	} else {
		if p.conn6 == nil {
			return errors.New("IPv6 is not available on this system")
		}
		conn = p.conn6
		id = p.id6
		mode = p.mode6
		proto = icmp6Protocol
	}

	seq := int(p.seq.Add(1) & 0xffff)
	key := makeKey(id, seq)

	ch := make(chan struct{}, 1)
	p.waiters.Store(key, ch)
	defer p.waiters.Delete(key)

	var msgType icmp.Type
	if proto == icmpProtocol {
		msgType = ipv4.ICMPTypeEcho
	} else {
		msgType = ipv6.ICMPTypeEchoRequest
	}

	msg := icmp.Message{
		Type: msgType,
		Code: 0,
		Body: &icmp.Echo{
			ID:   id,
			Seq:  seq,
			Data: []byte(payload),
		},
	}

	data, err := msg.Marshal(nil)
	if err != nil {
		return err
	}

	if _, err = conn.WriteTo(data, destination(ip, mode)); err != nil {
		return err
	}

	timer := p.clock.NewTimer(timeout)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-p.done:
		return errors.New("icmp probe closed")
	case <-ch:
		return nil
	case <-timer.C:
		return errors.New("timeout")
	}
}

// destination returns the appropriate net.Addr for the target IP, adapting to raw or UDP socket modes.
func destination(ip netip.Addr, mode string) net.Addr {
	stdIP := net.IP(ip.Unmap().AsSlice())
	if mode == "udp" {
		return &net.UDPAddr{IP: stdIP}
	}
	return &net.IPAddr{IP: stdIP}
}

// Run implements probe.Probe. It performs an ICMP reachability check, retrying up to the configured Tries limit on failure.
func (p *ICMPProbe) Run(ctx context.Context, ip netip.Addr) (result.Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var lastErr error

	for i := 0; i < int(p.tries); i++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		start := p.clock.Now()

		if err := p.Ping(ctx, ip, p.timeout); err != nil {
			lastErr = err
			continue
		}

		reportMode := p.mode4
		if !ip.Is4() {
			reportMode = p.mode6
		}

		return ICMPResult{
			IP:      ip,
			Latency: p.clock.Now().Sub(start),
			Tries:   i + 1,
			Mode:    reportMode,
		}, nil
	}

	return nil, lastErr
}

// Close implements probe.Probe. It terminates the background readers and closes the ICMP sockets.
func (p *ICMPProbe) Close() error {
	var errs []error

	p.closeOnce.Do(func() {
		close(p.done)

		if p.conn4 != nil {
			if err := p.conn4.Close(); err != nil {
				errs = append(errs, fmt.Errorf("close IPv4 ICMP socket: %w", err))
			}
		}

		if p.conn6 != nil {
			if err := p.conn6.Close(); err != nil {
				errs = append(errs, fmt.Errorf("close IPv6 ICMP socket: %w", err))
			}
		}
	})

	return errors.Join(errs...)
}

// defaultFactory attempts to open a raw ICMP socket, falling back to an unprivileged UDP socket if permissions are denied.
func defaultFactory(privileged, unprivileged, addr string) (socket, string, int, error) {
	conn, err := icmp.ListenPacket(privileged, addr)
	if err == nil {
		return conn, "raw", os.Getpid() & 0xffff, nil
	}

	conn, err = icmp.ListenPacket(unprivileged, addr)
	if err != nil {
		return nil, "", 0, err
	}

	id := conn.LocalAddr().(*net.UDPAddr).Port
	return conn, "udp", id, nil
}

