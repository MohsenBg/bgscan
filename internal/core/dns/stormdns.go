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
	Domain        string
	DataEncMethod EncMethod
	EncryptionKey string
	ResolverPort  int

	// DNSQueryType selects the tunnel query shape: TXT, NS, CNAME or
	// ROTATE. Empty means TXT. NS/CNAME/ROTATE require a server running
	// the same feature version.
	DNSQueryType string

	MTUTestTimeoutSec      float64
	MTUTestRetries         int
	SessionInitRetryMaxSec float64

	MinUploadMTU   int
	MaxUploadMTU   int
	MinDownloadMTU int
	MaxDownloadMTU int

	MTUParallelism int
	RxTxWorkers    int
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

	if strings.TrimSpace(c.EncryptionKey) == "" {
		errs["encryption_key"] = fmt.Errorf("encryption key is required")
	}

	if c.DataEncMethod < 0 || c.DataEncMethod > 5 {
		errs["enc_method"] = fmt.Errorf("must be between 0 and 5")
	}

	if c.ResolverPort == 0 {
		errs["resolver_port"] = fmt.Errorf("must be greater than zero")
	}

	switch strings.ToUpper(strings.TrimSpace(c.DNSQueryType)) {
	case "", StormDNSQueryTXT, StormDNSQueryNS, StormDNSQueryCNAME, StormDNSQueryRotate:
	default:
		errs["dns_query_type"] = fmt.Errorf("must be TXT, NS, CNAME or ROTATE")
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
		Port: cfg.ResolverPort,
	}}

	// Embedded use always scans the given resolver; the vendor bootstrap
	// resolves these into the active MTU test parameters.
	stormCfg.StartupMode = "resolvers"
	stormCfg.MTUTestTimeoutResolvers = cfg.MTUTestTimeoutSec
	stormCfg.MTUTestRetriesResolvers = cfg.MTUTestRetries

	stormCfg.SessionInitRetryMaxSeconds = cfg.SessionInitRetryMaxSec

	stormCfg.MinUploadMTU = cfg.MinUploadMTU
	stormCfg.MaxUploadMTU = cfg.MaxUploadMTU
	stormCfg.MinDownloadMTU = cfg.MinDownloadMTU
	stormCfg.MaxDownloadMTU = cfg.MaxDownloadMTU

	// Only override vendor worker defaults when explicitly set, so a
	// zero value never clobbers the sane upstream defaults.
	if cfg.MTUParallelism > 0 {
		stormCfg.MTUTestParallelismResolvers = cfg.MTUParallelism
	}
	if cfg.RxTxWorkers > 0 {
		stormCfg.RX_TX_Workers = cfg.RxTxWorkers
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
