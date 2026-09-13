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
	Domain        string    `toml:"domain" comment:"MasterDNS tunnel server domain."`
	DataEncMethod EncMethod `toml:"data_enc_method" comment:"Data encryption method. 0=None, 1=XOR, 2=ChaCha20, 3=AES-128-GCM, 4=AES-192-GCM, 5=AES-256-GCM."`
	EncryptionKey string    `toml:"encryption_key" comment:"Tunnel encryption key. Must match the server config."`
	ResolverPort  uint16    `toml:"resolver_port" comment:"Resolver port. Standard DNS uses 53."`

	MTUTestTimeoutSec      float64 `toml:"mtu_test_timeout_sec" comment:"MTU test timeout in seconds."`
	MTUTestRetries         uint8   `toml:"mtu_test_retries" comment:"MTU test retry attempts."`
	SessionInitRetryMaxSec float64 `toml:"session_init_retry_max_sec" comment:"Maximum session init retry window in seconds."`
	SessionInitRacingCount uint8   `toml:"session_init_racing_count" comment:"Number of parallel session init attempts."`

	MinUploadMTU   uint16 `toml:"min_upload_mtu" comment:"Minimum upload MTU."`
	MaxUploadMTU   uint16 `toml:"max_upload_mtu" comment:"Maximum upload MTU."`
	MinDownloadMTU uint16 `toml:"min_download_mtu" comment:"Minimum download MTU."`
	MaxDownloadMTU uint16 `toml:"max_download_mtu" comment:"Maximum download MTU."`

	MTUParallelism uint8 `toml:"mtu_parallelism" comment:"Number of parallel MTU test workers."`
	RxTxWorkers    uint8 `toml:"rx_tx_workers" comment:"Number of RX/TX worker goroutines."`
}

// DefaultMasterDNSConfig returns a MasterDNS configuration with recommended
// defaults. Domain and EncryptionKey must be provided for a usable config.
func DefaultMasterDNSConfig() MasterDNSConfig {
	return MasterDNSConfig{
		Domain:        "",
		EncryptionKey: "",
		DataEncMethod: EncXOR,
		ResolverPort:  53,

		MTUTestTimeoutSec:      3.0,
		MTUTestRetries:         2,
		SessionInitRetryMaxSec: 2.0,
		SessionInitRacingCount: 1,

		MinUploadMTU:   38,
		MaxUploadMTU:   150,
		MinDownloadMTU: 100,
		MaxDownloadMTU: 500,

		MTUParallelism: 1,
		RxTxWorkers:    4,
	}
}

// Validate validates the MasterDNS configuration.
func (c MasterDNSConfig) Validate() map[string]error {
	errs := make(map[string]error)

	if err := netutil.ValidateDomain(strings.TrimSpace(c.Domain)); err != nil {
		errs["domain"] = err
	}

	if c.DataEncMethod < 0 || c.DataEncMethod > 5 {
		errs["enc_method"] = fmt.Errorf("must be between 0 and 5")
	}

	if c.DataEncMethod != 0 {
		if err := validateEncryptionKey(c.EncryptionKey); err != nil {
			errs["encryption_key"] = err
		}
	}
	if c.ResolverPort == 0 {
		errs["resolver_port"] = fmt.Errorf("must be between 1 and 65535")
	}

	if c.MTUTestTimeoutSec < 0 || c.MTUTestTimeoutSec > 60 {
		errs["mtu_test_timeout_sec"] = fmt.Errorf("must be 0 (default) or between 1 and 60")
	}

	if c.MTUTestRetries > 50 {
		errs["mtu_test_retries"] = fmt.Errorf("must be between 0 and 50")
	}

	if c.SessionInitRetryMaxSec < 0 || c.SessionInitRetryMaxSec > 3600 {
		errs["session_init_retry_max_sec"] = fmt.Errorf("must be between 0 and 3600")
	}

	if c.SessionInitRacingCount < 1 || c.SessionInitRacingCount > 5 {
		errs["session_init_racing_count"] = fmt.Errorf("must be between 1 and 5")
	}

	if c.MinUploadMTU < 10 || c.MinUploadMTU > c.MaxUploadMTU {
		errs["min_upload_mtu"] = fmt.Errorf("must be between 10 and max_upload_mtu")
	}

	if c.MaxUploadMTU < 1 || c.MaxUploadMTU > 255 {
		errs["max_upload_mtu"] = fmt.Errorf("must be between 1 and 255")
	}

	if c.MinDownloadMTU < 1 || c.MinDownloadMTU > c.MaxDownloadMTU {
		errs["min_download_mtu"] = fmt.Errorf("must be between 1 and max_download_mtu")
	}

	if c.MaxDownloadMTU < 1 {
		errs["max_download_mtu"] = fmt.Errorf("must be between 1 and 65535")
	}

	if c.MTUParallelism < 1 || c.MTUParallelism > 100 {
		errs["mtu_parallelism"] = fmt.Errorf("must be between 1 and 100")
	}

	if c.RxTxWorkers < 1 || c.RxTxWorkers > 64 {
		errs["rx_tx_workers"] = fmt.Errorf("must be between 1 and 64")
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
		Port: int(cfg.ResolverPort),
	}}

	msdCfg.MTUTestTimeout = cfg.MTUTestTimeoutSec
	msdCfg.MTUTestRetries = int(cfg.MTUTestRetries)
	msdCfg.SessionInitRetryMaxSeconds = cfg.SessionInitRetryMaxSec
	msdCfg.SessionInitRacingCount = int(cfg.SessionInitRacingCount)

	msdCfg.MinUploadMTU = int(cfg.MinUploadMTU)
	msdCfg.MaxUploadMTU = int(cfg.MaxUploadMTU)
	msdCfg.MinDownloadMTU = int(cfg.MinDownloadMTU)
	msdCfg.MaxDownloadMTU = int(cfg.MaxDownloadMTU)

	msdCfg.MTUTestParallelism = int(cfg.MTUParallelism)
	msdCfg.RX_TX_Workers = int(cfg.RxTxWorkers)

	msdCfg.TunnelProcessWorkers = 1
	msdCfg.AutoDisableTimeoutServers = true

	app, err := tunnel.Bootstrap(msdCfg, "")
	if err != nil {
		return nil, fmt.Errorf("bootstrap client: %w", err)
	}

	return bootEmbedded(ctx, app)
}
