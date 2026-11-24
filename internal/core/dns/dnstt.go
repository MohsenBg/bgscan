package dns

import (
	"context"
	"fmt"
	"net"

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
	Domain       string       `toml:"domain" comment:"DNS tunnel server domain."`
	PubKey       string       `toml:"pub_key" comment:"Server public key (64 hex chars)."`
	ResolverType ResolverType `toml:"resolver_type" comment:"Resolver transport: udp, tcp, or dot (DNS over TLS)."`
	ResolverPort uint16       `toml:"resolver_port" comment:"Resolver port. 53 for DNS, 853 for DoT."`
	Fingerprint  string       `toml:"fingerprint" comment:"uTLS client fingerprint for DoT (e.g. Chrome, Firefox, random)."`
	RPS          float64      `toml:"rps" comment:"Max DNS requests per second. 0 = unlimited. Range: 0-500."`

	ProxyType      ResolverProxyType `toml:"proxy_type" comment:"Proxy type in front of the resolver: socks or ssh."`
	ProxyPort      uint16            `toml:"proxy_port" comment:"Proxy port on the resolver host."`
	AuthMethod     AuthMethod        `toml:"auth_method" comment:"Proxy authentication method: none, password, or key."`
	Username       string            `toml:"username" comment:"Proxy authentication username."`
	Password       string            `toml:"password" comment:"Proxy authentication password."`
	PrivateKey     string            `toml:"private_key" comment:"SSH private key (PEM) for key authentication."`
	KnownHostsFile string            `toml:"known_hosts_file" comment:"Path to the SSH known_hosts file."`
}

// DefaultDNSTTConfig returns recommended defaults. Domain and PubKey are
// deployment-specific and must be provided by the user.
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

// Validate returns all validation errors, keyed by configuration field.
func (c DNSTTConfig) Validate() map[Field]error {
	errs := make(map[Field]error)

	if err := netutil.ValidateDomain(c.Domain); err != nil {
		errs[FieldDomain] = err
	}

	if err := validatePubKey(c.PubKey); err != nil {
		errs[FieldPubKey] = err
	}

	if !c.ResolverType.IsValid() {
		errs[FieldResolverType] = fmt.Errorf("resolver type is invalid")
	}

	if c.ResolverPort == 0 {
		errs[FieldResolverPort] = fmt.Errorf("resolver port must be greater than zero")
	}

	if err := validateRPS(c.RPS); err != nil {
		errs[FieldRPS] = err
	}

	if err := validateFingerprint(c.Fingerprint); err != nil {
		errs[FieldFingerprint] = err
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
	// SaveConfig validates config and writes it as a new file named
	// name. It returns an error when the name is already in use or
	// the config is invalid.
	SaveConfig(config DNSTTConfig, name string) error
	// EditConfig validates config and replaces the stored config
	// named originalName. It returns an error when that config does
	// not exist or the config is invalid.
	EditConfig(config DNSTTConfig, originalName string) error
	// LoadConfig reads and returns the stored config named name.
	LoadConfig(name string) (DNSTTConfig, error)
	// GetAllConfigFiles returns every stored config that loads and
	// validates. Files that fail either check are skipped.
	GetAllConfigFiles() ([]DNSTTConfigFile, error)
	// ValidateAllConfigs returns one result for each stored file
	// that fails to load or validate. Valid files are not included.
	ValidateAllConfigs() ([]ConfigValidationResult, error)
	// RenameConfig renames the stored config oldName to newName.
	// It returns an error when oldName does not exist or newName is
	// already in use.
	RenameConfig(oldName, newName string) error
	// NewTunnel opens a tunnel stream to the resolver at
	// resolverAddr. The caller closes the returned connection when
	// done.
	NewTunnel(ctx context.Context, config DNSTTConfig, resolverAddr string) (net.Conn, error)
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

// NewDNSTTService returns a DNSTTService that stores configs in
// the default DNSTT directory unless an option changes it.
func NewDNSTTService(options ...DNSTTServiceOption) DNSTTService {
	service := &dnsttService{
		configs: newConfigStore[DNSTTConfig](tunnelConfigDir(dnsttDir), "DNSTT"),
	}

	for _, option := range options {
		option(service)
	}

	return service
}

// SaveConfig saves the config through the shared config store.
func (s *dnsttService) SaveConfig(config DNSTTConfig, name string) error {
	return s.configs.SaveConfig(config, name)
}

// EditConfig replaces the stored config through the shared config store.
func (s *dnsttService) EditConfig(config DNSTTConfig, originalName string) error {
	return s.configs.EditConfig(config, originalName)
}

// LoadConfig loads the config through the shared config store.
func (s *dnsttService) LoadConfig(name string) (DNSTTConfig, error) {
	return s.configs.LoadConfig(name)
}

// GetAllConfigFiles lists stored configs through the shared config store.
func (s *dnsttService) GetAllConfigFiles() ([]DNSTTConfigFile, error) {
	return s.configs.GetAllConfigFiles()
}

// ValidateAllConfigs validates stored configs through the shared config store.
func (s *dnsttService) ValidateAllConfigs() ([]ConfigValidationResult, error) {
	return s.configs.ValidateAllConfigs()
}

// RenameConfig renames the config through the shared config store.
func (s *dnsttService) RenameConfig(oldName, newName string) error {
	return s.configs.RenameConfig(oldName, newName)
}

// NewTunnel creates and initializes a DNSTT tunnel to the given resolver
// address, using the vaydns client in DNSTT compatibility mode.
func (s *dnsttService) NewTunnel(
	ctx context.Context,
	config DNSTTConfig,
	resolverAddr string,
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
