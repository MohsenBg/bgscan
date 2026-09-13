package dns

import (
	"context"
	"fmt"
	"io"
	"net/netip"
	"strings"

	stormtunnel "stormdns-go/pkg/tunnel"

	"github.com/MohsenBg/bgscan/internal/core/netutil"
)

const stormdnsDir = "stormdns"

// StormDNS query types supported for tunnel queries.
const (
	StormDNSQueryTXT    = "TXT"
	StormDNSQueryNS     = "NS"
	StormDNSQueryCNAME  = "CNAME"
	StormDNSQueryRotate = "ROTATE"
)

// StormDNSConfigFile pairs a StormDNS configuration with its on-disk
// identity.
type StormDNSConfigFile = ConfigFile[StormDNSConfig]

// StormDNSConfig contains the configuration required by the StormDNS client.
//
// The resolver address and listen port are provided at runtime by RunTunnel.
type StormDNSConfig struct {
	Domain        string    `toml:"domain" comment:"StormDNS tunnel server domain."`
	DataEncMethod EncMethod `toml:"data_enc_method" comment:"Data encryption method. 0=None, 1=XOR, 2=ChaCha20, 3=AES-128-GCM, 4=AES-192-GCM, 5=AES-256-GCM."`
	EncryptionKey string    `toml:"encryption_key" comment:"Tunnel encryption key. Must match the server config."`
	ResolverPort  uint16    `toml:"resolver_port" comment:"Resolver port. Standard DNS uses 53."`

	// DNSQueryType selects the tunnel query shape: TXT, NS, CNAME or
	// ROTATE. Empty means TXT. NS/CNAME/ROTATE require a server running
	// the same feature version.
	DNSQueryType string `toml:"dns_query_type" comment:"Tunnel query shape: TXT, NS, CNAME, or ROTATE. Empty = TXT."`

	MTUTestTimeoutSec      float64 `toml:"mtu_test_timeout_sec" comment:"MTU test timeout in seconds."`
	MTUTestRetries         uint8   `toml:"mtu_test_retries" comment:"MTU test retry attempts."`
	SessionInitRetryMaxSec float64 `toml:"session_init_retry_max_sec" comment:"Maximum session init retry window in seconds."`

	MinUploadMTU   uint16 `toml:"min_upload_mtu" comment:"Minimum upload MTU."`
	MaxUploadMTU   uint16 `toml:"max_upload_mtu" comment:"Maximum upload MTU."`
	MinDownloadMTU uint16 `toml:"min_download_mtu" comment:"Minimum download MTU."`
	MaxDownloadMTU uint16 `toml:"max_download_mtu" comment:"Maximum download MTU."`

	MTUParallelism uint8 `toml:"mtu_parallelism" comment:"Number of parallel MTU test workers."`
	RxTxWorkers    uint8 `toml:"rx_tx_workers" comment:"Number of RX/TX worker goroutines."`
}

// DefaultStormDNSConfig returns a StormDNS configuration with recommended
// defaults. Domain and EncryptionKey must be provided for a usable config.
func DefaultStormDNSConfig() StormDNSConfig {
	return StormDNSConfig{
		Domain:        "",
		EncryptionKey: "",
		DataEncMethod: EncXOR,
		ResolverPort:  53,
		DNSQueryType:  StormDNSQueryTXT,

		MTUTestTimeoutSec:      2.0,
		MTUTestRetries:         3,
		SessionInitRetryMaxSec: 60.0,

		MinUploadMTU:   100,
		MaxUploadMTU:   200,
		MinDownloadMTU: 1000,
		MaxDownloadMTU: 4000,

		MTUParallelism: 16,
		RxTxWorkers:    4,
	}
}

