package icmpprobe

import (
	"context"
	"errors"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"

	"bgscan/internal/core/result"
	"bgscan/internal/core/scanner/probe"
)

const (
	// icmpProtocol is the IP protocol number for ICMPv4.
	icmpProtocol = 1

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
// to measure reachability and latency for IPv4 targets.
//
// It maintains a single shared ICMP socket and a dedicated reader goroutine
// that demultiplexes echo replies back to waiting Ping callers.
type ICMPProbe struct {
	conn *icmp.PacketConn
	mode string

	// id is the ICMP identifier for echo requests, derived from the process ID.
	id int

	// seq is an atomically incremented sequence number used to construct
	// unique (id, seq) pairs for matching requests and replies.
	seq uint32

	timeout time.Duration
	tries   uint16

	// waiters holds per-request channels keyed by (id, seq).
	// Each active Ping registers a channel here and waits for a reply signal.
	waiters sync.Map

	done      chan struct{}
	closeOnce sync.Once
	startOnce sync.Once
}

// NewICMPProbe creates a new ICMPProbe. It attempts to open a raw ICMP socket
// ("ip4:icmp") and falls back to a UDP-based listener ("udp4") if raw sockets
// are not permitted (e.g., due to OS permissions).
func NewICMPProbe(timeout time.Duration, timesTry uint16) (probe.Probe, error) {
	conn, mode, id, err := openICMPSocket()
	if err != nil {
		return nil, err
	}

	return &ICMPProbe{
		conn:    conn,
		mode:    mode,
		id:      id,
		timeout: timeout,
		tries:   timesTry,
		done:    make(chan struct{}),
	}, nil
}

// Schema returns the result schema for ICMP probes.
func (p *ICMPProbe) Schema() result.ResultSchema {
	return Schema
}

// Init implements [probe.Probe] and starts the background reader goroutine
// on first invocation.
func (p *ICMPProbe) Init(_ context.Context) error {
	p.startOnce.Do(func() {
		go p.reader()
	})
	return nil
}

// openICMPSocket attempts to create an ICMP-capable socket. It first tries a
// raw ICMP socket ("ip4:icmp") and falls back to "udp4" if raw sockets are
// not permitted.
func openICMPSocket() (*icmp.PacketConn, string, int, error) {
	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err == nil {
		return conn, "raw", os.Getpid() & 0xffff, nil
	}

	conn, err = icmp.ListenPacket("udp4", "0.0.0.0")
	if err != nil {
		return nil, "", 0, err
	}

	id := conn.LocalAddr().(*net.UDPAddr).Port
	return conn, "udp", id, nil
}

// reader is the background goroutine responsible for consuming incoming
// ICMP packets from the shared socket.
func (p *ICMPProbe) reader() {
	buf := make([]byte, maxPacket)

	for {
		select {
		case <-p.done:
			return
		default:
		}

		_ = p.conn.SetReadDeadline(time.Now().Add(readTimeout))

		n, _, err := p.conn.ReadFrom(buf)
		if err != nil {
			if isTimeout(err) {
				continue
			}
			return
		}

		p.handlePacket(buf[:n])
	}
}

// handlePacket parses a single ICMP packet and, if it is an Echo Reply
// matching an active Ping request, notifies the corresponding waiter channel.
func (p *ICMPProbe) handlePacket(packet []byte) {
	msg, err := icmp.ParseMessage(icmpProtocol, packet)
	if err != nil || msg.Type != ipv4.ICMPTypeEchoReply {
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
// waits for a corresponding echo reply or timeout.
func (p *ICMPProbe) Ping(ctx context.Context, ip string, timeout time.Duration) error {
	dstIP := net.ParseIP(ip)
	if dstIP == nil {
		return errors.New("invalid ip")
	}

	seq := int(atomic.AddUint32(&p.seq, 1))
	key := makeKey(p.id, seq)

	ch := make(chan struct{}, 1)
	p.waiters.Store(key, ch)
	defer p.waiters.Delete(key)

	msg := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   p.id,
			Seq:  seq,
			Data: []byte(payload),
		},
	}

	data, err := msg.Marshal(nil)
	if err != nil {
		return err
	}

	if _, err = p.conn.WriteTo(data, p.destination(dstIP)); err != nil {
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

// destination returns the appropriate net.Addr for the current socket mode.
func (p *ICMPProbe) destination(ip net.IP) net.Addr {
	if p.mode == "udp" {
		return &net.UDPAddr{IP: ip}
	}
	return &net.IPAddr{IP: ip}
}

// Run implements [probe.Probe] and performs an ICMP-based reachability check
// for the given IP address.
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

		return ICMPResult{
			IP:      ip,
			Latency: time.Since(start),
			Tries:   i + 1,
			Mode:    p.mode,
		}, nil
	}

	return nil, lastErr
}

// Close terminates the ICMPProbe and releases associated resources.
func (p *ICMPProbe) Close() error {
	var err error

	p.closeOnce.Do(func() {
		close(p.done)
		if p.conn != nil {
			err = p.conn.Close()
		}
	})

	return err
}

// isTimeout reports whether the provided error represents a network timeout.
func isTimeout(err error) bool {
	if ne, ok := err.(net.Error); ok {
		return ne.Timeout()
	}
	return false
}
