package dns

import (
	"context"
	"fmt"
	"net"
	"net/netip"

	vaydns "github.com/net2share/vaydns/client"

	"github.com/MohsenBg/bgscan/internal/core/netutil"
)

const vaydnsDir = "vaydns"

// VayDNSConfigFile pairs a VayDNS configuration with its on-disk identity.
type VayDNSConfigFile = ConfigFile[VayDNSConfig]

// VayDNSConfig contains the configuration required to create a VayDNS tunnel.
//
// The resolver address is provided at runtime per target via NewTunnel,
// not stored in the config.
type VayDNSConfig struct {
	Domain       string     `toml:"domain" comment:"VayDNS tunnel server domain."`
	PubKey       string     `toml:"pub_key" comment:"VayDNS server public key (64 hex chars)."`
	ClientIDSize uint16     `toml:"client_id_size" comment:"Client ID size in bytes. Range: 1-8."`
	MaxQnameLen  uint8      `toml:"max_qname_len" comment:"Maximum DNS QNAME length. 0 = auto. Range: 0-253."`
	MaxNumLabels uint8      `toml:"max_num_labels" comment:"Maximum DNS labels. 0 = auto. Range: 0-4."`
	MTU          uint16     `toml:"mtu" comment:"Tunnel MTU. 0 = auto. Range: 0-1452."`
	RPS          float64    `toml:"rps" comment:"Max DNS requests per second. 0 = unlimited. Range: 0-500."`
	RecordType   RecordType `toml:"record_type" comment:"DNS record type used by the tunnel (e.g. TXT, A, AAAA)."`

	ResolverType ResolverType `toml:"resolver_type" comment:"Resolver transport: udp, tcp, or dot (DNS over TLS)."`
	ResolverPort uint16       `toml:"resolver_port" comment:"Resolver port. 53 for DNS, 853 for DoT."`
	Fingerprint  string       `toml:"fingerprint" comment:"uTLS client fingerprint for DoT (e.g. Chrome, Firefox, random)."`

	ProxyType      ResolverProxyType `toml:"proxy_type" comment:"Proxy type in front of the resolver: socks or ssh."`
	ProxyPort      uint16            `toml:"proxy_port" comment:"Proxy port on the resolver host."`
	AuthMethod     AuthMethod        `toml:"auth_method" comment:"Proxy authentication method: none, password, or key."`
	Username       string            `toml:"username" comment:"Proxy authentication username."`
	Password       string            `toml:"password" comment:"Proxy authentication password."`
	PrivateKey     string            `toml:"private_key" comment:"SSH private key (PEM) for key authentication."`
	KnownHostsFile string            `toml:"known_hosts_file" comment:"Path to the SSH known_hosts file."`
}

// DefaultVayDNSConfig returns a VayDNS configuration with recommended defaults.
//
// Domain and PubKey must be provided for a usable configuration.
func DefaultVayDNSConfig() VayDNSConfig {
	return VayDNSConfig{
		Domain:       "",
		PubKey:       "",
		ClientIDSize: 2,
		MaxQnameLen:  101,
		MaxNumLabels: 0, // 0 means auto.
		MTU:          0, // 0 means auto
		RPS:          0, // 0 means unlimited.
		RecordType:   TypeTXT,

		ResolverType: ResolverTypeUDP,
		ResolverPort: 53,
		Fingerprint:  "Chrome",

		ProxyType:      ResolverProxySOCKS,
		ProxyPort:      1080,
		AuthMethod:     AuthNone,
		Username:       "",
		Password:       "",
		PrivateKey:     "",
		KnownHostsFile: "",
	}
}

