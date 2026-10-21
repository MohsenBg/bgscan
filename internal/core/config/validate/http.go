package validate

import (
	"fmt"
	"strings"
	"time"

	"github.com/MohsenBg/bgscan/internal/core/config"
)

var allowedProtocols = []string{
	"http",
	"https",
}

var allowedTLSVersions = []string{
	"tls1.0",
	"tls1.1",
	"tls1.2",
	"tls1.3",
}

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

const (
	MinHTTPWorkers = 1
	MaxHTTPWorkers = 1000

	MinHTTPPort = 1
	MaxHTTPPort = 65535
)

const (
	MinHTTPTimeout = 100 * time.Millisecond
	MaxHTTPTimeout = 60 * time.Second
)

// ValidateHTTP strictly validates an HTTPConfig and returns errors by field name.
func ValidateHTTP(cfg config.HTTPConfig) map[Field]error {
	errs := map[Field]error{}

	if err := checkInt("Workers", cfg.Workers, MinHTTPWorkers, MaxHTTPWorkers); err != nil {
		errs[FieldWorkers] = err
	}

	if err := checkHost("Host", cfg.Host); err != nil {
		errs[FieldHost] = err
	}

	if err := checkSNI("ServerName", cfg.ServerName); err != nil {
		errs[FieldServerName] = err
	}

	if err := checkInt("Port", cfg.Port, MinHTTPPort, MaxHTTPPort); err != nil {
		errs[FieldPort] = err
	}

	if err := checkEnum("Protocol", cfg.Protocol, allowedProtocols); err != nil {
		errs[FieldProtocol] = err
	}

	if err := checkDuration("Timeout", cfg.Timeout.Duration(),
		MinHTTPTimeout, MaxHTTPTimeout); err != nil {
		errs[FieldTimeout] = err
	}

	if err := checkEnum("HTTP Version", cfg.Version, allowedHTTPVersions); err != nil {
		errs[FieldVersion] = err
	}

	if cfg.Fingerprint != "" && !config.ValidFingerprint(cfg.Fingerprint) {
		errs[FieldFingerprint] = fmt.Errorf("unknown fingerprint %q; valid: %s",
			cfg.Fingerprint, strings.Join(config.FingerprintLabels(), ", "))
	}

	hasTLSErr := false

	if err := checkEnum("MinTLSVersion", cfg.MinTLSVersion, allowedTLSVersions); err != nil {
		errs[FieldMinTLSVersion] = err
		hasTLSErr = true
	}

	if err := checkEnum("MaxTLSVersion", cfg.MaxTLSVersion, allowedTLSVersions); err != nil {
		errs[FieldMaxTLSVersion] = err
		hasTLSErr = true
	}

	if !hasTLSErr {
		if err := checkEnumOrder(
			"MinTLSVersion",
			"MaxTLSVersion",
			cfg.MinTLSVersion,
			cfg.MaxTLSVersion,
			allowedTLSVersions,
		); err != nil {
			errs[FieldMinTLSVersion] = err
			errs[FieldMaxTLSVersion] = err
		}
	}

	if err := checkPrefix("OutputPrefix", cfg.OutputPrefix); err != nil {
		errs[FieldOutputPrefix] = err
	}

	if err := checkStatusCodes("AcceptedStatusCodes", cfg.AcceptedStatusCodes); err != nil {
		errs[FieldAcceptedStatusCodes] = err
	}

	return errs
}

// NormalizeHTTP replaces invalid HTTPConfig fields with defaults and reports each correction.
func NormalizeHTTP(cfg *config.HTTPConfig) []Warning {
	def := config.DefaultHTTPConfig(nil)
	var warns []Warning

	fixInt(
		"Workers",
		&cfg.Workers,
		MinHTTPWorkers,
		MaxHTTPWorkers,
		def.Workers,
		&warns,
	)

	fixInt(
		"Port",
		&cfg.Port,
		MinHTTPPort,
		MaxHTTPPort,
		def.Port,
		&warns,
	)

	fixEnum(
		"Protocol",
		&cfg.Protocol,
		allowedProtocols,
		def.Protocol,
		&warns,
	)

	fixHost(
		"Host",
		&cfg.Host,
		def.Host,
		&warns,
	)

	fixHTTPStatusCodes(
		"AcceptedStatusCodes",
		&cfg.AcceptedStatusCodes,
		def.AcceptedStatusCodes,
		&warns,
	)

	fixSNI(
		"ServerName",
		&cfg.ServerName,
		def.ServerName,
		&warns,
	)

	fixEnum(
		"Version",
		&cfg.Version,
		allowedHTTPVersions,
		def.Version,
		&warns,
	)

	fixDurationMS(
		"Timeout",
		&cfg.Timeout,
		MinHTTPTimeout,
		MaxHTTPTimeout,
		def.Timeout,
		&warns,
	)

	fixEnum(
		"MinTLSVersion",
		&cfg.MinTLSVersion,
		allowedTLSVersions,
		def.MinTLSVersion,
		&warns,
	)

	fixEnum(
		"MaxTLSVersion",
		&cfg.MaxTLSVersion,
		allowedTLSVersions,
		def.MaxTLSVersion,
		&warns,
	)

	fixEnumOrder(
		"MinTLSVersion",
		"MaxTLSVersion",
		&cfg.MinTLSVersion,
		&cfg.MaxTLSVersion,
		def.MinTLSVersion,
		def.MaxTLSVersion,
		allowedTLSVersions,
		&warns,
	)

	fixString(
		"OutputPrefix",
		&cfg.OutputPrefix,
		def.OutputPrefix,
		&warns,
	)

	if cfg.Fingerprint != "" && !config.ValidFingerprint(cfg.Fingerprint) {
		warns = append(warns, Warning{
			Field:  "Fingerprint",
			OldVal: cfg.Fingerprint,
			NewVal: "",
			Reason: "unknown fingerprint → cleared",
		})
		cfg.Fingerprint = ""
	}

	return warns
}
