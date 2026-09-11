package dns

import (
	"context"
	"fmt"
	"io"
	"masterdnsvpn-go/pkg/tunnel"
	"net/netip"
	"strings"

	"github.com/MohsenBg/bgscan/internal/core/netutil"
)

const masterdnsDir = "masterdns"

// MasterDNSConfigFile pairs a MasterDNS configuration with its on-disk
// identity.
type MasterDNSConfigFile = ConfigFile[MasterDNSConfig]

// MasterDNSConfig contains the configuration required by the MasterDNS client.
//
// The resolver address and listen port are provided at runtime by RunTunnel.
type MasterDNSConfig struct {
	Domain        string
	DataEncMethod EncMethod
	EncryptionKey string
	ResolverPort  int

	MTUTestTimeoutSec      float64
	MTUTestRetries         int
	SessionInitRetryMaxSec float64
	SessionInitRacingCount int

	MinUploadMTU   int
	MaxUploadMTU   int
	MinDownloadMTU int
	MaxDownloadMTU int

	MTUParallelism int
	RxTxWorkers    int
}

// DefaultMasterDNSConfig returns a MasterDNS configuration with recommended
// defaults. Domain and EncryptionKey must be provided for a usable config.
func DefaultMasterDNSConfig() MasterDNSConfig {
	return MasterDNSConfig{
		Domain:        "",
		EncryptionKey: "",
		DataEncMethod: EncXOR,
		ResolverPort:  53,

		MTUTestTimeoutSec:      2.0,
		MTUTestRetries:         2,
		SessionInitRetryMaxSec: 60.0,
		SessionInitRacingCount: 3,

		MinUploadMTU:   38,
		MaxUploadMTU:   150,
		MinDownloadMTU: 100,
		MaxDownloadMTU: 500,
	}
}

// Validate validates the MasterDNS configuration.
func (c MasterDNSConfig) Validate() map[string]error {
	errs := make(map[string]error)

	if err := netutil.ValidateDomain(strings.TrimSpace(c.Domain)); err != nil {
		errs["domain"] = err
	}

	if strings.TrimSpace(c.EncryptionKey) == "" {
		errs["encryption_key"] = fmt.Errorf("encryption key is required")
	}

	if c.DataEncMethod < 0 || c.DataEncMethod > 5 {
		errs["enc_method"] = fmt.Errorf("must be between 0 and 5")
	}

	if c.ResolverPort == 0 {
		errs["resolver_port"] = fmt.Errorf("must be greater than zero")
	}

	return errs
}

// MasterDNSService manages MasterDNS configurations and tunnels.
type MasterDNSService interface {
	SaveConfig(config MasterDNSConfig, name string) error
	EditConfig(config MasterDNSConfig, originalName string) error
	LoadConfig(name string) (MasterDNSConfig, error)
	GetAllConfigFiles() ([]MasterDNSConfigFile, error)
	ValidateAllConfigs() ([]ConfigValidationResult, error)
	RenameConfig(oldName, newName string) error
	RunTunnel(
		ctx context.Context,
		config MasterDNSConfig,
		resolverIP string,
		listenPort uint16,
	) (io.Closer, error)
}

type masterDNSService struct {
	configs configStore[MasterDNSConfig]
}

// MasterDNSServiceOption configures a MasterDNSService.
type MasterDNSServiceOption func(*masterDNSService)

// WithMasterDNSDir sets the directory used to store MasterDNS configurations.
func WithMasterDNSDir(dir string) MasterDNSServiceOption {
	return func(service *masterDNSService) {
		if dir != "" {
			service.configs.dir = dir
		}
	}
}

// NewMasterDNSService creates a MasterDNS service.
func NewMasterDNSService(options ...MasterDNSServiceOption) MasterDNSService {
	service := &masterDNSService{
		configs: newConfigStore[MasterDNSConfig](tunnelConfigDir(masterdnsDir), "MasterDNS"),
	}

	for _, option := range options {
		option(service)
	}

	tunnel.DiscardLogger()
	return service
}

func (s *masterDNSService) SaveConfig(config MasterDNSConfig, name string) error {
	return s.configs.SaveConfig(config, name)
}

func (s *masterDNSService) EditConfig(config MasterDNSConfig, originalName string) error {
	return s.configs.EditConfig(config, originalName)
}

func (s *masterDNSService) LoadConfig(name string) (MasterDNSConfig, error) {
	return s.configs.LoadConfig(name)
}

func (s *masterDNSService) GetAllConfigFiles() ([]MasterDNSConfigFile, error) {
	return s.configs.GetAllConfigFiles()
}

func (s *masterDNSService) ValidateAllConfigs() ([]ConfigValidationResult, error) {
	return s.configs.ValidateAllConfigs()
}

func (s *masterDNSService) RenameConfig(oldName, newName string) error {
	return s.configs.RenameConfig(oldName, newName)
}

// RunTunnel starts an in-process MasterDNS client for the given resolver IP.
// The tunnel runs embedded: it exposes a local SOCKS5 listener on listenPort.
// The caller must Close the returned handle, otherwise goroutines and
// sockets leak.
func (s *masterDNSService) RunTunnel(
	ctx context.Context,
	config MasterDNSConfig,
	resolverIP string,
	listenPort uint16,
) (io.Closer, error) {
	if errs := config.Validate(); len(errs) > 0 {
		return nil, fmt.Errorf("invalid MasterDNS config: %v", errs)
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context canceled before tunnel setup: %w", err)
	}

	addr, err := parseTunnelEndpoint(resolverIP, listenPort)
	if err != nil {
		return nil, err
	}

	return startMasterTunnel(ctx, config, addr, listenPort)
}

// startMasterTunnel builds the internal client config from a
// MasterDNSConfig and boots the embedded client runtime against
// resolverIP. listenPort controls the local SOCKS5 listener port.
func startMasterTunnel(
	ctx context.Context,
	cfg MasterDNSConfig,
	resolverIP netip.Addr,
	listenPort uint16,
) (*embeddedHandle[*tunnel.Client], error) {
	msdCfg := tunnel.DefaultClientConfig()

	msdCfg.ListenPort = int(listenPort)

	msdCfg.Domains = []string{cfg.Domain}
	msdCfg.DataEncryptionMethod = int(cfg.DataEncMethod)
	msdCfg.EncryptionKey = cfg.EncryptionKey

	msdCfg.Resolvers = []tunnel.ResolverAddress{{
		IP:   resolverIP.String(),
		Port: cfg.ResolverPort,
	}}

	msdCfg.MTUTestTimeout = cfg.MTUTestTimeoutSec
	msdCfg.MTUTestRetries = cfg.MTUTestRetries
	msdCfg.SessionInitRetryMaxSeconds = cfg.SessionInitRetryMaxSec
	msdCfg.SessionInitRacingCount = cfg.SessionInitRacingCount

	msdCfg.MinUploadMTU = cfg.MinUploadMTU
	msdCfg.MaxUploadMTU = cfg.MaxUploadMTU
	msdCfg.MinDownloadMTU = cfg.MinDownloadMTU
	msdCfg.MaxDownloadMTU = cfg.MaxDownloadMTU

	msdCfg.MTUTestParallelism = cfg.MTUParallelism
	msdCfg.RX_TX_Workers = cfg.RxTxWorkers

	msdCfg.TunnelProcessWorkers = 1
	msdCfg.AutoDisableTimeoutServers = true

	app, err := tunnel.Bootstrap(msdCfg, "")
	if err != nil {
		return nil, fmt.Errorf("bootstrap client: %w", err)
	}

	return bootEmbedded(ctx, app)
}
