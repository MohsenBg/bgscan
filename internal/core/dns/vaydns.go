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

const vaydnsDir = "assets/dns-tunneling/vaydns/"

// VayDNSConfig contains the configuration required to create a VayDNS tunnel.
//
// ResolverAddr is intentionally absent: the resolver address is provided
// at runtime per target via NewTunnel, not stored in the config.
type VayDNSConfig struct {
	Domain       string
	PubKey       string
	ClientIDSize uint16
	MaxQnameLen  uint8
	MaxNumLabels uint8
	MTU          uint16
	RPS          float64
	RecordType   RecordType

	ResolverType ResolverType
	ResolverPort uint16
	Fingerprint  string

	ProxyType      ResolverProxyType
	ProxyPort      uint16
	AuthMethod     AuthMethod
	Username       string
	Password       string
	PrivateKey     string
	KnownHostsFile string
}

type VayDNSConfigFile struct {
	Name      string
	Path      string
	CreatedAt time.Time
	Config    VayDNSConfig
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

	if math.IsNaN(c.RPS) || math.IsInf(c.RPS, 0) || c.RPS < 0 || c.RPS > 500 {
		errs["rps"] = fmt.Errorf("must be between 0 and 500")
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
	dir string
}

// VayDNSServiceOption configures a VayDNSService.
type VayDNSServiceOption func(*vayDNSService)

// WithVayDNSDir sets the directory used to store VayDNS configurations.
func WithVayDNSDir(dir string) VayDNSServiceOption {
	return func(service *vayDNSService) {
		if dir != "" {
			service.dir = dir
		}
	}
}

// NewVayDNSService creates a VayDNS service.
func NewVayDNSService(options ...VayDNSServiceOption) VayDNSService {
	service := &vayDNSService{dir: getVayDNSDir()}

	for _, option := range options {
		option(service)
	}

	return service
}

// getVayDNSDir returns the default directory used to store VayDNS configs.
func getVayDNSDir() string {
	base, err := fileutil.BasePath()
	if err != nil {
		return vaydnsDir
	}

	return filepath.Join(base, vaydnsDir)
}

// SaveConfig validates and saves a VayDNS configuration to disk.
// It returns an error if a config with the given name already exists.
func (s *vayDNSService) SaveConfig(config VayDNSConfig, name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("config name is required")
	}

	path := s.configPath(name)

	if fileutil.CheckFileExists(path) {
		return fmt.Errorf("config %q already exists", name)
	}

	if errs := config.Validate(); len(errs) > 0 {
		return fmt.Errorf("invalid VayDNS config: %v", errs)
	}

	if err := fileutil.WriteTOMLFile(path, config); err != nil {
		return fmt.Errorf("save VayDNS config %q: %w", name, err)
	}

	return nil
}

// EditConfig updates an existing VayDNS configuration identified by originalName.
func (s *vayDNSService) EditConfig(config VayDNSConfig, originalName string) error {
	if strings.TrimSpace(originalName) == "" {
		return fmt.Errorf("original config name is required")
	}

	path := s.configPath(originalName)

	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("config %q does not exist", originalName)
	}

	if errs := config.Validate(); len(errs) > 0 {
		return fmt.Errorf("invalid VayDNS config: %v", errs)
	}

	if err := fileutil.WriteTOMLFile(path, config); err != nil {
		return fmt.Errorf("edit VayDNS config %q: %w", originalName, err)
	}

	return nil
}

// LoadConfig loads a VayDNS configuration from disk.
func (s *vayDNSService) LoadConfig(name string) (VayDNSConfig, error) {
	config, err := fileutil.ReadTOMLFile[VayDNSConfig](s.configPath(name))
	if err != nil {
		return VayDNSConfig{}, fmt.Errorf("load VayDNS config %q: %w", name, err)
	}

	return config, nil
}

// GetAllConfigFiles returns valid VayDNS TOML configuration files.
func (s *vayDNSService) GetAllConfigFiles() ([]VayDNSConfigFile, error) {
	if err := fileutil.EnsureDir(s.dir); err != nil {
		return nil, err
	}
	files, err := fileutil.ListFiles(s.dir, func(name string, _ os.FileInfo) bool {
		return strings.EqualFold(filepath.Ext(name), ".toml")
	})
	if err != nil {
		return nil, err
	}

	configs := make([]VayDNSConfigFile, 0, len(files))

	for _, file := range files {
		cfg, err := s.LoadConfig(file.Name)
		if err != nil {
			continue
		}

		configs = append(configs, VayDNSConfigFile{
			Name:      strings.TrimSuffix(file.Name, filepath.Ext(file.Name)),
			Path:      file.Path,
			CreatedAt: file.Info.ModTime(),
			Config:    cfg,
		})
	}

	return configs, nil
}

