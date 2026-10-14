package xray

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/netip"
	"path/filepath"
	"time"

	"github.com/MohsenBg/bgscan/internal/core/fileutil"
	core "github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/shared"

	// Registers all proxies/transports; without these the core rejects configs
	// with "proxy not registered".
	_ "github.com/xtls/xray-core/main/distro/all"
	_ "github.com/xtls/xray-core/main/json"
)

// XrayService builds scan configs and runs them on the embedded Xray core.
type XrayService interface {
	Version() string
	GetOutboundTemplateByName(name string) (*OutboundFile, error)

	// GenerateConfig pairs an outbound template with a local SOCKS inbound.
	// A nonzero remotePort overrides the dial port (written to the protocol's
	// port field, or the host:port endpoint for WireGuard); pass 0 to keep the
	// template's own port. localPort is the SOCKS listener port.
	GenerateConfig(outbound string, ip netip.Addr, localPort, remotePort uint16) (*Config, error)
	ValidateConfig(ctx context.Context, config *Config) error
	Start(ctx context.Context, config *Config) (Instance, error)

	// CloseShared releases process-global transport state shared by all instances.
	CloseShared() error
}

type Instance interface {
	Close() error
}

type xrayService struct{}

type ServiceOption func()

// WithScanTuning installs scanner-friendly transport limits process-wide.
// It must be applied before any scan runs.
func WithScanTuning(t shared.ScanTuning) ServiceOption {
	return func() { shared.ApplyScanTuning(t) }
}

func NewXrayService(opts ...ServiceOption) XrayService {
	for _, opt := range opts {
		opt()
	}
	return &xrayService{}
}

func (s *xrayService) Version() string {
	return core.Version()
}

func (s *xrayService) GetOutboundTemplateByName(name string) (*OutboundFile, error) {
	return GetOutboundTemplateByName(name)
}

func (s *xrayService) GenerateConfig(outboundName string, ip netip.Addr, localPort, remotePort uint16) (*Config, error) {
	if !ip.IsValid() {
		return nil, fmt.Errorf("invalid IP: %s", ip)
	}

	template, err := s.GetOutboundTemplateByName(outboundName)
	if err != nil {
		return nil, err
	}

	outbound, err := applyOutboundTemplate(template.Path, ip, remotePort)
	if err != nil {
		return nil, err
	}

	raw, err := json.Marshal(outbound)
	if err != nil {
		return nil, fmt.Errorf("marshal outbound: %w", err)
	}

	return &Config{
		Log:       &LogConfig{LogLevel: LogLevelNone},
		Inbounds:  []Inbound{getInbound(localPort)},
		Outbounds: []json.RawMessage{raw},
	}, nil
}

func (s *xrayService) ValidateConfig(ctx context.Context, config *Config) error {
	if err := validateConfigShape(config); err != nil {
		return err
	}

	configBytes, err := toCoreJSON(config)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cfg, err := core.LoadConfig("json", bytes.NewReader(configBytes))
	if err != nil {
		return fmt.Errorf("parse xray config: %w", err)
	}

	instance, err := core.NewWithContext(ctx, cfg)
	if err != nil {
		return fmt.Errorf("xray core rejected config: %w", err)
	}
	_ = instance.Close()

	return nil
}

func (s *xrayService) Start(ctx context.Context, config *Config) (Instance, error) {
	if err := validateConfigShape(config); err != nil {
		return nil, err
	}

	configBytes, err := toCoreJSON(config)
	if err != nil {
		return nil, err
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	instance, err := core.StartInstance("json", configBytes)
	if err != nil {
		return nil, fmt.Errorf("start xray instance: %w", err)
	}

	return instance, nil
}

func (s *xrayService) CloseShared() error {
	return shared.CloseAll()
}

func validateConfigShape(config *Config) error {
	if config == nil {
		return errors.New("xray config is nil")
	}
	if len(config.Inbounds) == 0 {
		return errors.New("xray config has no inbounds")
	}
	if len(config.Outbounds) == 0 {
		return errors.New("xray config has no outbounds")
	}
	return nil
}

func toCoreJSON(config *Config) ([]byte, error) {
	configBytes, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("marshal config: %w", err)
	}
	return configBytes, nil
}

func getAssetsPath(parts ...string) string {
	base, err := fileutil.BasePath()
	if err != nil {
		return filepath.Join(parts...)
	}
	return filepath.Join(append([]string{base}, parts...)...)
}

func templateDir() string {
	return getAssetsPath("assets", "xray", "outbounds")
}
