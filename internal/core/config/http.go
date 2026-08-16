package config

import (
	"time"

	"bgscan/internal/logger"
)

// HTTPConfig defines configuration for HTTP probing and TLS validation.
type HTTPConfig struct {
	Workers             int        `toml:"workers" comment:"Number of concurrent HTTP requests."`
	Host                string     `toml:"host" comment:"HTTP Host header sent with each request."`
	ServerName          string     `toml:"server_name" comment:"TLS server name used for certificate validation."`
	Port                int        `toml:"port" comment:"HTTP port to connect to."`
	Protocol            string     `toml:"protocol" comment:"HTTP protocol to use, such as http or https."`
	Version             string     `toml:"version" comment:"HTTP protocol versions to allow, such as h1 or h2."`
	TLSValidation       bool       `toml:"tls_validation" comment:"Whether to validate the TLS certificate."`
	MinTLSVersion       string     `toml:"min_tls_version" comment:"Minimum allowed TLS version."`
	MaxTLSVersion       string     `toml:"max_tls_version" comment:"Maximum allowed TLS version."`
	Timeout             DurationMS `toml:"timeout" comment:"Maximum time to wait for an HTTP request, in milliseconds."`
	OutputPrefix        string     `toml:"output_prefix" comment:"Filename prefix for HTTP scan results."`
	AcceptedStatusCodes []int      `toml:"accepted_status_codes" comment:"HTTP status codes considered successful."`
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
