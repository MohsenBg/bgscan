package xray

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/netip"
	"path/filepath"
	"time"

	"github.com/MohsenBg/bgscan/internal/core/fileutil"
	core "github.com/xtls/xray-core/core"

	// Registers all proxies/transports, otherwise the core rejects
	// configs with "proxy not registered".
	_ "github.com/xtls/xray-core/main/distro/all"
	// Registers the JSON config loader (core only ships protobuf).
	_ "github.com/xtls/xray-core/main/json"
)

// XrayService builds scan configs and runs them on the embedded Xray core,
// all in-process. Callers must Close started instances.
type XrayService interface {
	// Version reports the embedded core version.
	Version() string

	// GetOutboundTemplateByName finds a saved outbound template (".json" optional).
	GetOutboundTemplateByName(string) (*XrayOutboundsFile, error)

	// GenerateConfig pairs an outbound template with a local SOCKS inbound.
	GenerateConfig(outbound string, ip netip.Addr, port uint16) (*XrayConfig, error)

	// ValidateConfig builds (without starting) an instance to check the config.
	ValidateConfig(context.Context, *XrayConfig) error

	// Start launches a live instance. The caller owns it and must Close it.
	Start(context.Context, *XrayConfig) (*core.Instance, error)
}

// xrayService is the default XrayService implementation.
type xrayService struct{}

// NewXrayService creates an XrayService wired to the embedded core.
func NewXrayService() XrayService {
	return &xrayService{}
}

// Version reports the embedded core version.
func (s *xrayService) Version() string {
	return core.Version()
}

// GetOutboundTemplateByName finds a saved outbound template by name.
func (s *xrayService) GetOutboundTemplateByName(name string) (*XrayOutboundsFile, error) {
	return GetOutboundTemplateByName(name)
}

// GenerateConfig fills the named outbound template with the target IP and
// pairs it with a localhost SOCKS inbound.
func (s *xrayService) GenerateConfig(outboundName string, ip netip.Addr, port uint16) (*XrayConfig, error) {
	if !ip.IsValid() {
		return nil, fmt.Errorf("invalid IP: %s", ip)
	}

	template, err := GetOutboundTemplateByName(outboundName)
	if err != nil {
		return nil, err
	}

	outbound, err := applyOutboundTemplate(template.Path, ip)
	if err != nil {
		return nil, err
	}

	return &XrayConfig{
		Inbound:  []Inbound{getInbound(port)},
		Outbound: []any{outbound},
		Log:      &xrayLogConfig{Loglevel: "none"},
	}, nil
}

// toCoreJSON marshals the config to the JSON bytes the core loader reads.
func toCoreJSON(config *XrayConfig) ([]byte, error) {
	configBytes, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("marshal config: %w", err)
	}

	return configBytes, nil
}

// ValidateConfig checks the config by building (not starting) a core instance.
func (s *xrayService) ValidateConfig(ctx context.Context, config *XrayConfig) error {
	if config == nil {
		return fmt.Errorf("xray config is nil")
	}
	if len(config.Inbound) == 0 {
		return fmt.Errorf("xray config has no inbounds")
	}
	if len(config.Outbound) == 0 {
		return fmt.Errorf("xray config has no outbounds")
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
	defer instance.Close()

	return nil
}

// Start launches a live instance. Close it when done or it leaks.
func (s *xrayService) Start(ctx context.Context, config *XrayConfig) (*core.Instance, error) {
	if config == nil {
		return nil, fmt.Errorf("xray config is nil")
	}
	if len(config.Inbound) == 0 {
		return nil, fmt.Errorf("xray config has no inbounds")
	}
	if len(config.Outbound) == 0 {
		return nil, fmt.Errorf("xray config has no outbounds")
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

// getAssetsPath joins parts onto the app base directory.
func getAssetsPath(parts ...string) string {
	base, err := fileutil.BasePath()
	if err != nil {
		return filepath.Join(parts...)
	}

	return filepath.Join(append([]string{base}, parts...)...)
}

// templateDir is the on-disk folder of saved outbound templates.
func templateDir() string {
	return getAssetsPath("assets", "xray", "outbounds")
}
