package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/MohsenBg/bgscan/internal/logger"
)

type HTTPMode string

const (
	HTTPModeTarget HTTPMode = "target"
	HTTPModeHost   HTTPMode = "host"
)

// ParseHTTPMode converts a raw mode string into an HTTPMode.
// An empty value means the default mode (host), so config files written
// before the mode field existed keep their previous behavior.
func ParseHTTPMode(s string) (HTTPMode, error) {
	mode := HTTPMode(strings.ToLower(strings.TrimSpace(s)))

	switch mode {
	case "":
		return HTTPModeHost, nil
	case HTTPModeHost, HTTPModeTarget:
		return mode, nil
	default:
		return "", fmt.Errorf(
			"invalid HTTP mode %q: want %q or %q",
			s,
			HTTPModeHost,
			HTTPModeTarget,
		)
	}
}

// HTTPConfig defines configuration for HTTP probing and TLS validation.
type HTTPConfig struct {
	Workers             int        `toml:"workers" comment:"Concurrent HTTP workers. Range: 1-1000."`
	Mode                HTTPMode   `toml:"mode" comment:"Target mode: target or host."`
	Host                string     `toml:"host" comment:"Hostname used as HTTP Host and TLS SNI when applicable."`
	ServerName          string     `toml:"server_name" comment:"TLS SNI override. Empty = use Host when applicable."`
	Port                int        `toml:"port" comment:"Default HTTP port. Used when dynamic port is disabled or the target has no port."`
	DynamicPort         bool       `toml:"dynamic_port" comment:"Use the target's port when available; otherwise fall back to the configured Port."`
	Protocol            string     `toml:"protocol" comment:"Protocol: http or https."`
	Version             string     `toml:"version" comment:"HTTP version: h1, h2, h1,h2, h3."`
	Fingerprint         string     `toml:"fingerprint" comment:"uTLS fingerprint. Empty = standard crypto/tls."`
	TLSValidation       bool       `toml:"tls_validation" comment:"Validate TLS certificates."`
	MinTLSVersion       string     `toml:"min_tls_version" comment:"Minimum TLS version."`
	MaxTLSVersion       string     `toml:"max_tls_version" comment:"Maximum TLS version."`
	Timeout             DurationMS `toml:"timeout" comment:"Maximum HTTP response wait, in ms."`
	OutputPrefix        string     `toml:"output_prefix" comment:"Filename prefix for result files."`
	AcceptedStatusCodes []int      `toml:"accepted_status_codes" comment:"HTTP status codes to accept. Empty = accept all."`
}

func DefaultHTTPConfig(log *logger.Logger) HTTPConfig {
	platform := DetectPlatform()
	tier := SelectTier(CheckResources())

	tiers, ok := httpDefaults[platform]
	if !ok {
		log.Warn("bgscan: no HTTP defaults for platform %q, falling back to %q", platform, Desktop)
		platform = Desktop
		tiers = httpDefaults[platform]
	}

	cfg, ok := tiers[tier]
	if !ok {
		log.Warn("bgscan: no HTTP defaults for platform %q tier %q, falling back to %q", platform, tier, Mid)
		cfg = tiers[Mid]
	}

	return cfg
}

// httpBase holds fields shared across every platform/tier — only Workers
// and Timeout actually vary by platform/tier for HTTP probing.
func httpBase() HTTPConfig {
	return HTTPConfig{
		Host:                "www.cloudflare.com",
		Mode:                HTTPModeHost,
		Port:                443,
		Protocol:            "https",
		Version:             "h1,h2",
		TLSValidation:       true,
		MinTLSVersion:       "tls1.1",
		MaxTLSVersion:       "tls1.3",
		OutputPrefix:        "http_",
		AcceptedStatusCodes: []int{},
	}
}

var httpDefaults = map[Platform]map[Tier]HTTPConfig{
	Server: {
		Low:  withHTTP(httpBase(), 50, 4*time.Second),
		Mid:  withHTTP(httpBase(), 200, 3*time.Second),
		High: withHTTP(httpBase(), 500, 2*time.Second),
	},
	Desktop: {
		Low:  withHTTP(httpBase(), 20, 4*time.Second),
		Mid:  withHTTP(httpBase(), 50, 4*time.Second),
		High: withHTTP(httpBase(), 100, 3*time.Second),
	},
	Android: {
		Low:  withHTTP(httpBase(), 5, 5*time.Second),
		Mid:  withHTTP(httpBase(), 15, 5*time.Second),
		High: withHTTP(httpBase(), 30, 4*time.Second),
	},
}

// withHTTP returns a copy of base with Workers and Timeout overridden.
func withHTTP(base HTTPConfig, workers int, timeout time.Duration) HTTPConfig {
	base.Workers = workers
	base.Timeout = NewDurationMS(timeout)
	return base
}
