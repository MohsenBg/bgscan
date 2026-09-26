package dns

import (
	"context"
	"fmt"
	"net/netip"

	"github.com/MohsenBg/bgscan/internal/core/ffi/slipstream"
	"github.com/MohsenBg/bgscan/internal/core/netutil"
)

const slipstreamDir = "slipstream"

// SlipstreamConfigFile pairs a Slipstream configuration with its on-disk
// identity.
type SlipstreamConfigFile = ConfigFile[SlipstreamConfig]

type CongestionControl string

const (
	CongestionControlBBR   CongestionControl = "bbr"
	CongestionControlCubic CongestionControl = "cubic"
)

// IsValid reports whether the congestion control value is known.
// Empty means "use the library default" and is valid for backward
// compatibility with configs written before the field existed.
func (c CongestionControl) IsValid() bool {
	switch c {
	case "", CongestionControlBBR, CongestionControlCubic:
		return true
	default:
		return false
	}
}

type DNSResolution string

const (
	DNSResolutionRecursive     DNSResolution = "recursive"
	DNSResolutionAuthoritative DNSResolution = "authoritative"
)

// IsValid reports whether the DNS resolution mode is known.
// Empty is accepted for backward compatibility and means recursive.
func (r DNSResolution) IsValid() bool {
	switch r {
	case "", DNSResolutionRecursive, DNSResolutionAuthoritative:
		return true
	default:
		return false
	}
}

func (r DNSResolution) libValue() slipstream.DNSResolution {
	switch r {
	case DNSResolutionAuthoritative:
		return slipstream.ModeAuthoritative
	case DNSResolutionRecursive:
		fallthrough
	default:
		return slipstream.ModeRecursive
	}
}

// SlipstreamConfig contains the configuration required by the
// Slipstream DNS client.
type SlipstreamConfig struct {
	Domain       string `toml:"domain" comment:"Slipstream server domain."`
	ResolverPort uint16 `toml:"resolver_port" comment:"Resolver port. Standard DNS uses 53."`
	CertPath     string `toml:"cert_path" comment:"Path to the CA certificate used to verify the server. Empty = system pool."`

	DNSResolution     DNSResolution     `toml:"dns_resolution" comment:"DNS resolution mode: recursive or authoritative."`
	GSO               bool              `toml:"gso" comment:"Enable UDP generic segmentation offload."`
	CongestionControl CongestionControl `toml:"congestion_control" comment:"Congestion control: bbr or cubic. Empty = library default."`
	KeepAliveInterval uint16            `toml:"-"`

	ProxyType      ResolverProxyType `toml:"proxy_type" comment:"Proxy type in front of the resolver: socks or ssh."`
	ProxyPort      uint16            `toml:"proxy_port" comment:"Proxy port on the resolver host."`
	AuthMethod     AuthMethod        `toml:"auth_method" comment:"Proxy authentication method: none, password, or key."`
	Username       string            `toml:"username" comment:"Proxy authentication username."`
	Password       string            `toml:"password" comment:"Proxy authentication password."`
	PrivateKey     string            `toml:"private_key" comment:"SSH private key (PEM) for key authentication."`
	KnownHostsFile string            `toml:"known_hosts_file" comment:"Path to the SSH known_hosts file."`
}

// DefaultSlipstreamConfig returns a Slipstream configuration
// with recommended defaults.
//
// Domain is deployment-specific and must be provided by the user.
func DefaultSlipstreamConfig() SlipstreamConfig {
	return SlipstreamConfig{
		Domain:            "",
		ResolverPort:      53,
		CertPath:          "",
		DNSResolution:     DNSResolutionRecursive,
		GSO:               false,
		CongestionControl: CongestionControlCubic,
		KeepAliveInterval: 10,
		ProxyPort:         1080,
		ProxyType:         ResolverProxySOCKS,
		AuthMethod:        AuthNone,
		Username:          "",
		Password:          "",
		PrivateKey:        "",
		KnownHostsFile:    "",
	}
}

func (c SlipstreamConfig) Validate() map[string]error {
	errs := make(map[string]error)

	if err := netutil.ValidateDomain(c.Domain); err != nil {
		errs["domain"] = err
	}

	if c.ResolverPort == 0 {
		errs["dns_port"] = fmt.Errorf("DNS port must be greater than zero")
	}

	if !c.DNSResolution.IsValid() {
		errs["dns_resolution"] = fmt.Errorf("must be %q or %q", DNSResolutionRecursive, DNSResolutionAuthoritative)
	}

	if !c.CongestionControl.IsValid() {
		errs["congestion_control"] = fmt.Errorf("must be %q or %q", CongestionControlBBR, CongestionControlCubic)
	}

	for field, err := range validateProxyAuth(
		c.ProxyType, c.ProxyPort, c.AuthMethod,
		c.Username, c.Password, c.PrivateKey, c.KnownHostsFile,
	) {
		errs[field] = err
	}

	return errs
}

// SlipstreamService manages Slipstream configurations and tunnels.
type SlipstreamService interface {
	SaveConfig(config SlipstreamConfig, name string) error
	EditConfig(config SlipstreamConfig, originalName string) error
	LoadConfig(name string) (SlipstreamConfig, error)
	GetAllConfigFiles() ([]SlipstreamConfigFile, error)
	ValidateAllConfigs() ([]ConfigValidationResult, error)
	RenameConfig(oldName, newName string) error
	RunTunnel(
		ctx context.Context,
		config SlipstreamConfig,
		resolverIP netip.Addr,
		listenPort uint16,
		keepAlive uint16,
	) (slipstream.Client, error)
}

