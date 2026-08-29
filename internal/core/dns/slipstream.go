package dns

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"bgscan/internal/core/fileutil"
	"bgscan/internal/core/netutil"
	"bgscan/internal/core/process"
)

const slipstreamDir = "assets/dns-tunneling/slipstream/"

var ErrSlipstreamTunnelNotRunning = errors.New(
	"slipstream-client process is not running",
)

type ProcessStarter func(
	context.Context,
	string,
	...string,
) (process.Process, error)

// SlipstreamConfig contains the configuration required by the
// Slipstream DNS client.
type SlipstreamConfig struct {
	Domain       string
	ResolverPort uint16
	CertPath     string

	ProxyType  ResolverProxyType
	ProxyPort  uint16
	AuthMethod AuthMethod
	Username   string
	Password     string
	PrivateKey   string
	KnownHostsFile string
}

type SlipstreamConfigFile struct {
	Name      string
	Path      string
	CreatedAt time.Time
	Config    SlipstreamConfig
}

// DefaultSlipstreamConfig returns a Slipstream configuration
// with recommended defaults.
//
// Domain is deployment-specific and must be provided by the user.
func DefaultSlipstreamConfig() SlipstreamConfig {
	return SlipstreamConfig{
		Domain:       "",
		ResolverPort: 53,
		CertPath:     "",
		ProxyPort:    1080,
		ProxyType:    ResolverProxySOCKS,
		AuthMethod:   AuthNone,
		Username:     "",
		Password:       "",
		PrivateKey:     "",
		KnownHostsFile: "",
	}
}

