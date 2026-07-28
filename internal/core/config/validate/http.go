package validate

import (
	"time"

	"bgscan/internal/core/config"
)

var allowedProtocols = []string{"http", "https"}

var allowedTLSVersions = []string{"tls1.0", "tls1.1", "tls1.2", "tls1.3"}

var allowedHTTPVersions = []string{
	"h1",
	"http1",
	"http1.1",

	"h2",
	"http2",

	"h1,h2",
	"http1,http2",
	"http1.1,http2",
	"http2,http1",
	"http2,http1.1",

	"h3",
	"http3",
}

// ValidateHTTP strictly validates an HTTPConfig and returns errors by field name.
func ValidateHTTP(cfg config.HTTPConfig) map[string]error {
	errs := map[string]error{}

	if err := checkInt("Workers", cfg.Workers, 1, 5000); err != nil {
		errs["Workers"] = err
	}

	if err := checkHost("Host", cfg.Host); err != nil {
		errs["Host"] = err
	}

	if err := checkSNI("ServerName", cfg.ServerName); err != nil {
		errs["ServerName"] = err
	}

	if err := checkInt("Port", cfg.Port, 1, 65535); err != nil {
		errs["Port"] = err
	}

	if err := checkEnum("Protocol", cfg.Protocol, allowedProtocols); err != nil {
		errs["Protocol"] = err
	}

	if err := checkDuration("Timeout", cfg.Timeout.Duration(),
		100*time.Millisecond, 60*time.Second); err != nil {
		errs["Timeout"] = err
	}

	if err := checkEnum("HTTP Version", cfg.Version, allowedHTTPVersions); err != nil {
		errs["Version"] = err
	}

	hasTLSErr := false
	if err := checkEnum("MinTLSVersion", cfg.MinTLSVersion, allowedTLSVersions); err != nil {
		errs["MinTLSVersion"] = err
		hasTLSErr = true
	}

	if err := checkEnum("MaxTLSVersion", cfg.MaxTLSVersion, allowedTLSVersions); err != nil {
		errs["MaxTLSVersion"] = err
		hasTLSErr = true
	}

	if !hasTLSErr {
		if err := checkEnumOrder("MinTLSVersion", "MaxTLSVersion", cfg.MinTLSVersion, cfg.MaxTLSVersion, allowedTLSVersions); err != nil {
			errs["MinTLSVersion"] = err
			errs["MaxTLSVersion"] = err
		}
	}

	if err := checkPrefix("PrefixOutput", cfg.OutputPrefix); err != nil {
		errs["PrefixOutput"] = err
	}

	if err := checkStatusCodes("AcceptedStatusCodes", cfg.AcceptedStatusCodes); err != nil {
		errs["AcceptedStatusCodes"] = err
	}

	return errs
}

// NormalizeHTTP replaces invalid HTTPConfig fields with defaults and reports each correction.
func NormalizeHTTP(cfg *config.HTTPConfig) []Warning {
	def := config.DefaultHTTPConfig()
	var warns []Warning

	fixInt("Workers", &cfg.Workers, 1, 5000, def.Workers, &warns)
	fixInt("Port", &cfg.Port, 1, 65535, def.Port, &warns)
	fixEnum("Protocol", &cfg.Protocol, allowedProtocols, def.Protocol, &warns)
	fixHost("Host", &cfg.Host, def.Host, &warns)
	fixHTTPStatusCodes("AcceptedStatusCodes", &cfg.AcceptedStatusCodes, def.AcceptedStatusCodes, &warns)

	fixSNI("ServerName", &cfg.ServerName, def.ServerName, &warns)
	fixEnum("Version", &cfg.Version, allowedHTTPVersions, def.Version, &warns)
	fixDurationMS("Timeout", &cfg.Timeout,
		100*time.Millisecond, 60*time.Second, def.Timeout, &warns)

	fixEnum("MinTLSVersion", &cfg.MinTLSVersion, allowedTLSVersions, def.MinTLSVersion, &warns)
	fixEnum("MaxTLSVersion", &cfg.MaxTLSVersion, allowedTLSVersions, def.MaxTLSVersion, &warns)
	fixEnumOrder(
		"MinTLSVersion", "MaxTLSVersion",
		&cfg.MinTLSVersion, &cfg.MaxTLSVersion,
		def.MinTLSVersion, def.MaxTLSVersion,
		allowedTLSVersions,
		&warns,
	)
	fixString("PrefixOutput", &cfg.OutputPrefix, def.OutputPrefix, &warns)

	return warns
}
