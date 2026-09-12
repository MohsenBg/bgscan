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

type TheFeedConfigFile = ConfigFile[TheFeedConfig]

type TheFeedConfig struct {
	Domain       string       `toml:"domain" comment:"TheFeed server domain (the DNS subdomain that receives feed queries)."`
	Passphrase   string       `toml:"passphrase" comment:"Encryption passphrase. Used to derive the AES query/response keys."`
	ResolverPort uint16       `toml:"resolver_port" comment:"Resolver port. Standard DNS uses 53; thefeed servers often use 5300."`
	ResolverType ResolverType `toml:"resolver_type" comment:"Resolver transport: udp, tcp, or dot (DNS over TLS)."`
	QueryMode    string       `toml:"query_mode" comment:"DNS query encoding: single (base32 label) or double (multi-label hex)."`
}

func DefaultTheFeedConfig() TheFeedConfig {
	return TheFeedConfig{
		Domain:       "",
		Passphrase:   "",
		ResolverPort: 53,
		ResolverType: ResolverTypeUDP,
		QueryMode:    "single",
	}
}

func (c TheFeedConfig) Validate() map[string]error {
	errs := make(map[string]error)

	if err := netutil.ValidateDomain(strings.TrimSpace(c.Domain)); err != nil {
		errs["domain"] = err
	}

	if c.Passphrase == "" {
		errs["passphrase"] = fmt.Errorf("passphrase is required")
	}

	if c.ResolverPort == 0 {
		errs["resolver_port"] = fmt.Errorf("must be between 1 and 65535")
	}

	if c.QueryMode != "single" && c.QueryMode != "double" {
		errs["query_mode"] = fmt.Errorf("must be single or double")
	}

	if !c.ResolverType.IsValid() {
		errs["resolver_type"] = fmt.Errorf("invalid resolver type")
	}

	return errs
}

type TheFeedService interface {
	SaveConfig(config TheFeedConfig, name string) error
	EditConfig(config TheFeedConfig, originalName string) error
	LoadConfig(name string) (TheFeedConfig, error)
	GetAllConfigFiles() ([]ConfigFile[TheFeedConfig], error)
	ValidateAllConfigs() ([]ConfigValidationResult, error)
	RenameConfig(oldName, newName string) error
	RunTunnel(ctx context.Context, config TheFeedConfig, resolverAddr netip.Addr, timeout time.Duration) error
}

type thefeedService struct {
	resolver Resolver
	configs  configStore[TheFeedConfig]
}

type TheFeedServiceOption func(*thefeedService)

func WithTheFeedDir(dir string) TheFeedServiceOption {
	return func(s *thefeedService) {
		if dir != "" {
			s.configs.dir = dir
		}
	}
}

func WithTheFeedResolver(resolver Resolver) TheFeedServiceOption {
	return func(s *thefeedService) {
		if resolver != nil {
			s.resolver = resolver
		}
	}
}

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

func (s *thefeedService) SaveConfig(config TheFeedConfig, name string) error {
	return s.configs.SaveConfig(config, name)
}

func (s *thefeedService) EditConfig(config TheFeedConfig, originalName string) error {
	return s.configs.EditConfig(config, originalName)
}

func (s *thefeedService) LoadConfig(name string) (TheFeedConfig, error) {
	return s.configs.LoadConfig(name)
}

func (s *thefeedService) GetAllConfigFiles() ([]ConfigFile[TheFeedConfig], error) {
	return s.configs.GetAllConfigFiles()
}

func (s *thefeedService) ValidateAllConfigs() ([]ConfigValidationResult, error) {
	return s.configs.ValidateAllConfigs()
}

func (s *thefeedService) RenameConfig(oldName, newName string) error {
	return s.configs.RenameConfig(oldName, newName)
}

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
	if decodeErr == nil {
		return fmt.Errorf("response decode: %w", decodeErr)
	}

	return nil
}