// Validate validates the configuration and returns all validation errors
// keyed by configuration field.
func (c VayDNSConfig) Validate() map[string]error {
	errs := make(map[string]error)

	if err := netutil.ValidateDomain(c.Domain); err != nil {
		errs["domain"] = err
	}

	if err := validatePubKey(c.PubKey); err != nil {
		errs["pub_key"] = err
	}

	if c.ClientIDSize < 1 || c.ClientIDSize > 8 {
		errs["client_id_size"] = fmt.Errorf("must be between 1 and 8")
	}

	if c.MaxQnameLen > 253 {
		errs["max_qname_len"] = fmt.Errorf("must be between 0 and 253")
	}

	if c.MaxNumLabels > 4 {
		errs["max_num_labels"] = fmt.Errorf("must be between 0 and 4")
	}

	if c.MTU > 1452 {
		errs["mtu"] = fmt.Errorf("must be between 0 and 1452")
	}

	if err := validateRPS(c.RPS); err != nil {
		errs["rps"] = err
	}

	if !c.RecordType.IsValid() {
		errs["record_type"] = fmt.Errorf("invalid record type")
	}

	if !c.ResolverType.IsValid() {
		errs["resolver_type"] = fmt.Errorf("invalid resolver type")
	}

	if c.ResolverPort == 0 {
		errs["resolver_port"] = fmt.Errorf("must be greater than zero")
	}

	if err := validateFingerprint(c.Fingerprint); err != nil {
		errs["fingerprint"] = err
	}

	for field, err := range validateProxyAuth(
		c.ProxyType, c.ProxyPort, c.AuthMethod,
		c.Username, c.Password, c.PrivateKey, c.KnownHostsFile,
	) {
		errs[field] = err
	}

	return errs
}

// VayDNSService manages VayDNS configurations and tunnels.
type VayDNSService interface {
	SaveConfig(config VayDNSConfig, name string) error
	EditConfig(config VayDNSConfig, originalName string) error
	LoadConfig(name string) (VayDNSConfig, error)
	GetAllConfigFiles() ([]VayDNSConfigFile, error)
	ValidateAllConfigs() ([]ConfigValidationResult, error)
	RenameConfig(oldName, newName string) error
	NewTunnel(ctx context.Context, config VayDNSConfig, resolverAddr netip.Addr) (net.Conn, error)
}

type vayDNSService struct {
	configs configStore[VayDNSConfig]
}

// VayDNSServiceOption configures a VayDNSService.
type VayDNSServiceOption func(*vayDNSService)

// WithVayDNSDir sets the directory used to store VayDNS configurations.
func WithVayDNSDir(dir string) VayDNSServiceOption {
	return func(service *vayDNSService) {
		if dir != "" {
			service.configs.dir = dir
		}
	}
}

// NewVayDNSService creates a VayDNS service.
func NewVayDNSService(options ...VayDNSServiceOption) VayDNSService {
	service := &vayDNSService{
		configs: newConfigStore[VayDNSConfig](tunnelConfigDir(vaydnsDir), "VayDNS"),
	}

	for _, option := range options {
		option(service)
	}

	return service
}

func (s *vayDNSService) SaveConfig(config VayDNSConfig, name string) error {
	return s.configs.SaveConfig(config, name)
}

func (s *vayDNSService) EditConfig(config VayDNSConfig, originalName string) error {
	return s.configs.EditConfig(config, originalName)
}

func (s *vayDNSService) LoadConfig(name string) (VayDNSConfig, error) {
	return s.configs.LoadConfig(name)
}

func (s *vayDNSService) GetAllConfigFiles() ([]VayDNSConfigFile, error) {
	return s.configs.GetAllConfigFiles()
}

func (s *vayDNSService) ValidateAllConfigs() ([]ConfigValidationResult, error) {
	return s.configs.ValidateAllConfigs()
}

func (s *vayDNSService) RenameConfig(oldName, newName string) error {
	return s.configs.RenameConfig(oldName, newName)
}

// NewTunnel creates and initializes a VayDNS tunnel to the given resolver
// address.
func (s *vayDNSService) NewTunnel(
	ctx context.Context,
	config VayDNSConfig,
	resolverAddr netip.Addr,
) (net.Conn, error) {
	if errs := config.Validate(); len(errs) > 0 {
		return nil, fmt.Errorf("invalid VayDNS config: %v", errs)
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context canceled before tunnel setup: %w", err)
	}

	resolver, err := newVayDNSResolver(
		config.ResolverType,
		config.ResolverPort,
		config.Fingerprint,
		resolverAddr,
	)
	if err != nil {
		return nil, err
	}

	server, err := newVayDNSTunnelServer(config)
	if err != nil {
		return nil, err
	}

	tunnel, err := vaydns.NewTunnel(*resolver, *server)
	if err != nil {
		return nil, fmt.Errorf("create tunnel: %w", err)
	}

	return establishTunnelStream(ctx, tunnel, server)
}

