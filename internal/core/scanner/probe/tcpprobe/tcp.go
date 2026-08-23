package tcpprobe

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"strconv"
	"time"

	"bgscan/internal/core/netutil"
	"bgscan/internal/core/result"
	"bgscan/internal/core/scanner/probe"
	"bgscan/internal/logger"
)

// TCPProbe verifies Layer-4 reachability and measures handshake RTT without
// making assumptions about the application-layer protocol.
type TCPProbe struct {
	port    uint16
	timeout time.Duration
	dialer  net.Dialer
	tries   uint16
}

// NewTCPProbe creates a TCPProbe targeting the specified port.
// If the port string is invalid or out of range, it defaults to 80.
func NewTCPProbe(port string, timeout time.Duration, tries uint16) probe.Probe {
	p, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		p = 80
	}

	return &TCPProbe{
		port:    uint16(p),
		tries:   tries,
		timeout: timeout,
		dialer: net.Dialer{
			Timeout: timeout,
		},
	}
}

// Schema returns the result schema for TCP probes.
func (p *TCPProbe) Schema() result.ResultSchema {
	return Schema
}

// Init implements probe.Probe. It is a no-op since TCPProbe is stateless.
func (p *TCPProbe) Init(_ context.Context) error {
	return nil
}

// Run implements probe.Probe. It attempts up to `tries` TCP handshakes to the
// target IP. Timeouts are retried, while other errors (e.g., connection refused)
// fail immediately to preserve scanning throughput. On success, it returns a
// TCPResult with the latency of the first successful attempt.
func (p *TCPProbe) Run(ctx context.Context, ip netip.Addr) (result.Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	address := net.JoinHostPort(ip.String(), strconv.FormatUint(uint64(p.port), 10))
	var lastErr error

	for i := 0; i < int(p.tries); i++ {
		start := time.Now()

		conn, err := p.dialer.DialContext(ctx, "tcp", address)
		if err != nil {
			lastErr = err
			if netutil.IsTimeout(err) {
				continue
			}
			return nil, err
		}

		rtt := time.Since(start)

		if err := conn.Close(); err != nil {
			logger.CoreError("error closing connection: %v", err)
		}

		return TCPResult{
			IP:      ip,
			Port:    p.port,
			Latency: rtt,
			Tries:   i + 1,
		}, nil
	}

	return nil, fmt.Errorf("tcp probe failed after %d tries: %w", p.tries, lastErr)
}

// Close implements probe.Probe. It is a no-op, as connections are released
// immediately after each attempt within Run.
func (p *TCPProbe) Close() error {
	return nil
}


