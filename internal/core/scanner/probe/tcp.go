package probe

import (
	"bgscan/internal/core/result"
	"context"
	"fmt"
	"net"
	"time"
)

// TCPProbe performs a lightweight TCP connectivity probe against a target IP address.
//
// It is designed to verify Layer-4 transport reachability and measure handshake latency
// (RTT) without relying on or making assumptions about application-layer protocols
// (such as HTTP, SSH, or TLS).
//
// The probe executes up to a specified number of attempts (tries) and returns
// immediately upon the first successful connection handshake.
//
// Internally, it:
//   - Establishes a raw TCP connection using net.Dialer
//   - Measures the round-trip time (RTT) taken to complete the 3-way handshake
//   - Cleanly closes the connection immediately to release network resources
//
// This probe is optimized for high-throughput network scanning and does not attempt
// application protocol negotiation or banner parsing.
type TCPProbe struct {
	port    string
	timeout time.Duration
	dialer  net.Dialer
	tries   uint16
}

// NewTCPProbe creates, configures, and returns a new TCPProbe instance.
//
// Parameters:
//   - port: The target destination port string (e.g., "80", "443").
//   - timeout: The max duration to wait for an individual connection attempt to establish.
//   - tries: The maximum number of retry attempts before declaring the target unreachable.
//
// Example:
//
//	p := NewTCPProbe("443", 3*time.Second, 3)
func NewTCPProbe(port string, timeout time.Duration, tries uint16) Probe {
	return &TCPProbe{
		port:    port,
		tries:   tries,
		timeout: timeout,
		dialer: net.Dialer{
			Timeout: timeout,
		},
	}
}

// Init initializes the probe instance.
//
// TCPProbe does not require persistent state initialization and always returns nil.
// This method exists to satisfy the global Probe interface lifecycle.
func (p *TCPProbe) Init(ctx context.Context) error {
	return nil
}

// Run executes the TCP probe loop against the given target IP address.
//
// It attempts to establish a raw TCP connection to the destination ip:port.
// If a connection succeeds, it captures the network latency and stops.
//
// If an attempt fails due to a network timeout, it logs the failure and retries until
// the maximum try threshold is exhausted. Non-timeout errors (e.g., Connection Refused)
// fail fast and abort immediately to speed up scan routines.
//
// Returns an IPScanResult struct on success, or an error detailing the root cause of failure.
func (p *TCPProbe) Run(ctx context.Context, ip string) (*result.IPScanResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	address := net.JoinHostPort(ip, p.port)
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

		conn.Close()

		return &result.IPScanResult{
			IP:      ip,
			Latency: rtt,
		}, nil
	}

	return nil, fmt.Errorf("tcp probe failed after %d tries: %w", p.tries, lastErr)
}

// Close releases any persistent infrastructure or network resources held by the probe.
//
// Since TCPProbe drops sockets instantly inside Run, it holds no persistent state,
func (p *TCPProbe) Close() error {
	return nil
}

