package dns

import (
	"context"
	"fmt"
	"net"
	"time"

	"bgscan/internal/core/process"
	"bgscan/internal/core/scanner/portmgr"
)

// DNSTTClientPaths returns candidate filesystem paths where the dnstt-client
// binary may be located. These paths are searched in order.
func DNSTTClientPaths() []string {
	return []string{
		"assets/dnstt-client",
		"assets/dns/dnstt-client",
		"dnstt-client",
	}
}

// DNSTTClient wraps execution and lifecycle management of the dnstt-client
// binary used to establish DNS tunnels.
type DNSTTClient struct {
	bin       string
	transport Transport
	publicKey string
	domain    string
	process   *process.Process
	dnsPort   uint16
}

// NewDNSTTClient creates a new DNSTTClient instance. It automatically locates
// the dnstt-client binary and initializes client settings.
func NewDNSTTClient(domain, pubKey string, transport Transport, dnsPort uint16) (*DNSTTClient, error) {
	path, err := FindDNSTTClient()
	if err != nil {
		return nil, err
	}

	return &DNSTTClient{
		bin:       path,
		domain:    domain,
		transport: transport,
		publicKey: pubKey,
		dnsPort:   dnsPort,
	}, nil
}

// FindDNSTTClient returns the path to the dnstt-client binary by scanning
// a set of known asset directories and falling back to PATH.
func FindDNSTTClient() (string, error) {
	return process.FindBinaryInPaths("dnstt-client", DNSTTClientPaths())
}

// RunTunnel starts the dnstt-client process and attempts to establish
// a DNS tunnel to the specified ip:port endpoint.
//
// The process handle is stored internally and returned for external use.
func (d *DNSTTClient) RunTunnel(ctx context.Context, ip string, port uint16) (*process.Process, error) {
	args := []string{
		getDNSTransportFlag(d.transport),
		net.JoinHostPort(ip, fmt.Sprint(d.dnsPort)),
		"-pubkey", d.publicKey,
		d.domain,
		fmt.Sprintf("127.0.0.1:%d", port),
	}

	proc, err := process.Start(ctx, d.bin, args...)
	if err != nil {
		fmt.Printf("%s", err.Error())
		return proc, err
	}

	d.process = proc
	return proc, nil
}

// StopTunnel gracefully terminates the running dnstt-client process.
func (d *DNSTTClient) StopTunnel(ctx context.Context) error {
	if d.process == nil {
		return fmt.Errorf("dnstt-client process not running")
	}
	return d.process.StopGracefully(2 * time.Second)
}

// getDNSTransportFlag converts a Transport value into the correct CLI flag
// for dnstt-client. For unsupported transports, a safe fallback is used.
func getDNSTransportFlag(transport Transport) string {
	switch transport {
	case UDP:
		return "-udp"
	case TCP, DOT:
		return "-dot"
	case DOH:
		// DOH not implemented; fallback to DOT.
		return "-dot"
	default:
		return "-udp"
	}
}

// VerifyDNSTTClient performs a functional health check of the dnstt-client
// binary by launching a temporary tunnel and validating readiness.
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
		return fmt.Errorf("failed to start tunnel: %w", err)
	}
	defer func() { _ = proc.Kill() }()

	addr := net.JoinHostPort("127.0.0.1", fmt.Sprint(testPort))

	if err := portmgr.WaitPortOpen(ctx, addr, 3*time.Second); err != nil {
		return fmt.Errorf("tunnel did not become ready: %w", err)
	}

	return nil
}
