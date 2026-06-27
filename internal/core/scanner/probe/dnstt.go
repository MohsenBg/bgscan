package probe

import (
	"context"
	"fmt"
	"time"

	"bgscan/internal/core/dns"
	"bgscan/internal/core/result"
	"bgscan/internal/core/scanner/portmgr"
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
	// and end‑to‑end tunnel checks performed by the probe. If set to a
	// non‑positive value, a sensible default is applied.
	Timeout time.Duration
}

// DNSTTProbe implements the Probe interface using a DNSTT tunnel as the
// underlying measurement primitive. For each Run call, a DNSTT tunnel is
// established to the target resolver, a local SOCKS5 proxy is exposed,
// and connectivity through the tunnel is verified.
type DNSTTProbe struct {
	pm              *portmgr.PortManager
	processRegistry *ProcessRegistry
	config          DNSTTConfig
}

// NewDNSTTProbe constructs a new DNSTTProbe using the provided configuration
// and PortManager. The returned value implements the Probe interface.
//
// The function validates basic configuration invariants:
//
//   - pm must not be nil
//   - config.Domain must be non‑empty
//   - config.Timeout <= 0 is replaced with a default value
//
// NewDNSTTProbe does not start any background goroutines. Call Init on the
// returned Probe to initialize internal state and start the process registry.
func NewDNSTTProbe(config DNSTTConfig, pm *portmgr.PortManager) (Probe, error) {
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
		processRegistry: NewProcessRegistry(),
		config:          config,
	}, nil
}

// Init starts the internal ProcessRegistry used to track and manage DNSTT
// tunnel processes. Init is intended to be called once during scanner
// startup, prior to any Run calls.
//
// The provided context governs the lifecycle of the registry's monitoring
// goroutine: when ctx is canceled, all tracked processes are terminated
// and the registry shuts down.
//
// Init is idempotent with respect to the ProcessRegistry's Start method,
// but should conventionally be called exactly once per DNSTTProbe instance.
func (d *DNSTTProbe) Init(ctx context.Context) error {
	d.processRegistry.Start(ctx)
	return nil
}

// Run executes a single DNSTT measurement against the given IP address.
//
// The execution flow is:
//
//  1. Allocate a local ephemeral port via PortManager.
//  2. Construct a DNSTT client using the probe configuration.
//  3. Start a DNSTT tunnel to the target resolver IP on the allocated port.
//  4. Register the tunnel process in the ProcessRegistry for coordinated cleanup.
//  5. Wait for the local SOCKS5 proxy to become reachable.
//  6. Perform a connectivity check through the proxy using dns.TestProxy.
//  7. Measure and return the end‑to‑end latency as an IPScanResult.
//
// Run is fully context‑aware:
//
//   - If ctx is canceled at any point, the method aborts early and returns ctx.Err().
//   - Port allocation, tunnel startup, port‑open wait, and proxy validation all
//     honor the provided context.
//
// On success, Run returns a *result.IPScanResult populated with the target IP
// and measured latency. On failure, a non‑nil error is returned and all
// allocated resources (ports, processes, tunnels) are cleaned up best‑effort.
func (d *DNSTTProbe) Run(ctx context.Context, ip string) (*result.IPScanResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Allocate local port.
	localPort, err := d.pm.GetPort(ctx)
	if err != nil {
		return nil, err
	}
	defer d.pm.ReleasePort(localPort)

	// Construct DNSTT client.
	client, err := dns.NewDNSTTClient(
		d.config.Domain,
		d.config.PubKey,
		d.config.Transport,
		d.config.DNSPort,
	)
	if err != nil {
		return nil, err
	}

	// Start DNSTT tunnel.
	proc, err := client.RunTunnel(ctx, ip, localPort)
	if err != nil {
		return nil, err
	}

	// Track process for coordinated shutdown via ProcessRegistry.
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

	return &result.IPScanResult{
		IP:      ip,
		Latency: time.Since(start),
	}, nil
}

// Close implements the Probe interface's Close method for DNSTTProbe.
// Currently DNSTTProbe does not maintain any long‑lived resources that
// require explicit teardown beyond the per‑Run cleanup, so Close is a
// no‑op and returns nil.
//
// The method is provided for future‑proofing and to conform to the Probe
// lifecycle contract (Init → Run → Close).
func (d *DNSTTProbe) Close() error {
	return nil
}
