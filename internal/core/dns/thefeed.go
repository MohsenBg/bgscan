package dns

import (
	"context"
	"fmt"
	"net/netip"
	"strings"
	"time"

	"github.com/miekg/dns"
	"github.com/sartoopjj/thefeed/pkg/protocol"

	"github.com/MohsenBg/bgscan/internal/core/netutil"
)

const thefeedDir = "thefeed"

// TheFeedConfigFile pairs a TheFeed configuration with its on-disk
// identity.
type TheFeedConfigFile = ConfigFile[TheFeedConfig]

// TheFeedConfig contains the configuration required by the TheFeed
// client.
type TheFeedConfig struct {
	Domain       string       `toml:"domain" comment:"TheFeed server domain (the DNS subdomain that receives feed queries)."`
	Passphrase   string       `toml:"passphrase" comment:"Encryption passphrase. Used to derive the AES query/response keys."`
	ResolverPort uint16       `toml:"resolver_port" comment:"Resolver port. Standard DNS uses 53; thefeed servers often use 5300."`
	ResolverType ResolverType `toml:"resolver_type" comment:"Resolver transport: udp, tcp, or dot (DNS over TLS)."`
	QueryMode    string       `toml:"query_mode" comment:"DNS query encoding: single (base32 label) or double (multi-label hex)."`
}

// DefaultTheFeedConfig returns the default TheFeed configuration.
// Domain and Passphrase are deployment-specific and stay empty.
func DefaultTheFeedConfig() TheFeedConfig {
	return TheFeedConfig{
		Domain:       "",
		Passphrase:   "",
		ResolverPort: 53,
		ResolverType: ResolverTypeUDP,
		QueryMode:    "single",
	}
}

// Validate returns all validation errors for c, keyed by
// configuration field. The map is empty when the config is valid.
func (c TheFeedConfig) Validate() map[Field]error {
	errs := make(map[Field]error)

	if err := netutil.ValidateDomain(strings.TrimSpace(c.Domain)); err != nil {
		errs[FieldDomain] = err
	}

	if c.Passphrase == "" {
		errs[FieldPassphrase] = fmt.Errorf("passphrase is required")
	}

	if c.ResolverPort == 0 {
		errs[FieldResolverPort] = fmt.Errorf("must be between 1 and 65535")
	}

	if c.QueryMode != "single" && c.QueryMode != "double" {
		errs[FieldQueryMode] = fmt.Errorf("must be single or double")
	}

	if !c.ResolverType.IsValid() {
		errs[FieldResolverType] = fmt.Errorf("invalid resolver type")
	}

	return errs
}

// TheFeedService manages TheFeed configurations and tunnel probes.
type TheFeedService interface {
	// SaveConfig validates config and writes it as a new file named
	// name. It returns an error when the name is already in use or
	// the config is invalid.
	SaveConfig(config TheFeedConfig, name string) error
	// EditConfig validates config and replaces the stored config
	// named originalName. It returns an error when that config does
	// not exist or the config is invalid.
	EditConfig(config TheFeedConfig, originalName string) error
	// LoadConfig reads and returns the stored config named name.
	LoadConfig(name string) (TheFeedConfig, error)
	// GetAllConfigFiles returns every stored config that loads and
	// validates. Files that fail either check are skipped.
	GetAllConfigFiles() ([]ConfigFile[TheFeedConfig], error)
	// ValidateAllConfigs returns one result for each stored file
	// that fails to load or validate. Valid files are not included.
	ValidateAllConfigs() ([]ConfigValidationResult, error)
	// RenameConfig renames the stored config oldName to newName.
	// It returns an error when oldName does not exist or newName is
	// already in use.
	RenameConfig(oldName, newName string) error
	// RunTunnel tests the tunnel by sending one encoded query to
	// the resolver at resolverAddr and decoding the reply. A
	// timeout at or below zero uses 15 seconds.
	RunTunnel(ctx context.Context, config TheFeedConfig, resolverAddr netip.Addr, timeout time.Duration) error
	// Version returns the version of the TheFeed protocol library.
	Version() string
}

// thefeedService is the default TheFeedService implementation.
type thefeedService struct {
	resolver Resolver
	configs  configStore[TheFeedConfig]
}

// TheFeedServiceOption configures a TheFeedService.
type TheFeedServiceOption func(*thefeedService)

