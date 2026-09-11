package dns

import (
	"context"
	"fmt"
	"net"
	"net/netip"

	vaydns "github.com/net2share/vaydns/client"

	"github.com/MohsenBg/bgscan/internal/core/netutil"
)

const dnsttDir = "dnstt"

// DNSTTConfigFile pairs a DNSTT configuration with its on-disk identity.
type DNSTTConfigFile = ConfigFile[DNSTTConfig]

// DNSTTConfig contains the configuration required by the DNSTT client.
//
// The resolver address is provided at runtime per target via NewTunnel,
// not stored in the config.
type DNSTTConfig struct {
	Domain       string
	PubKey       string
	ResolverType ResolverType
	ResolverPort uint16
	Fingerprint  string
	RPS          float64

	ProxyType      ResolverProxyType
	ProxyPort      uint16
	AuthMethod     AuthMethod
	Username       string
	Password       string
	PrivateKey     string
	KnownHostsFile string
}

// DefaultDNSTTConfig returns a DNSTT configuration with recommended defaults.
//
// Domain and PubKey are deployment-specific and must be provided by the user.
func DefaultDNSTTConfig() DNSTTConfig {
	return DNSTTConfig{
		Domain:         "",
		ProxyPort:      1080,
		ResolverType:   ResolverTypeUDP,
		ResolverPort:   53,
		Fingerprint:    "Chrome",
		RPS:            0, // 0 = unlimited.
		AuthMethod:     AuthNone,
		ProxyType:      ResolverProxySOCKS,
		PubKey:         "",
		Username:       "",
		Password:       "",
		PrivateKey:     "",
		KnownHostsFile: "",
	}
}

// Validate validates the DNSTT configuration.
//
// All validation errors are returned, keyed by the corresponding
// configuration field.
func (c DNSTTConfig) Validate() map[string]error {
	errs := make(map[string]error)

	if err := netutil.ValidateDomain(c.Domain); err != nil {
		errs["domain"] = err
	}

	if err := validatePubKey(c.PubKey); err != nil {
		errs["pub_key"] = err
	}

	if !c.ResolverType.IsValid() {
		errs["resolver_type"] = fmt.Errorf("resolver type is invalid")
	}

	if c.ResolverPort == 0 {
		errs["resolver_port"] = fmt.Errorf("resolver port must be greater than zero")
	}

	if err := validateRPS(c.RPS); err != nil {
		errs["rps"] = err
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

// DNSTTService manages DNSTT configurations and tunnels.
type DNSTTService interface {
	SaveConfig(config DNSTTConfig, name string) error
	EditConfig(config DNSTTConfig, originalName string) error
	LoadConfig(name string) (DNSTTConfig, error)
	GetAllConfigFiles() ([]DNSTTConfigFile, error)
	ValidateAllConfigs() ([]ConfigValidationResult, error)
	RenameConfig(oldName, newName string) error
	NewTunnel(ctx context.Context, config DNSTTConfig, resolverAddr netip.Addr) (net.Conn, error)
}

// dnsttService is the default DNSTTService implementation.
type dnsttService struct {
	configs configStore[DNSTTConfig]
}

// DNSTTServiceOption configures a DNSTTService.
type DNSTTServiceOption func(*dnsttService)

// WithDNSTTDir sets the directory used to store DNSTT configurations.
func WithDNSTTDir(dir string) DNSTTServiceOption {
	return func(service *dnsttService) {
		if dir != "" {
			service.configs.dir = dir
		}
	}
}

// NewDNSTTService creates a DNSTT service.
func NewDNSTTService(options ...DNSTTServiceOption) DNSTTService {
	service := &dnsttService{
		configs: newConfigStore[DNSTTConfig](tunnelConfigDir(dnsttDir), "DNSTT"),
	}

	for _, option := range options {
		option(service)
	}

	return service
}

func (s *dnsttService) SaveConfig(config DNSTTConfig, name string) error {
	return s.configs.SaveConfig(config, name)
}

func (s *dnsttService) EditConfig(config DNSTTConfig, originalName string) error {
	return s.configs.EditConfig(config, originalName)
}

func (s *dnsttService) LoadConfig(name string) (DNSTTConfig, error) {
	return s.configs.LoadConfig(name)
}

func (s *dnsttService) GetAllConfigFiles() ([]DNSTTConfigFile, error) {
	return s.configs.GetAllConfigFiles()
}

func (s *dnsttService) ValidateAllConfigs() ([]ConfigValidationResult, error) {
	return s.configs.ValidateAllConfigs()
}

func (s *dnsttService) RenameConfig(oldName, newName string) error {
	return s.configs.RenameConfig(oldName, newName)
}

// NewTunnel creates and initializes a DNSTT tunnel to the given resolver
// address, using the vaydns client in DNSTT compatibility mode.
func (s *dnsttService) NewTunnel(
	ctx context.Context,
	config DNSTTConfig,
	resolverAddr netip.Addr,
) (net.Conn, error) {
	if errs := config.Validate(); len(errs) > 0 {
		return nil, fmt.Errorf("invalid DNSTT config: %v", errs)
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

	server, err := newDNSTTTunnelServer(config)
	if err != nil {
		return nil, err
	}

	tunnel, err := vaydns.NewTunnel(*resolver, *server)
	if err != nil {
		return nil, fmt.Errorf("create tunnel: %w", err)
	}

	return establishTunnelStream(ctx, tunnel, server)
}

// newDNSTTTunnelServer creates a DNSTT-compatible tunnel server.
func newDNSTTTunnelServer(config DNSTTConfig) (*vaydns.TunnelServer, error) {
	server, err := vaydns.NewTunnelServer(config.Domain, config.PubKey)
	if err != nil {
		return nil, fmt.Errorf("create tunnel server: %w", err)
	}

	server.RPS = config.RPS
	server.DnsttCompat = true

	return &server, nil
}
