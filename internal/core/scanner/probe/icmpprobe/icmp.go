package icmpprobe

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"

	"bgscan/internal/core/result"
	"bgscan/internal/core/scanner/probe"
)

const (
	// icmpProtocol is the IP protocol number for ICMPv4.
	icmpProtocol = 1

	// icmp6Protocol is the IP protocol number for ICMPv6.
	icmp6Protocol = 58

	// maxPacket is the maximum size of an incoming ICMP packet buffer.
	maxPacket = 4096

	// readTimeout is the per-iteration read deadline used by the reader loop.
	// Short timeouts ensure the goroutine wakes up frequently to check for shutdown.
	readTimeout = 200 * time.Millisecond

	// payload is the optional ICMP echo payload.
	// Kept empty to minimize packet size for scanning workloads.
	payload = ""
)

// ICMPProbe implements the [probe.Probe] interface using ICMP echo requests
// to measure reachability and latency for IPv4 and IPv6 targets.
//
// It maintains two shared ICMP sockets (one per IP family) and a dedicated
// reader goroutine per socket that demultiplexes echo replies back to waiting
// Ping callers. IPv6 support is best-effort: if the host has no IPv6 capability,
// conn6 is nil and IPv6 targets return an error instead of crashing.
type ICMPProbe struct {
	// IPv4 socket state
	conn4 *icmp.PacketConn
	mode4 string
	id4   int

	// IPv6 socket state (nil if unavailable on this system)
	conn6 *icmp.PacketConn
	mode6 string
	id6   int

	// seq is an atomically incremented sequence number shared across both
	// sockets. Masked to 16-bit before use to match the ICMP wire format.
	seq atomic.Uint32

	timeout time.Duration
	tries   uint16

	// waiters holds per-request channels keyed by (id, seq).
	// Both the v4 and v6 readers signal into this shared map.
	// Collisions between v4 and v6 are impossible because id4 != id6.
	waiters sync.Map

	done      chan struct{}
	closeOnce sync.Once
	startOnce sync.Once
}

// NewICMPProbe creates a new ICMPProbe. It always opens an IPv4 socket and
// attempts to open an IPv6 socket as well. IPv6 failure is non-fatal: targets
// that resolve to IPv6 addresses will return an error at Ping time.
func NewICMPProbe(timeout time.Duration, timesTry uint16) (probe.Probe, error) {
	conn4, mode4, id4, err := openSocket("ip4:icmp", "udp4", "0.0.0.0")
	if err != nil {
		return nil, err
	}

	// IPv6 is best-effort. A nil conn6 means "not available on this system".
	conn6, mode6, id6, _ := openSocket("ip6:ipv6-icmp", "udp6", "::")

	return &ICMPProbe{
		conn4:   conn4,
		mode4:   mode4,
		id4:     id4,
		conn6:   conn6,
		mode6:   mode6,
		id6:     id6,
		timeout: timeout,
		tries:   timesTry,
		done:    make(chan struct{}),
	}, nil
}

// Schema returns the result schema for ICMP probes.
func (p *ICMPProbe) Schema() result.ResultSchema {
	return Schema
}

// Init implements [probe.Probe] and starts the background reader goroutines
// on first invocation. One goroutine is started per open socket.
func (p *ICMPProbe) Init(_ context.Context) error {
	p.startOnce.Do(func() {
		go p.reader(p.conn4, icmpProtocol)
		if p.conn6 != nil {
			go p.reader(p.conn6, icmp6Protocol)
		}
	})
	return nil
}

// openSocket attempts to create an ICMP-capable socket. It first tries the
// privileged network type (e.g. "ip4:icmp") and falls back to the unprivileged
// UDP type (e.g. "udp4") if raw sockets are not permitted.
func openSocket(privileged, unprivileged, addr string) (*icmp.PacketConn, string, int, error) {
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

// reader is the background goroutine responsible for consuming incoming
// ICMP packets from a single socket. protocol must match the socket family:
// icmpProtocol (1) for IPv4, icmp6Protocol (58) for IPv6.
func (p *ICMPProbe) reader(conn *icmp.PacketConn, protocol int) {
	buf := make([]byte, maxPacket)

	for {
		select {
		case <-p.done:
			return
		default:
		}

		_ = conn.SetReadDeadline(time.Now().Add(readTimeout))

		n, _, err := conn.ReadFrom(buf)
		if err != nil {
			if isTimeout(err) {
				continue
			}
			return
		}

		p.handlePacket(buf[:n], protocol)
	}
}

// handlePacket parses a single ICMP packet and, if it is an Echo Reply
// matching an active Ping request, notifies the corresponding waiter channel.
// protocol selects which ICMP message type to expect as a reply.
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

// makeKey composes a 64-bit key from ICMP identifier and sequence number.
func makeKey(id, seq int) uint64 {
	return uint64(id)<<32 | uint64(seq)
}

// Ping sends a single ICMP echo request to the given IP address and
// waits for a corresponding echo reply or timeout. It selects the correct
// socket (v4 or v6) based on the target address family.
func (p *ICMPProbe) Ping(ctx context.Context, ip string, timeout time.Duration) error {
	dstIP := net.ParseIP(ip)
	if dstIP == nil {
		return errors.New("invalid ip")
	}

	// Select socket, id, and mode based on IP family.
	var (
		conn  *icmp.PacketConn
		id    int
		mode  string
		proto int
	)

	if dstIP.To4() != nil {
		// IPv4 target
		conn = p.conn4
		id = p.id4
		mode = p.mode4
		proto = icmpProtocol
	} else {
		// IPv6 target
		if p.conn6 == nil {
			return errors.New("IPv6 is not available on this system")
		}
		conn = p.conn6
		id = p.id6
		mode = p.mode6
		proto = icmp6Protocol
	}

	// Mask to 16-bit to match the ICMP wire format (Seq field is uint16).
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

	if _, err = conn.WriteTo(data, destination(dstIP, mode)); err != nil {
		return err
	}

	timer := time.NewTimer(timeout)
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

// destination returns the appropriate net.Addr for the given IP and socket mode.
func destination(ip net.IP, mode string) net.Addr {
	if mode == "udp" {
		return &net.UDPAddr{IP: ip}
	}
	return &net.IPAddr{IP: ip}
}

// Run implements [probe.Probe] and performs an ICMP-based reachability check
// for the given IP address. It handles both IPv4 and IPv6 targets transparently.
//
// It executes up to p.tries Ping attempts, each using the Probe's configured
// timeout. It returns an ICMPResult on the first successful ping.
func (p *ICMPProbe) Run(ctx context.Context, ip string) (result.Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var lastErr error

	for i := 0; i < int(p.tries); i++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		start := time.Now()

		if err := p.Ping(ctx, ip, p.timeout); err != nil {
			lastErr = err
			continue
		}

		// Determine which mode to report based on the target IP family.
		reportMode := p.mode4
		if addr := net.ParseIP(ip); addr != nil && addr.To4() == nil {
			reportMode = p.mode6
		}

		return ICMPResult{
			IP:      ip,
			Latency: time.Since(start),
			Tries:   i + 1,
			Mode:    reportMode,
		}, nil
	}

	return nil, lastErr
}

// Close terminates the ICMPProbe and releases all associated resources.
// Both the IPv4 and IPv6 sockets are closed if open.
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

// isTimeout reports whether the provided error represents a network timeout.
func isTimeout(err error) bool {
	if ne, ok := err.(net.Error); ok {
		return ne.Timeout()
	}
	return false
}
