package dns

import (
	"context"
	"fmt"
	"math"
	"net"
	"net/netip"
	"os"
	"path/filepath"
	"strings"
	"time"

	"bgscan/internal/core/fileutil"
	"bgscan/internal/core/netutil"

	vaydns "github.com/net2share/vaydns/client"
)

const dnsttDir = "assets/dns-tunneling/dnstt/"

// DNSTTConfig contains the configuration required by the DNSTT client.
//
// ResolverAddr is intentionally absent: the resolver address is provided
// at runtime per target via NewTunnel, not stored in the config.
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

type DNSTTConfigFile struct {
	Name      string
	Path      string
	CreatedAt time.Time
	Config    DNSTTConfig
}

type dnsttConn struct {
	net.Conn
	tunnel *vaydns.Tunnel
}

func (c *dnsttConn) Close() error {
	var err error

	if c.tunnel != nil {
		if tunnelErr := c.tunnel.Close(); tunnelErr != nil {
			err = tunnelErr
		}
	}

	return err
}

// DefaultDNSTTConfig returns a DNSTT configuration with recommended defaults.
//
// Domain and PubKey are deployment-specific and must be provided by the user.
func DefaultDNSTTConfig() DNSTTConfig {
	return DNSTTConfig{
		Domain:         "",
		ProxyPort:      1080,
		ResolverType:   ResolverType(vaydns.ResolverTypeUDP),
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

	domain := strings.TrimSpace(c.Domain)
	if err := netutil.ValidateDomain(domain); err != nil {
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

	if math.IsNaN(c.RPS) ||
		math.IsInf(c.RPS, 0) ||
		c.RPS < 0 ||
		c.RPS > 500 {
		errs["rps"] = fmt.Errorf("RPS must be between 0 and 500")
	}

	fingerprint := strings.TrimSpace(c.Fingerprint)
	if fingerprint == "" {
		errs["fingerprint"] = fmt.Errorf("TLS fingerprint is required")
	} else if _, err := parseClientHelloID(fingerprint); err != nil {
		errs["fingerprint"] = fmt.Errorf("invalid TLS fingerprint: %w", err)
	}

	if c.ProxyType != "" {
		if c.ProxyPort == 0 {
			errs["proxy_port"] = fmt.Errorf("proxy port must be greater than zero")
		}

		switch c.ProxyType {
		case ResolverProxySSH:
			if c.AuthMethod == AuthNone {
				errs["auth_method"] = fmt.Errorf("authentication is required for SSH proxy")
			}
		case ResolverProxySOCKS:
			if c.AuthMethod == AuthKey {
				errs["auth_method"] = fmt.Errorf("key auth is not supported for SOCKS proxy")
			}
		default:
			errs["proxy_type"] = fmt.Errorf("proxy type must be socks or ssh")
		}
	}

	if c.AuthMethod == AuthPassword {
		if strings.TrimSpace(c.Username) == "" {
			errs["username"] = fmt.Errorf("username is required for password auth")
		}
		if strings.TrimSpace(c.Password) == "" {
			errs["password"] = fmt.Errorf("password is required for password auth")
		}
	}

	if c.AuthMethod == AuthKey {
		if strings.TrimSpace(c.Username) == "" {
			errs["username"] = fmt.Errorf("username is required for key auth")
		}
		if err := validatePrivateKey(c.PrivateKey); err != nil {
			errs["private_key"] = err
		}
		if c.KnownHostsFile != "" {
			if err := validateKnownHostsFile(c.KnownHostsFile); err != nil {
				errs["known_hosts_file"] = err
			}
		}
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
	dir string
}

// DNSTTServiceOption configures a DNSTTService.
type DNSTTServiceOption func(*dnsttService)

// WithDNSTTDir sets the directory used to store DNSTT configurations.
func WithDNSTTDir(dir string) DNSTTServiceOption {
	return func(service *dnsttService) {
		if dir != "" {
			service.dir = dir
		}
	}
}

// NewDNSTTService creates a DNSTT service.
func NewDNSTTService(options ...DNSTTServiceOption) DNSTTService {
	service := &dnsttService{
		dir: getDNSTTDir(),
	}

	for _, option := range options {
		option(service)
	}

	return service
}

// getDNSTTDir returns the default directory used to store DNSTT configs.
func getDNSTTDir() string {
	base, err := fileutil.BasePath()
	if err != nil {
		return dnsttDir
	}

	return filepath.Join(base, dnsttDir)
}

// SaveConfig saves a DNSTT configuration to disk.
// It returns an error if a config with the given name already exists.
func (s *dnsttService) SaveConfig(
	config DNSTTConfig,
	name string,
) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("config name is required")
	}

	path := s.configPath(name)

	if fileutil.CheckFileExists(path) {
		return fmt.Errorf("config %q already exists", name)
	}

	if errs := config.Validate(); len(errs) > 0 {
		return fmt.Errorf("invalid DNSTT config: %v", errs)
	}

	if err := fileutil.WriteTOMLFile(path, config); err != nil {
		return fmt.Errorf(
			"save DNSTT config %q: %w",
			name,
			err,
		)
	}

	return nil
}

// EditConfig updates an existing DNSTT configuration identified by originalName.
func (s *dnsttService) EditConfig(config DNSTTConfig, name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("original config name is required")
	}

	path := s.configPath(name)

	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("config %q does not exist", name)
	}

	if errs := config.Validate(); len(errs) > 0 {
		return fmt.Errorf("invalid DNSTT config: %v", errs)
	}

	if err := fileutil.WriteTOMLFile(path, config); err != nil {
		return fmt.Errorf("edit DNSTT config %q: %w", name, err)
	}

	return nil
}

