package dns

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"bgscan/internal/core/process"
	"bgscan/internal/core/scanner/portmgr"
)

var ErrTunnelNotRunning = errors.New("dnstt-client process is not running")

// DNSTTClient manages a dnstt-client tunnel process.
type DNSTTClient interface {
	RunTunnel(context.Context, string, uint16) (process.Process, error)
}

// ProcessStarter starts an external process.
type ProcessStarter func(context.Context, string, ...string) (process.Process, error)

type dnsttClient struct {
	bin       string
	transport Transport
	publicKey string
	domain    string
	dnsPort   uint16

	start ProcessStarter
}

// DNSTTClientOption configures a dnstt-client instance.
type DNSTTClientOption func(*dnsttClient)

// WithProcessStarter replaces the process launcher.
//
// It allows tests to verify RunTunnel without executing an external binary.
func WithProcessStarter(start ProcessStarter) DNSTTClientOption {
	return func(client *dnsttClient) {
		if start != nil {
			client.start = start
		}
	}
}

// WithDNSTTClientBinary uses bin as the dnstt-client executable path instead
// of searching the known locations.
func WithDNSTTClientBinary(bin string) DNSTTClientOption {
	return func(client *dnsttClient) {
		if bin != "" {
			client.bin = bin
		}
	}
}

// NewDNSTTClient creates a client for a DNS tunnel.
//
// Unless WithDNSTTClientBinary is provided, it locates dnstt-client before
// returning the client.
func NewDNSTTClient(
	domain,
	publicKey string,
	transport Transport,
	dnsPort uint16,
	opts ...DNSTTClientOption,
) (DNSTTClient, error) {
	client := &dnsttClient{
		domain:    domain,
		publicKey: publicKey,
		transport: transport,
		dnsPort:   dnsPort,
		start:     process.Start,
	}

	for _, opt := range opts {
		opt(client)
	}

	if client.bin == "" {
		bin, err := FindDNSTTClient()
		if err != nil {
			return nil, err
		}

		client.bin = bin
	}

	return client, nil
}

// FindDNSTTClient locates the dnstt-client binary in known locations or PATH.
func FindDNSTTClient() (string, error) {
	return process.FindBinaryInPaths("dnstt-client", getDNSTTPaths())
}

// RunTunnel starts dnstt-client and connects the local port to ip:port through
// the configured DNS transport.
func (d *dnsttClient) RunTunnel(
	ctx context.Context,
	ip string,
	port uint16,
) (process.Process, error) {
	args := []string{
		getDNSTransportFlag(d.transport),
		net.JoinHostPort(ip, fmt.Sprint(d.dnsPort)),
		"-pubkey", d.publicKey,
		d.domain,
		net.JoinHostPort("127.0.0.1", fmt.Sprint(port)),
	}

	proc, err := d.start(ctx, d.bin, args...)
	if err != nil {
		return nil, err
	}

	return proc, nil
}

// getDNSTransportFlag maps a transport to its corresponding dnstt-client flag.
func getDNSTransportFlag(transport Transport) string {
	switch transport {
	case UDP:
		return "-udp"

	case TCP, DOT:
		return "-dot"

	case DOH:
		// dnstt-client does not currently support DoH directly.
		return "-dot"

	default:
		return "-udp"
	}
}

// VerifyDNSTTClient starts a temporary tunnel and waits for its local listener
// to accept TCP connections.
func VerifyDNSTTClient() error {
	client, err := NewDNSTTClient(
		"example.com",
		"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		UDP,
		53,
	)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	const testPort uint16 = 9999

	proc, err := client.RunTunnel(ctx, "8.8.8.8", testPort)
	if err != nil {
		return fmt.Errorf("start tunnel: %w", err)
	}
	defer func() {
		_ = proc.Kill()
	}()

	addr := net.JoinHostPort("127.0.0.1", fmt.Sprint(testPort))

	if err := portmgr.WaitOpen(ctx, addr, 3*time.Second); err != nil {
		return fmt.Errorf("wait for tunnel: %w", err)
	}

	return nil
}

func getDNSTTPaths() []string {
	exe, err := os.Executable()
	if err != nil {
		return []string{"assets/dnstt-client", "assets/dns/dnstt-client", "dnstt-client", ""}
	}

	base := filepath.Dir(exe)

	return []string{
		filepath.Join(base, "assets", "dnstt-client"),
		filepath.Join(base, "assets", "dns", "dnstt-client"),
		filepath.Join(base, "dnstt-client"),
		base,
	}
}
