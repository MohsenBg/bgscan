package portmgr

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"os"
	"sync"
	"time"

	"bgscan/internal/logger"
)

// PortManager manages a pool of reusable TCP ports.
//
// Ports are handed out to callers on demand and returned to the pool when
// released. The manager verifies that a port is actually free before
// returning it, preventing conflicts with other processes.
//
// The type is safe for concurrent use.
type PortManager struct {
	ports     chan uint16
	done      chan struct{}
	closeOnce sync.Once
}

// GenerateInstancePortRange returns a port range derived from the current
// process ID and a random offset.
//
// This helps reduce collisions when multiple instances of the program run
// on the same machine by shifting the starting port of the pool.
func GenerateInstancePortRange(base, poolSize uint16) (uint16, uint16) {
	pid := os.Getpid()
	randomPart := uint16(rand.Intn(1500))
	start := base + uint16(pid%20000) + randomPart
	return start, poolSize
}

// NewPortManager creates a new PortManager with a sequential port pool
// starting at startPort and containing count ports.
func NewPortManager(startPort, count uint16) (*PortManager, error) {
	if count == 0 {
		return nil, errors.New("port count must be greater than zero")
	}

	pm := &PortManager{
		ports: make(chan uint16, count),
		done:  make(chan struct{}),
	}

	for i := range count {
		pm.ports <- startPort + i
	}

	return pm, nil
}

// GetPort retrieves an available port from the pool.
//
// The function blocks until a port becomes available, the context is
// canceled, or the manager is closed. Before returning a port, the manager
// verifies that it can be bound locally.
func (pm *PortManager) GetPort(ctx context.Context) (uint16, error) {
	for {
		select {
		case <-pm.done:
			return 0, errors.New("port manager closed")

		case <-ctx.Done():
			return 0, ctx.Err()

		case port, ok := <-pm.ports:
			if !ok {
				return 0, errors.New("port manager closed")
			}

			if isPortFree(port) {
				return port, nil
			}
		}
	}
}

// ReleasePort returns a previously acquired port back to the pool.
//
// If the pool is already full or the manager is closed, the port is silently
// discarded.
func (pm *PortManager) ReleasePort(port uint16) {
	select {
	case <-pm.done:
		return
	case pm.ports <- port:
	default:
	}
}

// Close shuts down the PortManager and releases all resources.
//
// After calling Close, all blocked GetPort calls will return an error and
// future operations become no‑ops.
func (pm *PortManager) Close() {
	pm.closeOnce.Do(func() {
		close(pm.done)
		close(pm.ports)
	})
}

// isPortFree checks whether a TCP port can be bound locally.
func isPortFree(port uint16) bool {
	addr := fmt.Sprintf("127.0.0.1:%d", port)

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return false
	}

	if err := ln.Close(); err != nil {
		logger.CoreError("error closing listener: %v", err)
	}

	return true
}

// WaitPortOpen waits until a TCP service becomes reachable at addr.
//
// The function repeatedly attempts to connect until either the port opens
// or the timeout expires.
func WaitPortOpen(ctx context.Context, addr string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	dialer := net.Dialer{Timeout: 300 * time.Millisecond}

	for {
		select {
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				return fmt.Errorf("timeout waiting for port %s", addr)
			}
			return ctx.Err()

		case <-ticker.C:
			conn, err := dialer.DialContext(ctx, "tcp", addr)
			if err == nil {
				if err := conn.Close(); err != nil {
					logger.CoreError("error closing connection: %v", err)
				}
				return nil
			}
		}
	}
}

// RandomBasePort returns a randomized base port suitable for allocating
// a port pool of the given size.
//
// The generated range avoids the system ephemeral port range.
func RandomBasePort(poolSize uint16) uint16 {
	const min uint16 = 20000
	const ephemeralStart uint16 = 49152

	maxBase := ephemeralStart - poolSize
	return min + uint16(rand.Intn(int(maxBase-min)))
}