// LoadConfig loads a DNSTT configuration from disk.
func (s *dnsttService) LoadConfig(name string) (DNSTTConfig, error) {
	path := s.configPath(name)

	config, err := fileutil.ReadTOMLFile[DNSTTConfig](path)
	if err != nil {
		return DNSTTConfig{}, fmt.Errorf(
			"load DNSTT config %q: %w",
			name,
			err,
		)
	}

	return config, nil
}

// GetAllConfigFiles returns all valid DNSTT TOML configuration files.
func (s *dnsttService) GetAllConfigFiles() ([]DNSTTConfigFile, error) {
	if err := fileutil.EnsureDir(s.dir); err != nil {
		return nil, err
	}

	files, err := fileutil.ListFiles(s.dir, func(name string, _ os.FileInfo) bool {
		return strings.EqualFold(filepath.Ext(name), ".toml")
	})
	if err != nil {
		return nil, err
	}

	configs := make([]DNSTTConfigFile, 0, len(files))

	for _, file := range files {
		cfg, err := s.LoadConfig(file.Name)
		if err != nil || len(cfg.Validate()) != 0 {
			continue
		}

		configs = append(configs, DNSTTConfigFile{
			Name:      strings.TrimSuffix(file.Name, filepath.Ext(file.Name)),
			Path:      file.Path,
			CreatedAt: file.Info.ModTime(),
			Config:    cfg,
		})
	}

	return configs, nil
}

func (s *dnsttService) ValidateAllConfigs() ([]ConfigValidationResult, error) {
	if err := fileutil.EnsureDir(s.dir); err != nil {
		return nil, err
	}
	files, err := fileutil.ListFiles(s.dir, func(name string, _ os.FileInfo) bool {
		return strings.EqualFold(filepath.Ext(name), ".toml")
	})
	if err != nil {
		return nil, err
	}

	results := make([]ConfigValidationResult, 0, len(files))

	for _, file := range files {
		cfg, err := s.LoadConfig(file.Name)
		if err != nil {
			results = append(results, ConfigValidationResult{
				File:   file,
				Errors: map[string]error{},
			})
			continue
		}

		validationErrors := cfg.Validate()
		if len(validationErrors) == 0 {
			continue
		}

		results = append(results, ConfigValidationResult{
			File:   file,
			Errors: validationErrors,
		})
	}

	return results, nil
}