// Validate validates the Slipstream configuration.
func (c SlipstreamConfig) Validate() map[string]error {
	errs := make(map[string]error)

	domain := strings.TrimSpace(c.Domain)
	if err := netutil.ValidateDomain(domain); err != nil {
		errs["domain"] = err
	}

	if c.ResolverPort == 0 {
		errs["dns_port"] = fmt.Errorf(
			"DNS port must be greater than zero",
		)
	}

	if c.ProxyType != "" {
		if c.ProxyPort == 0 {
			errs["proxy_port"] = fmt.Errorf(
				"proxy port is required when proxy is enabled",
			)
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

// slipstreamService is the default SlipstreamService implementation.
type slipstreamService struct {
	dir   string
	bin   string
	start ProcessStarter
}

// SlipstreamServiceOption configures a Slipstream service.
type SlipstreamServiceOption func(*slipstreamService)

// WithSlipstreamDir sets the directory used to store Slipstream configurations.
func WithSlipstreamDir(dir string) SlipstreamServiceOption {
	return func(service *slipstreamService) {
		if dir != "" {
			service.dir = dir
		}
	}
}

// WithSlipstreamProcessStarter replaces the process launcher.
//
// It allows tests to run without starting a real slipstream-client binary.
func WithSlipstreamProcessStarter(
	start ProcessStarter,
) SlipstreamServiceOption {
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
		dir:   getSlipstreamDir(),
		start: process.Start,
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

// getSlipstreamDir returns the default directory used to store
// Slipstream configurations.
func getSlipstreamDir() string {
	base, err := fileutil.BasePath()
	if err != nil {
		return slipstreamDir
	}

	return filepath.Join(base, slipstreamDir)
}

// SaveConfig saves a Slipstream configuration to disk.
// It returns an error if a config with the given name already exists.
func (s *slipstreamService) SaveConfig(
	config SlipstreamConfig,
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
		return fmt.Errorf(
			"invalid Slipstream config: %v",
			errs,
		)
	}

	if err := fileutil.WriteTOMLFile(path, config); err != nil {
		return fmt.Errorf(
			"save Slipstream config %q: %w",
			name,
			err,
		)
	}

	return nil
}

// EditConfig updates an existing Slipstream configuration identified by originalName.
func (s *slipstreamService) EditConfig(config SlipstreamConfig, originalName string) error {
	if strings.TrimSpace(originalName) == "" {
		return fmt.Errorf("original config name is required")
	}

	path := s.configPath(originalName)

	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("config %q does not exist", originalName)
	}

	if errs := config.Validate(); len(errs) > 0 {
		return fmt.Errorf(
			"invalid Slipstream config: %v",
			errs,
		)
	}

	if err := fileutil.WriteTOMLFile(path, config); err != nil {
		return fmt.Errorf(
			"edit Slipstream config %q: %w",
			originalName,
			err,
		)
	}

	return nil
}

// LoadConfig loads a Slipstream configuration from disk.
func (s *slipstreamService) LoadConfig(
	name string,
) (SlipstreamConfig, error) {
	path := s.configPath(name)

	config, err := fileutil.ReadTOMLFile[SlipstreamConfig](path)
	if err != nil {
		return SlipstreamConfig{}, fmt.Errorf(
			"load Slipstream config %q: %w",
			name,
			err,
		)
	}

	return config, nil
}

// GetAllConfigFiles returns all valid Slipstream TOML configuration files.
func (s *slipstreamService) GetAllConfigFiles() ([]SlipstreamConfigFile, error) {
	if err := fileutil.EnsureDir(s.dir); err != nil {
		return nil, err
	}
	files, err := fileutil.ListFiles(
		s.dir,
		func(name string, _ os.FileInfo) bool {
			return strings.EqualFold(filepath.Ext(name), ".toml")
		},
	)
	if err != nil {
		return nil, err
	}

	configs := make([]SlipstreamConfigFile, 0, len(files))

	for _, file := range files {
		cfg, err := s.LoadConfig(file.Name)
		if err != nil || len(cfg.Validate()) > 0 {
			continue
		}

		configs = append(configs, SlipstreamConfigFile{
			Name:      strings.TrimSuffix(file.Name, filepath.Ext(file.Name)),
			Path:      file.Path,
			CreatedAt: file.Info.ModTime(),
			Config:    cfg,
		})
	}

	return configs, nil
}

func (s *slipstreamService) ValidateAllConfigs() ([]ConfigValidationResult, error) {
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

// RunTunnel starts a Slipstream DNS tunnel.
func (s *slipstreamService) RunTunnel(
	ctx context.Context,
	config SlipstreamConfig,
	resolverIP string,
	listenPort uint16,
) (process.Process, error) {
	if errs := config.Validate(); len(errs) > 0 {
		return nil, fmt.Errorf(
			"invalid Slipstream config: %v",
			errs,
		)
	}

	if strings.TrimSpace(resolverIP) == "" {
		return nil, fmt.Errorf("resolver IP is required")
	}

	if listenPort == 0 {
		return nil, fmt.Errorf("listen port must be greater than zero")
	}

	args := []string{
		"-d",
		config.Domain,
		"-r",
		net.JoinHostPort(
			resolverIP,
			fmt.Sprint(config.ResolverPort),
		),
		"-l",
		fmt.Sprint(listenPort),
	}

	if config.CertPath != "" {
		args = append(
			args,
			"--cert",
			config.CertPath,
		)
	}

	proc, err := s.start(
		ctx,
		s.bin,
		args...,
	)
	if err != nil {
		return nil, err
	}

	return proc, nil
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
		filepath.Join(
			base,
			"assets",
			"slipstream-client",
		),
		filepath.Join(
			base,
			"assets",
			"slipstream",
			"slipstream-client",
		),
		filepath.Join(
			base,
			"slipstream-client",
		),
		base,
	}
}

// VerifySlipstreamClient verifies that slipstream-client can execute.
func VerifySlipstreamClient() error {
	path, err := FindSlipstreamClient()
	if err != nil {
		return fmt.Errorf(
			"find slipstream-client: %w",
			err,
		)
	}

	if err := exec.Command(path, "--help").Run(); err != nil {
		return fmt.Errorf(
			"run slipstream-client: %w",
			err,
		)
	}

	return nil
}

// configPath returns the path for a named Slipstream configuration.
func (s *slipstreamService) configPath(name string) string {
	return filepath.Join(
		s.dir,
		normalizeConfigName(name)+".toml",
	)
}

func (s *slipstreamService) RenameConfig(oldName, newName string) error {
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
