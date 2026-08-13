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
)

var ErrClosed = errors.New("port manager closed")

type Manager interface {
	Get(context.Context) (uint16, error)
	Release(uint16)
	Close()
	WaitOpen(ctx context.Context, addr string, timeout time.Duration) error
}

// manager provides reusable TCP ports from a managed pool.
type manager struct {
	ports chan uint16
	done  chan struct{}

	closeOnce sync.Once

	isFree func(uint16) bool
}

// New creates a port manager with ports starting from startPort.
func New(startPort, count uint16) (Manager, error) {
	if count == 0 {
		return nil, errors.New("port count must be greater than zero")
	}

	m := &manager{
		ports:  make(chan uint16, count),
		done:   make(chan struct{}),
		isFree: portAvailable,
	}

	for i := range count {
		m.ports <- startPort + i
	}

	return m, nil
}

// Get returns an available port.
// It waits until a port is free, the context is canceled,
// or the manager is closed.
func (m *manager) Get(ctx context.Context) (uint16, error) {
	for {
		select {
		case <-m.done:
			return 0, ErrClosed
		default:
		}

		select {
		case <-m.done:
			return 0, ErrClosed

		case <-ctx.Done():
			return 0, ctx.Err()

		case port := <-m.ports:
			select {
			case <-m.done:
				return 0, ErrClosed
			default:
			}

			if m.isFree(port) {
				return port, nil
			}
		}
	}
}

// Release puts a port back into the pool.
func (m *manager) Release(port uint16) {
	select {
	case <-m.done:
		return
	default:
	}

	select {
	case <-m.done:
		return

	case m.ports <- port:
	default:
		// Pool is full.
	}
}

// Close releases all resources.
func (m *manager) Close() {
	m.closeOnce.Do(func() {
		close(m.done)
	})
}

func portAvailable(port uint16) bool {
	addr := fmt.Sprintf("127.0.0.1:%d", port)

	conn, err := net.Listen("tcp", addr)
	if err != nil {
		return false
	}

	return conn.Close() == nil
}

// GenerateInstanceRange creates a different port range for each process.
func GenerateInstanceRange(base, size uint16) (uint16, uint16) {
	offset := uint16(os.Getpid()%20000) + uint16(rand.Intn(1500))

	return base + offset, size
}

// RandomBase returns a random starting port avoiding ephemeral ports.
func RandomBase(size uint16) uint16 {
	const (
		minPort       = 20000
		ephemeralPort = 49152
	)

	max := ephemeralPort - size

	return minPort + uint16(rand.Intn(int(max-minPort)))
}

// WaitOpen waits until a TCP service accepts connections.
func (s *manager) WaitOpen(ctx context.Context, addr string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	dialer := net.Dialer{
		Timeout: 300 * time.Millisecond,
	}

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("waiting for %s: %w", addr, ctx.Err())

		case <-ticker.C:
			conn, err := dialer.DialContext(ctx, "tcp", addr)
			if err != nil {
				continue
			}

			defer func() {
				_ = conn.Close()
			}()
			return nil
		}
	}
}
