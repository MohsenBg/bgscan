package dns

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/MohsenBg/bgscan/internal/core/fileutil"
	"github.com/MohsenBg/bgscan/internal/core/netutil"
	"github.com/MohsenBg/bgscan/internal/core/process"
)

const slipstreamDir = "slipstream"

// SlipstreamConfigFile pairs a Slipstream configuration with its on-disk
// identity.
type SlipstreamConfigFile = ConfigFile[SlipstreamConfig]

// SlipstreamConfig contains the configuration required by the
// Slipstream DNS client.
type SlipstreamConfig struct {
	Domain       string `toml:"domain" comment:"Slipstream server domain."`
	ResolverPort uint16 `toml:"resolver_port" comment:"Resolver port. Standard DNS uses 53."`
	CertPath     string `toml:"cert_path" comment:"Path to the CA certificate used to verify the server. Empty = system pool."`

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
		Domain:         "",
		ResolverPort:   53,
		CertPath:       "",
		ProxyPort:      1080,
		ProxyType:      ResolverProxySOCKS,
		AuthMethod:     AuthNone,
		Username:       "",
		Password:       "",
		PrivateKey:     "",
		KnownHostsFile: "",
	}
}

// Validate validates the Slipstream configuration.
func (c SlipstreamConfig) Validate() map[string]error {
	errs := make(map[string]error)

	if err := netutil.ValidateDomain(c.Domain); err != nil {
		errs["domain"] = err
	}

	if c.ResolverPort == 0 {
		errs["dns_port"] = fmt.Errorf("DNS port must be greater than zero")
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
		resolverIP string,
		listenPort uint16,
	) (process.Process, error)
}

// processStarter starts the slipstream-client binary. A function type
// rather than an interface: the only variation is how the process is
// launched (tests replace it).
type processStarter func(
	context.Context,
	string,
	...string,
) (process.Process, error)

// slipstreamService is the default SlipstreamService implementation.
type slipstreamService struct {
	configs configStore[SlipstreamConfig]
	bin     string
	start   processStarter
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

// WithSlipstreamProcessStarter replaces the process launcher.
//
// It allows tests to run without starting a real slipstream-client binary.
func WithSlipstreamProcessStarter(start processStarter) SlipstreamServiceOption {
	return func(service *slipstreamService) {
		if start != nil {
			service.start = start
		}
	}
}

// WithSlipstreamClientBinary uses bin as the slipstream-client executable
// instead of searching the known locations.
func WithSlipstreamClientBinary(bin string) SlipstreamServiceOption {
	return func(service *slipstreamService) {
		if bin != "" {
			service.bin = bin
		}
	}
}

// NewSlipstreamService creates a Slipstream service.
//
// Unless WithSlipstreamClientBinary is provided, it locates
// slipstream-client before returning the service.
func NewSlipstreamService(
	opts ...SlipstreamServiceOption,
) (SlipstreamService, error) {
	service := &slipstreamService{
		configs: newConfigStore[SlipstreamConfig](tunnelConfigDir(slipstreamDir), "Slipstream"),
		start:   process.Start,
	}

	for _, opt := range opts {
		opt(service)
	}

	if service.bin == "" {
		bin, err := FindSlipstreamClient()
		if err != nil {
			return nil, err
		}

		service.bin = bin
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

// RunTunnel starts a Slipstream DNS tunnel.
func (s *slipstreamService) RunTunnel(
	ctx context.Context,
	config SlipstreamConfig,
	resolverIP string,
	listenPort uint16,
) (process.Process, error) {
	if errs := config.Validate(); len(errs) > 0 {
		return nil, fmt.Errorf("invalid Slipstream config: %v", errs)
	}

	if strings.TrimSpace(resolverIP) == "" {
		return nil, fmt.Errorf("resolver IP is required")
	}

	if listenPort == 0 {
		return nil, fmt.Errorf("listen port must be greater than zero")
	}

	args := []string{
		"-d", config.Domain,
		"-r", net.JoinHostPort(resolverIP, fmt.Sprint(config.ResolverPort)),
		"-l", fmt.Sprint(listenPort),
	}

	if config.CertPath != "" {
		args = append(args, "--cert", config.CertPath)
	}

	return s.start(ctx, s.bin, args...)
}

// FindSlipstreamClient locates slipstream-client in known locations
// or PATH.
func FindSlipstreamClient() (string, error) {
	return process.FindBinaryInPaths(
		"slipstream-client",
		getSlipstreamPaths(),
	)
}

// getSlipstreamPaths returns the locations searched for
// slipstream-client.
func getSlipstreamPaths() []string {
	base, err := fileutil.BasePath()
	if err != nil {
		return nil
	}

	return []string{
		filepath.Join(base, "assets", "slipstream-client"),
		filepath.Join(base, "assets", "slipstream", "slipstream-client"),
		filepath.Join(base, "slipstream-client"),
		base,
	}
}

// VerifySlipstreamClient verifies that slipstream-client can execute.
func VerifySlipstreamClient() error {
	path, err := FindSlipstreamClient()
	if err != nil {
		return fmt.Errorf("find slipstream-client: %w", err)
	}

	if err := exec.Command(path, "--help").Run(); err != nil {
		return fmt.Errorf("run slipstream-client: %w", err)
	}

	return nil
}
