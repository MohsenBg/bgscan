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

	"github.com/MohsenBg/bgscan/internal/core/netutil"
	"github.com/MohsenBg/bgscan/internal/core/result"
	"github.com/MohsenBg/bgscan/internal/core/scanner/probe"
	"github.com/MohsenBg/bgscan/internal/core/target"
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

// waiter tracks an in-flight Ping awaiting a reply. addr pins the target IP
// because the (protocol, id, seq) key can collide past ~65536 concurrent
// pings per protocol (ICMP sequence numbers are 16 bits on the wire).
type waiter struct {
	ch   chan struct{}
	addr netip.Addr
}

// ICMPProbe measures reachability and latency over shared ICMP sockets with
// dedicated reader goroutines. IPv6 is best-effort; without it, IPv6
// targets fail.
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

	waiters   sync.Map // key: uint64 from makeKey -> *waiter
	done      chan struct{}
	closeOnce sync.Once
	startOnce sync.Once
}

// Options configures an ICMPProbe.
type Options struct {
	Timeout time.Duration
	Tries   uint16
	Clock   clock         // Optional, for testing.
	Factory socketFactory // Optional, for testing.
}

// NewICMPProbe creates an ICMPProbe, filling in real Clock/Factory when
// opts leaves them nil.
func NewICMPProbe(opts Options) (*ICMPProbe, error) {
	if opts.Clock == nil {
		opts.Clock = realClock{}
	}
	if opts.Factory == nil {
		opts.Factory = defaultFactory
	}

	conn4, mode4, id4, err := opts.Factory("ip4:icmp", "udp4", "0.0.0.0")
	if err != nil {
		return nil, fmt.Errorf("create IPv4 ICMP socket: %w", err)
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

func (p *ICMPProbe) Schema() result.ResultSchema {
	return Schema
}

// Init starts the background reader goroutines (once).
func (p *ICMPProbe) Init(_ context.Context) error {
	p.startOnce.Do(func() {
		go p.reader(p.conn4, icmpProtocol)
		if p.conn6 != nil {
			go p.reader(p.conn6, icmp6Protocol)
		}
	})
	return nil
}

// socketFor returns the connection, echo ID, mode and protocol for ip's
// address family. IPv6 targets error when no IPv6 socket is available.
func (p *ICMPProbe) socketFor(ip netip.Addr) (conn socket, id int, mode string, proto int, err error) {
	if ip.Is4() {
		return p.conn4, p.id4, p.mode4, icmpProtocol, nil
	}
	if p.conn6 == nil {
		return nil, 0, "", 0, errors.New("IPv6 is not available on this system")
	}
	return p.conn6, p.id6, p.mode6, icmp6Protocol, nil
}

// modeFor reports the socket mode ("raw" or "udp") for ip's address family.
func (p *ICMPProbe) modeFor(ip netip.Addr) string {
	if ip.Is4() {
		return p.mode4
	}
	return p.mode6
}

// reader demultiplexes incoming ICMP replies to waiting Ping callers.
func (p *ICMPProbe) reader(conn socket, protocol int) {
	buf := make([]byte, maxPacket)

	for {
		select {
		case <-p.done:
			return
		default:
		}

		_ = conn.SetReadDeadline(p.clock.Now().Add(readTimeout))

		n, addr, err := conn.ReadFrom(buf)
		if err != nil {
			if netutil.IsTimeout(err) {
				continue
			}
			return
		}

		p.handlePacket(buf[:n], protocol, addr)
	}
}

// handlePacket signals the waiter matching an Echo Reply. The (protocol,
// id, seq) key narrows candidates, but ICMP sequence numbers are only 16
// bits on the wire, so two in-flight targets can legitimately share a key
// past ~65536 concurrent pings per protocol. The packet source must also
// match the waiter's target before it counts as a genuine reply.
func (p *ICMPProbe) handlePacket(packet []byte, protocol int, from net.Addr) {
	msg, err := icmp.ParseMessage(protocol, packet)
	if err != nil {
		return
	}

	var replyType icmp.Type
	switch protocol {
	case icmpProtocol:
		replyType = ipv4.ICMPTypeEchoReply
	case icmp6Protocol:
		replyType = ipv6.ICMPTypeEchoReply
	default:
		return
	}

	if msg.Type != replyType {
		return
	}

	body, ok := msg.Body.(*icmp.Echo)
	if !ok {
		return
	}

	key := makeKey(protocol, body.ID, body.Seq)

	v, ok := p.waiters.Load(key)
	if !ok {
		return
	}

	w := v.(*waiter)

	// Key collision (sequence wraparound under very high concurrency):
	// the reply belongs to a different in-flight target.
	if !addrMatches(w.addr, from) {
		return
	}

	select {
	case w.ch <- struct{}{}:
	default:
	}
}

// addrMatches reports whether the sender address of an ICMP reply corresponds to the given target IP.
func addrMatches(target netip.Addr, from net.Addr) bool {
	var fromIP net.IP

	switch a := from.(type) {
	case *net.IPAddr:
		fromIP = a.IP
	case *net.UDPAddr:
		fromIP = a.IP
	default:
		return false
	}

	got, ok := netip.AddrFromSlice(fromIP)
	if !ok {
		return false
	}

	return got.Unmap() == target.Unmap()
}

// makeKey packs protocol, ID and sequence into 64 bits. The protocol bits
// keep an IPv4 waiter from matching an IPv6 reply with the same ID and
// sequence (or vice versa).
func makeKey(protocol, id, seq int) uint64 {
	return uint64(protocol)<<48 | uint64(id)<<32 | uint64(seq)
}

// Ping sends a single ICMP echo request to the target IP and waits for a reply or timeout.
func (p *ICMPProbe) Ping(ctx context.Context, ip netip.Addr, timeout time.Duration) error {
	conn, id, mode, proto, err := p.socketFor(ip)
	if err != nil {
		return err
	}

	seq := int(p.seq.Add(1) & 0xffff)
	key := makeKey(proto, id, seq)

	w := &waiter{ch: make(chan struct{}, 1), addr: ip}
	p.waiters.Store(key, w)
	defer p.waiters.Delete(key)

	var msgType icmp.Type = ipv4.ICMPTypeEcho
	if proto == icmp6Protocol {
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
		return fmt.Errorf("marshal icmp message: %w", err)
	}

	if _, err = conn.WriteTo(data, destination(ip, mode)); err != nil {
		return fmt.Errorf("icmp write: %w", err)
	}

	timer := p.clock.NewTimer(timeout)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-p.done:
		return errors.New("icmp probe closed")
	case <-w.ch:
		return nil
	case <-timer.C:
		return probe.ErrTimeout
	}
}

// destination adapts the target IP to the socket mode (raw vs UDP).
func destination(ip netip.Addr, mode string) net.Addr {
	stdIP := net.IP(ip.Unmap().AsSlice())
	if mode == "udp" {
		return &net.UDPAddr{IP: stdIP}
	}
	return &net.IPAddr{IP: stdIP}
}

func (*ICMPProbe) Input() target.Kind {
	return target.KindIP
}

// Run pings the target address, retrying up to the configured Tries limit.
// Domain targets are refused: ICMP needs an address, and nothing resolves
// names yet.
func (p *ICMPProbe) Run(ctx context.Context, entry target.Entry) (result.Result, error) {
	if entry.Kind != target.KindIP {
		return nil, fmt.Errorf("%w: %s", probe.ErrDomainTarget, entry.Domain)
	}

	ip := entry.IP.Addr()
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

		return ICMPResult{
			Entry:   entry,
			Latency: p.clock.Now().Sub(start),
			Tries:   i + 1,
			Mode:    p.modeFor(ip),
		}, nil
	}

	return nil, probe.NormalizeErr(lastErr)
}

// Close stops the background readers and closes the ICMP sockets.
func (p *ICMPProbe) Close() error {
	var errs []error

	p.closeOnce.Do(func() {
		close(p.done)

		closeSocket := func(name string, conn socket) {
			if conn == nil {
				return
			}
			if err := conn.Close(); err != nil {
				errs = append(errs, fmt.Errorf("close %s ICMP socket: %w", name, err))
			}
		}
		closeSocket("IPv4", p.conn4)
		closeSocket("IPv6", p.conn6)
	})

	return errors.Join(errs...)
}

// defaultFactory opens a raw ICMP socket, falling back to unprivileged UDP
// when permissions are denied.
func defaultFactory(privileged, unprivileged, addr string) (socket, string, int, error) {
	conn, err := icmp.ListenPacket(privileged, addr)
	if err == nil {
		return conn, "raw", os.Getpid() & 0xffff, nil
	}

	conn, err = icmp.ListenPacket(unprivileged, addr)
	if err != nil {
		return nil, "", 0, fmt.Errorf("listen icmp socket: %w", err)
	}

	id := conn.LocalAddr().(*net.UDPAddr).Port
	return conn, "udp", id, nil
}