// NewTunnel creates and initializes a DNSTT tunnel to the given resolver address.
func (s *dnsttService) NewTunnel(
	ctx context.Context,
	config DNSTTConfig,
	resolverAddr netip.Addr,
) (net.Conn, error) {
	if errs := config.Validate(); len(errs) > 0 {
		return nil, fmt.Errorf("invalid DNSTT config: %v", errs)
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf(
			"context canceled before tunnel setup: %w",
			err,
		)
	}

	resolver, err := newDNSTTResolver(config, resolverAddr)
	if err != nil {
		return nil, err
	}

	tunnelServer, err := newDNSTTTunnelServer(config)
	if err != nil {
		return nil, err
	}

	tunnel, err := vaydns.NewTunnel(*resolver, *tunnelServer)
	if err != nil {
		return nil, fmt.Errorf(
			"create tunnel: %w",
			err,
		)
	}

	// From this point, tunnel owns resources.
	// Every failure path must close it.

	if err := ctx.Err(); err != nil {
		_ = tunnel.Close()
		return nil, fmt.Errorf(
			"context canceled before resolver connection: %w",
			err,
		)
	}

	if err := tunnel.InitiateResolverConnection(ctx); err != nil {
		_ = tunnel.Close()
		return nil, fmt.Errorf(
			"initiate resolver connection: %w",
			err,
		)
	}

	if err := ctx.Err(); err != nil {
		_ = tunnel.Close()
		return nil, fmt.Errorf(
			"context canceled before DNS packet connection: %w",
			err,
		)
	}

	if err := tunnel.InitiateDNSPacketConn(
		ctx,
		tunnelServer.Addr,
	); err != nil {
		_ = tunnel.Close()
		return nil, fmt.Errorf(
			"initiate DNS packet connection: %w",
			err,
		)
	}

	if err := ctx.Err(); err != nil {
		_ = tunnel.Close()
		return nil, fmt.Errorf(
			"context canceled before KCP connection: %w",
			err,
		)
	}

	if err := tunnel.InitiateKCPConn(
		tunnelServer.MTU,
	); err != nil {
		_ = tunnel.Close()
		return nil, fmt.Errorf(
			"initiate KCP connection: %w",
			err,
		)
	}

	if err := ctx.Err(); err != nil {
		_ = tunnel.Close()
		return nil, fmt.Errorf(
			"context canceled before Noise channel: %w",
			err,
		)
	}

	if err := tunnel.InitiateNoiseChannel(ctx); err != nil {
		_ = tunnel.Close()
		return nil, fmt.Errorf(
			"initiate Noise channel: %w",
			err,
		)
	}

	if err := ctx.Err(); err != nil {
		_ = tunnel.Close()
		return nil, fmt.Errorf(
			"context canceled before smux session: %w",
			err,
		)
	}

	if err := tunnel.InitiateSmuxSession(); err != nil {
		_ = tunnel.Close()
		return nil, fmt.Errorf(
			"initiate smux session: %w",
			err,
		)
	}

	if err := ctx.Err(); err != nil {
		_ = tunnel.Close()
		return nil, fmt.Errorf(
			"context canceled before opening stream: %w",
			err,
		)
	}

	stream, err := tunnel.OpenStream()
	if err != nil {
		_ = tunnel.Close()
		return nil, fmt.Errorf(
			"open tunnel stream: %w",
			err,
		)
	}

	return &dnsttConn{
		Conn:   stream,
		tunnel: tunnel,
	}, nil
}

// newDNSTTResolver creates a DNSTT resolver from the configuration.
func newDNSTTResolver(
	config DNSTTConfig,
	resolverAddr netip.Addr,
) (*vaydns.Resolver, error) {
	addr := netip.AddrPortFrom(
		resolverAddr,
		config.ResolverPort,
	).String()

	resolver, err := vaydns.NewResolver(
		vaydns.ResolverType(config.ResolverType),
		addr,
	)
	if err != nil {
		return nil, fmt.Errorf("create resolver: %w", err)
	}

	clientHelloID, err := parseClientHelloID(config.Fingerprint)
	if err != nil {
		return nil, fmt.Errorf(
			"parse TLS fingerprint: %w",
			err,
		)
	}

	resolver.UTLSClientHelloID = &clientHelloID
	resolver.UDPSharedSocket = true

	return &resolver, nil
}

// newDNSTTTunnelServer creates a DNSTT-compatible tunnel server.
func newDNSTTTunnelServer(
	config DNSTTConfig,
) (*vaydns.TunnelServer, error) {
	server, err := vaydns.NewTunnelServer(
		config.Domain,
		normalizeConfigName(config.PubKey),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"create tunnel server: %w",
			err,
		)
	}

	server.RPS = config.RPS
	server.DnsttCompat = true

	return &server, nil
}

// configPath returns the path for a named DNSTT configuration.
func (s *dnsttService) configPath(name string) string {
	return filepath.Join(
		s.dir,
		normalizeConfigName(name)+".toml",
	)
}

func (s *dnsttService) RenameConfig(oldName, newName string) error {
	oldPath := s.configPath(oldName)
	newPath := s.configPath(newName)

	if _, err := os.Stat(oldPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("config %q does not exist", oldName)
		}

		return fmt.Errorf("check current config: %w", err)
	}

	if _, err := os.Stat(newPath); err == nil {
		return fmt.Errorf("config %q already exists", newName)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("check destination config: %w", err)
	}

	return fileutil.RenameFile(oldPath, newPath)
}
