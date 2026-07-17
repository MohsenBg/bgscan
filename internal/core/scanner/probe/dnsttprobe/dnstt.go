package dnsttprobe

import (
	"context"
	"fmt"
	"time"

	"bgscan/internal/core/dns"
	"bgscan/internal/core/result"
	"bgscan/internal/core/scanner/portmgr"
	"bgscan/internal/core/scanner/probe"
	"bgscan/internal/logger"
)

// DNSTTConfig describes the static configuration required to establish
// a DNSTT tunnel from a scan worker. All fields are intended to be
// configured once per worker and reused across multiple probe runs.
type DNSTTConfig struct {
	// Domain is the DNSTT front domain used for encapsulating DNS traffic.
	// This is typically a domain controlled by the operator or provided
	// by the DNSTT service configuration.
	Domain string

	// PubKey is the server's public key used by the DNSTT client for
	// establishing a secure tunnel. The exact format is defined by the
	// underlying dns.DNSTTClient implementation.
	PubKey string

	// Transport selects the DNS transport mechanism (e.g. UDP, DoH, DoT)
	// used by the tunnel. The concrete options are defined in the
	// bgscan/internal/core/dns package.
	Transport dns.Transport

	// DNSPort is the remote DNS port on the target resolver. For standard
	// DNS this is usually 53, but DNSTT deployments may use alternative
	// ports depending on the environment.
	DNSPort uint16

	// Timeout defines the maximum duration allowed for proxy validation
	// and end-to-end tunnel checks performed by the probe. If set to a
	// non-positive value, a sensible default is applied.
	Timeout time.Duration
}

// DNSTTProbe implements the Probe interface using a DNSTT tunnel as the
// underlying measurement primitive. For each Run call, a DNSTT tunnel is
// established to the target resolver, a local SOCKS5 proxy is exposed,
// and connectivity through the tunnel is verified.
type DNSTTProbe struct {
	pm              *portmgr.PortManager
	processRegistry *probe.ProcessRegistry
	config          DNSTTConfig
}

// NewDNSTTProbe returns a new DNSTTProbe configured with the given
// DNSTTConfig and PortManager. If config.Timeout is non-positive, a
// default of 5 seconds is applied. It does not start any background
// goroutines; call Init to start the process registry.
func NewDNSTTProbe(config DNSTTConfig, pm *portmgr.PortManager) (probe.Probe, error) {
	if pm == nil {
		return nil, fmt.Errorf("port manager cannot be nil")
	}

	if config.Domain == "" {
		return nil, fmt.Errorf("dns domain cannot be empty")
	}

	if config.Timeout <= 0 {
		config.Timeout = 5 * time.Second
	}

	return &DNSTTProbe{
		pm:              pm,
		processRegistry: probe.NewProcessRegistry(),
		config:          config,
	}, nil
}

// Schema returns the result schema for DNSTT probes.
func (d *DNSTTProbe) Schema() result.ResultSchema {
	return Schema
}

// Init starts the internal process registry. The provided context governs
// the lifecycle of the registry's monitoring goroutine; canceling it
// terminates all tracked processes.
func (d *DNSTTProbe) Init(ctx context.Context) error {
	d.processRegistry.Start(ctx)
	return nil
}

// Run executes a single DNSTT measurement against the given IP address.
// It allocates a local port, establishes a DNSTT tunnel, and verifies
// connectivity through the resulting local SOCKS5 proxy. All operations
// honor the provided context, and resources are cleaned up on failure
// or early cancellation.
func (d *DNSTTProbe) Run(ctx context.Context, ip string) (result.Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	localPort, err := d.pm.GetPort(ctx)
	if err != nil {
		return nil, err
	}
	defer d.pm.ReleasePort(localPort)

	client, err := dns.NewDNSTTClient(
		d.config.Domain,
		d.config.PubKey,
		d.config.Transport,
		d.config.DNSPort,
	)
	if err != nil {
		return nil, err
	}

	proc, err := client.RunTunnel(ctx, ip, localPort)
	if err != nil {
		return nil, err
	}

	// Register process for coordinated shutdown.
	id, err := d.processRegistry.Register(ctx, proc)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := d.processRegistry.Unregister(ctx, id); err != nil {
			logger.CoreError("failed to unregister process %s: %s", id, err)
		}
	}()

	// Ensure the tunnel is stopped when Run returns.
	defer func() {
		_ = client.StopTunnel(context.Background())
	}()

	localProxyAddr := fmt.Sprintf("127.0.0.1:%d", localPort)

	// Wait for local SOCKS5 proxy to accept connections.
	if err := portmgr.WaitPortOpen(ctx, localProxyAddr, time.Second); err != nil {
		return nil, err
	}

	start := time.Now()

	// Validate connectivity through the tunnelled proxy.
	if ok := dns.TestProxy(ctx, localProxyAddr, d.config.Timeout); !ok {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, fmt.Errorf("tunnel connectivity failed: %s", ip)
	}

	return DNSTTResult{
		IP:        ip,
		Latency:   time.Since(start),
		Transport: d.config.Transport,
		Port:      localPort,
	}, nil
}

// Close is a no-op as DNSTTProbe does not maintain long-lived resources
// beyond per-Run cleanup.
func (d *DNSTTProbe) Close() error {
	return nil
}