func (s *vayDNSService) ValidateAllConfigs() ([]ConfigValidationResult, error) {
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

// vayDNSConn wraps the smux stream returned by an established VayDNS tunnel
type vayDNSConn struct {
	net.Conn
	tunnel *vaydns.Tunnel
}

func (c *vayDNSConn) Close() error {
	var tunnelErr error
	if c.tunnel != nil {
		tunnelErr = c.tunnel.Close()
	}

	return tunnelErr
}

func (s *vayDNSService) NewTunnel(ctx context.Context, config VayDNSConfig, resolverAddr netip.Addr) (net.Conn, error) {
	if errs := config.Validate(); len(errs) > 0 {
		return nil, fmt.Errorf("invalid VayDNS config: %v", errs)
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context canceled before tunnel setup: %w", err)
	}

	resolver, err := newVayDNSResolver(config, resolverAddr)
	if err != nil {
		return nil, err
	}

	tunnelServer, err := newVayDNSTunnelServer(config)
	if err != nil {
		return nil, err
	}

	tunnel, err := vaydns.NewTunnel(*resolver, *tunnelServer)
	if err != nil {
		return nil, fmt.Errorf("create tunnel: %w", err)
	}

	if err := ctx.Err(); err != nil {
		_ = tunnel.Close()
		return nil, fmt.Errorf("context canceled before resolver connection: %w", err)
	}
	if err := tunnel.InitiateResolverConnection(ctx); err != nil {
		_ = tunnel.Close()
		return nil, fmt.Errorf("initiate resolver connection: %w", err)
	}

	if err := ctx.Err(); err != nil {
		_ = tunnel.Close()
		return nil, fmt.Errorf("context canceled before DNS packet conn: %w", err)
	}
	if err := tunnel.InitiateDNSPacketConn(ctx, tunnelServer.Addr); err != nil {
		_ = tunnel.Close()
		return nil, fmt.Errorf("initiate DNS packet connection: %w", err)
	}

	if err := ctx.Err(); err != nil {
		_ = tunnel.Close()
		return nil, fmt.Errorf("context canceled before KCP connection: %w", err)
	}
	if err := tunnel.InitiateKCPConn(tunnelServer.MTU); err != nil {
		_ = tunnel.Close()
		return nil, fmt.Errorf("initiate KCP connection: %w", err)
	}

	if err := ctx.Err(); err != nil {
		_ = tunnel.Close()
		return nil, fmt.Errorf("context canceled before Noise channel: %w", err)
	}
	if err := tunnel.InitiateNoiseChannel(ctx); err != nil {
		_ = tunnel.Close()
		return nil, fmt.Errorf("initiate Noise channel: %w", err)
	}

	if err := ctx.Err(); err != nil {
		_ = tunnel.Close()
		return nil, fmt.Errorf("context canceled before smux session: %w", err)
	}
	if err := tunnel.InitiateSmuxSession(); err != nil {
		_ = tunnel.Close()
		return nil, fmt.Errorf("initiate smux session: %w", err)
	}

	stream, err := tunnel.OpenStream()
	if err != nil {
		_ = tunnel.Close()
		return nil, fmt.Errorf("open tunnel stream: %w", err)
	}

	// tunnel is wrapped alongside stream so Close() releases the full
	// stack (see vayDNSConn.Close), not just this one smux stream.
	return &vayDNSConn{Conn: stream, tunnel: tunnel}, nil
}

func newVayDNSResolver(config VayDNSConfig, resolverAddr netip.Addr) (*vaydns.Resolver, error) {
	addr := netip.AddrPortFrom(resolverAddr, config.ResolverPort).String()

	resolver, err := vaydns.NewResolver(
		toVayDNSResolverType(config.ResolverType),
		addr,
	)
	if err != nil {
		return nil, fmt.Errorf("create resolver: %w", err)
	}

	clientHelloID, err := parseClientHelloID(config.Fingerprint)
	if err != nil {
		return nil, fmt.Errorf("parse TLS fingerprint: %w", err)
	}

	resolver.UTLSClientHelloID = &clientHelloID
	resolver.UDPSharedSocket = true

	return &resolver, nil
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

func (s *vayDNSService) configPath(name string) string {
	return filepath.Join(s.dir, normalizeConfigName(name)+".toml")
}

func toVayDNSResolverType(t ResolverType) vaydns.ResolverType {
	switch t {
	case ResolverTypeTCP:
		return vaydns.ResolverTypeTCP
	case ResolverTypeDOT:
		return vaydns.ResolverTypeDOT
	case ResolverTypeUDP:
		return vaydns.ResolverTypeUDP
	default:
		return vaydns.ResolverTypeUDP
	}
}

func (s *vayDNSService) RenameConfig(oldName, newName string) error {
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
