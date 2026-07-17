package tcpprobe

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	"bgscan/internal/core/result"
	"bgscan/internal/core/scanner/probe"
	"bgscan/internal/logger"
)

// TCPProbe performs a lightweight TCP connectivity probe against a target IP.
// It verifies Layer-4 reachability and measures handshake RTT without making
// assumptions about the application-layer protocol.
type TCPProbe struct {
	port    uint16
	timeout time.Duration
	dialer  net.Dialer
	tries   uint16
}

// NewTCPProbe creates a TCPProbe targeting the given port. If port is not a
// valid uint16, 80 is used as a fallback. timeout bounds each DialContext
// attempt and tries caps the number of attempts.
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

// Schema returns the TCP result schema.
func (p *TCPProbe) Schema() result.ResultSchema {
	return Schema
}

// Init is a no-op; TCPProbe holds no persistent state.
func (p *TCPProbe) Init(_ context.Context) error {
	return nil
}

// Run attempts up to p.tries TCP handshakes to ip:port. A timeout on an attempt
// is retried; any other error (e.g. connection refused) fails fast to keep
// scanning throughput high. On success, returns a TCPResult with the first
// successful attempt's latency; on failure, returns an error wrapping the last
// error seen.
func (p *TCPProbe) Run(ctx context.Context, ip string) (result.Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	address := net.JoinHostPort(ip, strconv.FormatUint(uint64(p.port), 10))
	var lastErr error

	for i := 0; i < int(p.tries); i++ {
		start := time.Now()

		conn, err := p.dialer.DialContext(ctx, "tcp", address)
		if err != nil {
			lastErr = err
			if isTimeout(err) {
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

// Close is a no-op; TCPProbe drops sockets inside Run.
func (p *TCPProbe) Close() error {
	return nil
}

func isTimeout(err error) bool {
	if ne, ok := err.(net.Error); ok {
		return ne.Timeout()
	}
	return false
}