// WithTheFeedDir sets the directory used to store TheFeed
// configurations. An empty value is ignored.
func WithTheFeedDir(dir string) TheFeedServiceOption {
	return func(s *thefeedService) {
		if dir != "" {
			s.configs.dir = dir
		}
	}
}

// WithTheFeedResolver sets the Resolver used for tunnel queries.
// A nil value is ignored.
func WithTheFeedResolver(resolver Resolver) TheFeedServiceOption {
	return func(s *thefeedService) {
		if resolver != nil {
			s.resolver = resolver
		}
	}
}

// NewTheFeedService returns a TheFeedService that stores configs in
// the default TheFeed directory unless an option changes it.
func NewTheFeedService(options ...TheFeedServiceOption) TheFeedService {
	service := &thefeedService{
		configs: newConfigStore[TheFeedConfig](tunnelConfigDir(thefeedDir), "TheFeed"),
	}
	for _, opt := range options {
		opt(service)
	}
	if service.resolver == nil {
		service.resolver = NewResolver()
	}
	return service
}

// SaveConfig saves the config through the shared config store.
func (s *thefeedService) SaveConfig(config TheFeedConfig, name string) error {
	return s.configs.SaveConfig(config, name)
}

// EditConfig replaces the stored config through the shared config store.
func (s *thefeedService) EditConfig(config TheFeedConfig, originalName string) error {
	return s.configs.EditConfig(config, originalName)
}

// LoadConfig loads the config through the shared config store.
func (s *thefeedService) LoadConfig(name string) (TheFeedConfig, error) {
	return s.configs.LoadConfig(name)
}

// GetAllConfigFiles lists stored configs through the shared config store.
func (s *thefeedService) GetAllConfigFiles() ([]ConfigFile[TheFeedConfig], error) {
	return s.configs.GetAllConfigFiles()
}

// ValidateAllConfigs validates stored configs through the shared config store.
func (s *thefeedService) ValidateAllConfigs() ([]ConfigValidationResult, error) {
	return s.configs.ValidateAllConfigs()
}

// RenameConfig renames the config through the shared config store.
func (s *thefeedService) RenameConfig(oldName, newName string) error {
	return s.configs.RenameConfig(oldName, newName)
}

// RunTunnel sends one encoded TheFeed query to the resolver and
// checks that the reply decodes with the response key.
func (s *thefeedService) RunTunnel(
	ctx context.Context,
	config TheFeedConfig,
	resolverAddr netip.Addr,
	timeout time.Duration,
) error {
	if errs := config.Validate(); len(errs) > 0 {
		return fmt.Errorf("invalid config: %v", errs)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if timeout <= 0 {
		timeout = 15 * time.Second
	}

	queryKey, responseKey, err := protocol.DeriveKeys(config.Passphrase)
	if err != nil {
		return fmt.Errorf("derive keys: %w", err)
	}

	var queryMode int
	if config.QueryMode == "double" {
		queryMode = 1
	}

	qname, err := protocol.EncodeQuerySimple(queryKey, uint16(0), 0, config.Domain, queryMode)
	if err != nil {
		return fmt.Errorf("encode query: %w", err)
	}

	resp, err := s.resolver.Query(ctx, Query{
		Resolver:         resolverAddr.String(),
		Port:             config.ResolverPort,
		Domain:           qname,
		Transport:        ResolverTypeUDP,
		RecordType:       TypeTXT,
		EDNSBufSize:      4096,
		RecursionDesired: true,
		Timeout:          timeout,
	})

	if err != nil || resp == nil {
		return fmt.Errorf("dns exchange: %w", err)
	}

	var txtConcat strings.Builder
	for _, ans := range resp.Answer {
		if txt, ok := ans.(*dns.TXT); ok {
			for _, t := range txt.Txt {
				txtConcat.WriteString(t)
			}
		}
	}

	if txtConcat.Len() == 0 {
		return fmt.Errorf("no TXT records received")
	}

	_, decodeErr := protocol.DecodeResponse(responseKey, txtConcat.String())
	if decodeErr != nil {
		return fmt.Errorf("response decode: %w", decodeErr)
	}

	return nil
}

// Version returns the version reported by the TheFeed protocol
// library.
func (s *thefeedService) Version() string {
	return protocol.Version()
}
