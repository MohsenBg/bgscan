package config

import (
	"time"

	"bgscan/internal/logger"
)

// HTTPConfig defines configuration for HTTP probing and TLS validation.
type HTTPConfig struct {
	Workers             int        `toml:"workers" comment:"Concurrent HTTP workers. Range: 1-1000. Higher = faster but more CPU/network."`
	Host                string     `toml:"host" comment:"Target host or URL. Used as Host header and SNI fallback."`
	ServerName          string     `toml:"server_name" comment:"TLS SNI override. Empty = use host field. Must be a valid domain."`
	Port                int        `toml:"port" comment:"HTTP port. Range: 1-65535. Common: 80 (HTTP), 443 (HTTPS)."`
	Protocol            string     `toml:"protocol" comment:"Protocol: http or https."`
	Version             string     `toml:"version" comment:"HTTP version: h1 (HTTP/1.1 only), h2 (HTTP/2 only), h1,h2 (negotiate), h3 (QUIC)."`
	TLSValidation       bool       `toml:"tls_validation" comment:"Validate TLS certificates. false = accept self-signed/expired certs."`
	MinTLSVersion       string     `toml:"min_tls_version" comment:"Min TLS version: tls1.0, tls1.1, tls1.2, tls1.3. Must be <= max_tls_version."`
	MaxTLSVersion       string     `toml:"max_tls_version" comment:"Max TLS version: tls1.0, tls1.1, tls1.2, tls1.3. Must be >= min_tls_version."`
	Timeout             DurationMS `toml:"timeout" comment:"Max wait for HTTP response, in ms. Range: 100-60000."`
	OutputPrefix        string     `toml:"output_prefix" comment:"Filename prefix for result files."`
	AcceptedStatusCodes []int      `toml:"accepted_status_codes" comment:"HTTP status codes to accept. Empty = accept all codes."`
}

func DefaultHTTPConfig() HTTPConfig {
	platform := DetectPlatform()
	tier := SelectTier(CheckResources())

	tiers, ok := httpDefaults[platform]
	if !ok {
		logger.CoreWarn("bgscan: no HTTP defaults for platform %q, falling back to %q", platform, Desktop)
		platform = Desktop
		tiers = httpDefaults[platform]
	}

	cfg, ok := tiers[tier]
	if !ok {
		logger.CoreWarn("bgscan: no HTTP defaults for platform %q tier %q, falling back to %q", platform, tier, Mid)
		cfg = tiers[Mid]
	}

	return cfg
}

// httpBase holds fields shared across every platform/tier — only Workers
// and Timeout actually vary by platform/tier for HTTP probing.
func httpBase() HTTPConfig {
	return HTTPConfig{
		Host:                "example.com",
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