// slipstreamStarter starts a Slipstream tunnel. *slipstream.Lib implements
// it; tests inject a fake.
type slipstreamStarter interface {
	Start(slipstream.Config) (slipstream.Client, error)
}

// slipstreamService is the default SlipstreamService implementation.
type slipstreamService struct {
	configs configStore[SlipstreamConfig]
	slip    slipstreamStarter
}

// SlipstreamServiceOption configures a Slipstream service.
type SlipstreamServiceOption func(*slipstreamService)

// WithSlipstreamDir sets the directory used to store Slipstream configurations.
func WithSlipstreamDir(dir string) SlipstreamServiceOption {
	return func(service *slipstreamService) {
		if dir != "" {
			service.configs.dir = dir
		}
	}
}

// WithSlipstreamClientLib uses lib as the loaded libslipstream library
// instead of loading it from the known locations.
func WithSlipstreamClientLib(lib *slipstream.Lib) SlipstreamServiceOption {
	return func(service *slipstreamService) {
		if lib != nil {
			service.slip = lib
		}
	}
}

// WithSlipstreamStarter replaces the tunnel starter. It allows tests to run
// without loading the real shared library.
func WithSlipstreamStarter(start slipstreamStarter) SlipstreamServiceOption {
	return func(service *slipstreamService) {
		if start != nil {
			service.slip = start
		}
	}
}

// NewSlipstreamService loads libslipstream before returning the service,
// unless WithSlipstreamClientLib (or WithSlipstreamStarter) is provided.
func NewSlipstreamService(
	opts ...SlipstreamServiceOption,
) (SlipstreamService, error) {
	service := &slipstreamService{
		configs: newConfigStore[SlipstreamConfig](tunnelConfigDir(slipstreamDir), "Slipstream"),
	}

	for _, opt := range opts {
		opt(service)
	}

	if service.slip == nil {
		slip, err := slipstream.Load()
		if err != nil {
			return nil, err
		}

		service.slip = slip
	}

	return service, nil
}

func (s *slipstreamService) SaveConfig(config SlipstreamConfig, name string) error {
	return s.configs.SaveConfig(config, name)
}

func (s *slipstreamService) EditConfig(config SlipstreamConfig, originalName string) error {
	return s.configs.EditConfig(config, originalName)
}

func (s *slipstreamService) LoadConfig(name string) (SlipstreamConfig, error) {
	return s.configs.LoadConfig(name)
}

func (s *slipstreamService) GetAllConfigFiles() ([]SlipstreamConfigFile, error) {
	return s.configs.GetAllConfigFiles()
}

func (s *slipstreamService) ValidateAllConfigs() ([]ConfigValidationResult, error) {
	return s.configs.ValidateAllConfigs()
}

func (s *slipstreamService) RenameConfig(oldName, newName string) error {
	return s.configs.RenameConfig(oldName, newName)
}

// RunTunnel starts a Slipstream DNS tunnel via the embedded library.
//
// keepAlive overrides config.KeepAliveInterval when non-zero; pass zero to
// use the config value (zero ultimately means "library default").
func (s *slipstreamService) RunTunnel(
	ctx context.Context,
	config SlipstreamConfig,
	resolverIP netip.Addr,
	listenPort uint16,
	keepAlive uint16,
) (slipstream.Client, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if errs := config.Validate(); len(errs) > 0 {
		return nil, fmt.Errorf("invalid Slipstream config: %v", errs)
	}

	if !resolverIP.IsValid() {
		return nil, fmt.Errorf("resolver IP is not valid")
	}

	if listenPort == 0 {
		return nil, fmt.Errorf("listen port must be greater than zero")
	}

	if s.slip == nil {
		return nil, fmt.Errorf("slipstream library is not loaded")
	}

	if keepAlive == 0 {
		keepAlive = config.KeepAliveInterval
	}

	cfg := config.LibConfig(resolverIP, netutil.Loopback(resolverIP), listenPort, keepAlive)

	return s.slip.Start(cfg)
}

// LibConfig builds the library configuration for a tunnel run.
// Invalid resolver/listen addresses produce empty host strings; callers
// should validate beforehand via RunTunnel.
func (c *SlipstreamConfig) LibConfig(
	resolverAddr netip.Addr,
	listenAddr netip.Addr,
	listenPort uint16,
	keepAlive uint16,
) slipstream.Config {
	if c == nil {
		return slipstream.Config{}
	}

	resolverHost := ""
	if resolverAddr.IsValid() {
		resolverHost = resolverAddr.String()
	}

	listenHost := ""
	if listenAddr.IsValid() {
		listenHost = listenAddr.String()
	}

	return slipstream.Config{
		Resolvers: []slipstream.Resolver{
			{
				Host: resolverHost,
				Port: c.ResolverPort,
				Mode: c.DNSResolution.libValue(),
			},
		},
		ListenHost:        listenHost,
		ListenPort:        listenPort,
		KeepAliveInterval: keepAlive,
		Domain:            c.Domain,
		CongestionControl: string(c.CongestionControl),
		GSO:               c.GSO,
		CertPath:          c.CertPath,
	}
}