// Validate validates the StormDNS configuration.
func (c StormDNSConfig) Validate() map[string]error {
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

	switch strings.ToUpper(strings.TrimSpace(c.DNSQueryType)) {
	case "", StormDNSQueryTXT, StormDNSQueryNS, StormDNSQueryCNAME, StormDNSQueryRotate:
	default:
		errs["dns_query_type"] = fmt.Errorf("must be TXT, NS, CNAME or ROTATE")
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

// StormDNSService manages StormDNS configurations and tunnels.
type StormDNSService interface {
	SaveConfig(config StormDNSConfig, name string) error
	EditConfig(config StormDNSConfig, originalName string) error
	LoadConfig(name string) (StormDNSConfig, error)
	GetAllConfigFiles() ([]StormDNSConfigFile, error)
	ValidateAllConfigs() ([]ConfigValidationResult, error)
	RenameConfig(oldName, newName string) error
	RunTunnel(
		ctx context.Context,
		config StormDNSConfig,
		resolverIP string,
		listenPort uint16,
	) (io.Closer, error)
}

type stormDNSService struct {
	configs configStore[StormDNSConfig]
}

// StormDNSServiceOption configures a StormDNSService.
type StormDNSServiceOption func(*stormDNSService)

// WithStormDNSDir sets the directory used to store StormDNS configurations.
func WithStormDNSDir(dir string) StormDNSServiceOption {
	return func(service *stormDNSService) {
		if dir != "" {
			service.configs.dir = dir
		}
	}
}

// NewStormDNSService creates a StormDNS service.
func NewStormDNSService(options ...StormDNSServiceOption) StormDNSService {
	service := &stormDNSService{
		configs: newConfigStore[StormDNSConfig](tunnelConfigDir(stormdnsDir), "StormDNS"),
	}

	for _, option := range options {
		option(service)
	}

	stormtunnel.DiscardLogger()
	return service
}

func (s *stormDNSService) SaveConfig(config StormDNSConfig, name string) error {
	return s.configs.SaveConfig(config, name)
}

func (s *stormDNSService) EditConfig(config StormDNSConfig, originalName string) error {
	return s.configs.EditConfig(config, originalName)
}

func (s *stormDNSService) LoadConfig(name string) (StormDNSConfig, error) {
	return s.configs.LoadConfig(name)
}

func (s *stormDNSService) GetAllConfigFiles() ([]StormDNSConfigFile, error) {
	return s.configs.GetAllConfigFiles()
}

func (s *stormDNSService) ValidateAllConfigs() ([]ConfigValidationResult, error) {
	return s.configs.ValidateAllConfigs()
}

func (s *stormDNSService) RenameConfig(oldName, newName string) error {
	return s.configs.RenameConfig(oldName, newName)
}

// RunTunnel starts an in-process StormDNS client for the given resolver IP.
// The tunnel runs embedded: it exposes a local SOCKS5 listener on listenPort.
// The caller must Close the returned handle, otherwise goroutines and
// sockets leak.
func (s *stormDNSService) RunTunnel(
	ctx context.Context,
	config StormDNSConfig,
	resolverIP string,
	listenPort uint16,
) (io.Closer, error) {
	if errs := config.Validate(); len(errs) > 0 {
		return nil, fmt.Errorf("invalid StormDNS config: %v", errs)
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context canceled before tunnel setup: %w", err)
	}

	addr, err := parseTunnelEndpoint(resolverIP, listenPort)
	if err != nil {
		return nil, err
	}

	return startStormTunnel(ctx, config, addr, listenPort)
}

// startStormTunnel builds the internal client config from a
// StormDNSConfig and boots the embedded client runtime against
// resolverIP. listenPort controls the local SOCKS5 listener port.
func startStormTunnel(
	ctx context.Context,
	cfg StormDNSConfig,
	resolverIP netip.Addr,
	listenPort uint16,
) (*embeddedHandle[*stormtunnel.Client], error) {
	stormCfg := stormtunnel.DefaultClientConfig()

	stormCfg.ListenPort = int(listenPort)

	stormCfg.Domains = []string{cfg.Domain}
	stormCfg.DataEncryptionMethod = int(cfg.DataEncMethod)
	stormCfg.EncryptionKey = cfg.EncryptionKey

	stormCfg.Resolvers = []stormtunnel.ResolverAddress{{
		IP:   resolverIP.String(),
		Port: int(cfg.ResolverPort),
	}}

	// Embedded use always scans the given resolver; the vendor bootstrap
	// resolves these into the active MTU test parameters.
	stormCfg.StartupMode = "resolvers"
	stormCfg.MTUTestTimeoutResolvers = cfg.MTUTestTimeoutSec
	stormCfg.MTUTestRetriesResolvers = int(cfg.MTUTestRetries)

	stormCfg.SessionInitRetryMaxSeconds = cfg.SessionInitRetryMaxSec

	stormCfg.MinUploadMTU = int(cfg.MinUploadMTU)
	stormCfg.MaxUploadMTU = int(cfg.MaxUploadMTU)
	stormCfg.MinDownloadMTU = int(cfg.MinDownloadMTU)
	stormCfg.MaxDownloadMTU = int(cfg.MaxDownloadMTU)

	if cfg.MTUParallelism > 0 {
		stormCfg.MTUTestParallelismResolvers = int(cfg.MTUParallelism)
	}
	if cfg.RxTxWorkers > 0 {
		stormCfg.RX_TX_Workers = int(cfg.RxTxWorkers)
	}

	stormCfg.DNSQueryType = strings.ToUpper(strings.TrimSpace(cfg.DNSQueryType))
	if stormCfg.DNSQueryType == "" {
		stormCfg.DNSQueryType = StormDNSQueryTXT
	}

	// Embedded mode: no log files on disk, console already discarded.
	stormCfg.LogToFile = false
	stormCfg.AutoDisableTimeoutServers = true

	app, err := stormtunnel.Bootstrap(stormCfg, "")
	if err != nil {
		return nil, fmt.Errorf("bootstrap client: %w", err)
	}

	return bootEmbedded(ctx, app)
}