func newVayDNSTunnelServer(config VayDNSConfig) (*vaydns.TunnelServer, error) {
	server, err := vaydns.NewTunnelServer(config.Domain, config.PubKey)
	if err != nil {
		return nil, fmt.Errorf("create tunnel server: %w", err)
	}

	server.ClientIDSize = int(config.ClientIDSize)
	server.MaxQnameLen = int(config.MaxQnameLen)
	server.MaxNumLabels = int(config.MaxNumLabels)
	server.MTU = int(config.MTU)
	server.RPS = config.RPS
	server.DnsttCompat = false

	return &server, nil
}

func toVayDNSResolverType(t ResolverType) vaydns.ResolverType {
	switch t {
	case ResolverTypeTCP:
		return vaydns.ResolverTypeTCP
	case ResolverTypeDOT:
		return vaydns.ResolverTypeDOT
	default:
		return vaydns.ResolverTypeUDP
	}
}

// Shared helpers for the protocols built on the vaydns client library
// (VayDNS itself and DNSTT compatibility mode).

// tunnelConn wraps a tunnel stream so closing it releases the whole
// tunnel stack, not just the single smux stream.
type tunnelConn struct {
	net.Conn
	tunnel *vaydns.Tunnel
}

func (c *tunnelConn) Close() error {
	if c.tunnel != nil {
		return c.tunnel.Close()
	}

	return nil
}

// newVayDNSResolver builds a vaydns resolver for the given endpoint,
// configured with the TLS fingerprint and a shared UDP socket.
func newVayDNSResolver(
	resolverType ResolverType,
	resolverPort uint16,
	fingerprint string,
	resolverAddr netip.Addr,
) (*vaydns.Resolver, error) {
	addr := netip.AddrPortFrom(resolverAddr, resolverPort).String()

	resolver, err := vaydns.NewResolver(
		toVayDNSResolverType(resolverType),
		addr,
	)
	if err != nil {
		return nil, fmt.Errorf("create resolver: %w", err)
	}

	clientHelloID, err := parseClientHelloID(fingerprint)
	if err != nil {
		return nil, fmt.Errorf("parse TLS fingerprint: %w", err)
	}

	resolver.UTLSClientHelloID = &clientHelloID
	resolver.UDPSharedSocket = true

	return &resolver, nil
}

// establishTunnelStream runs the tunnel handshake sequence and opens a
// stream. From the first step on, the tunnel owns resources: every
// failure path closes it. The returned net.Conn must be closed by the
// caller.
func establishTunnelStream(
	ctx context.Context,
	tunnel *vaydns.Tunnel,
	server *vaydns.TunnelServer,
) (net.Conn, error) {
	fail := func(msg string, err error) (net.Conn, error) {
		_ = tunnel.Close()
		return nil, fmt.Errorf("%s: %w", msg, err)
	}

	steps := []struct {
		name string
		run  func() error
	}{
		{"resolver connection", func() error { return tunnel.InitiateResolverConnection(ctx) }},
		{"DNS packet connection", func() error { return tunnel.InitiateDNSPacketConn(ctx, server.Addr) }},
		{"KCP connection", func() error { return tunnel.InitiateKCPConn(server.MTU) }},
		{"Noise channel", func() error { return tunnel.InitiateNoiseChannel(ctx) }},
		{"smux session", tunnel.InitiateSmuxSession},
	}

	for _, step := range steps {
		if err := ctx.Err(); err != nil {
			return fail("context canceled before "+step.name, err)
		}

		if err := step.run(); err != nil {
			return fail("initiate "+step.name, err)
		}
	}

	if err := ctx.Err(); err != nil {
		return fail("context canceled before opening stream", err)
	}

	stream, err := tunnel.OpenStream()
	if err != nil {
		return fail("open tunnel stream", err)
	}

	return &tunnelConn{Conn: stream, tunnel: tunnel}, nil
}
