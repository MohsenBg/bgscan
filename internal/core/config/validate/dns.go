package validate

import (
	"math"
	"time"

	"github.com/MohsenBg/bgscan/internal/core/config"
)

var allowedDNSTransport = []string{
	"udp",
	"tcp",
	"dot",
	"doh",
}

const (
	// Resolver limits
	MinDNSResolverWorkers = 1
	MaxDNSResolverWorkers = 2500

	MinDNSResolverPort = 1
	MaxDNSResolverPort = math.MaxUint16

	MinDNSResolverTries = 1
	MaxDNSResolverTries = 10

	MinDNSResolverDPITries = 1
	MaxDNSResolverDPITries = 10
)

const (
	MinDNSResolverTimeout = 100 * time.Millisecond
	MaxDNSResolverTimeout = 30 * time.Second

	MinDNSResolverDPITimeout = 100 * time.Millisecond
	MaxDNSResolverDPITimeout = 10 * time.Second
)

const (
	// DNSTunneling limits
	MinDNSTunnelingWorkers = 1
	MaxDNSTunnelingWorkers = 500

	MinDNSTunnelingTimeout = 100 * time.Millisecond
	MaxDNSTunnelingTimeout = 60 * time.Second
)

// ValidateDNS strictly validates a DNSConfig and returns errors by field name.
// Nested field names use dot notation, such as "Resolver.Workers".
func ValidateDNS(cfg config.DNSConfig) map[string]error {
	errs := map[string]error{}

	for k, v := range validateResolver(cfg.Resolver) {
		errs["Resolver."+k] = v
	}

	for k, v := range validateDNSTunneling(cfg.DNSTunneling) {
		errs["DNSTunneling."+k] = v
	}

	return errs
}

func validateResolver(r config.ResolverConfig) map[string]error {
	errs := map[string]error{}

	if err := checkInt(
		"Workers",
		r.Workers,
		MinDNSResolverWorkers,
		MaxDNSResolverWorkers,
	); err != nil {
		errs["Workers"] = err
	}

	if err := checkEnum("Protocol", r.Transport, allowedDNSTransport); err != nil {
		errs["Transport"] = err
	}

	if err := checkDomain("Domain", r.Domain); err != nil {
		errs["Domain"] = err
	}

	if err := checkUint16(
		"Port",
		r.Port,
		MinDNSResolverPort,
		MaxDNSResolverPort,
	); err != nil {
		errs["Port"] = err
	}

	if err := checkStringSlice("CheckTypes", r.CheckTypes); err != nil {
		errs["CheckTypes"] = err
	}

	if err := checkDuration(
		"Timeout",
		r.Timeout.Duration(),
		MinDNSResolverTimeout,
		MaxDNSResolverTimeout,
	); err != nil {
		errs["Timeout"] = err
	}

	if err := checkInt(
		"Tries",
		r.Tries,
		MinDNSResolverTries,
		MaxDNSResolverTries,
	); err != nil {
		errs["Tries"] = err
	}

	if err := checkInt(
		"DPITries",
		r.DPI.Tries,
		MinDNSResolverDPITries,
		MaxDNSResolverDPITries,
	); err != nil {
		errs["DPI.Tries"] = err
	}

	if err := checkDuration(
		"DPITimeout",
		r.DPI.Timeout.Duration(),
		MinDNSResolverDPITimeout,
		MaxDNSResolverDPITimeout,
	); err != nil {
		errs["DPI.Timeout"] = err
	}

	if err := checkPrefix("OutputPrefix", r.OutputPrefix); err != nil {
		errs["OutputPrefix"] = err
	}

	return errs
}

func validateDNSTunneling(d config.DNSTunneling) map[string]error {
	errs := map[string]error{}

	if err := checkInt(
		"Workers",
		d.Workers,
		MinDNSTunnelingWorkers,
		MaxDNSTunnelingWorkers,
	); err != nil {
		errs["Workers"] = err
	}

	if err := checkDuration(
		"Timeout",
		d.Timeout.Duration(),
		MinDNSTunnelingTimeout,
		MaxDNSTunnelingTimeout,
	); err != nil {
		errs["Timeout"] = err
	}

	if err := checkInt(
		"Tries",
		d.Tries,
		MinDNSResolverTries,
		MaxDNSResolverTries,
	); err != nil {
		errs["Tries"] = err
	}

	if err := checkPrefix("OutputPrefix", d.OutputPrefix); err != nil {
		errs["OutputPrefix"] = err
	}

	return errs
}

// NormalizeDNS replaces invalid DNSConfig fields with defaults and reports each correction.
func NormalizeDNS(cfg *config.DNSConfig) []Warning {
	var warns []Warning

	warns = append(warns, normalizeResolver(&cfg.Resolver)...)
	warns = append(warns, normalizeDNSTunneling(&cfg.DNSTunneling)...)

	return warns
}

func normalizeResolver(r *config.ResolverConfig) []Warning {
	def := config.DefaultDNSConfig().Resolver
	var warns []Warning

	fixInt(
		"Resolver.Workers",
		&r.Workers,
		MinDNSResolverWorkers,
		MaxDNSResolverWorkers,
		def.Workers,
		&warns,
	)

	fixEnum(
		"Resolver.Transport",
		&r.Transport,
		allowedDNSTransport,
		def.Transport,
		&warns,
	)

	fixDomain(
		"Resolver.Domain",
		&r.Domain,
		def.Domain,
		&warns,
	)

	fixUint16(
		"Resolver.Port",
		&r.Port,
		MinDNSResolverPort,
		MaxDNSResolverPort,
		def.Port,
		&warns,
	)

	fixStringSlice(
		"Resolver.CheckTypes",
		&r.CheckTypes,
		def.CheckTypes,
		&warns,
	)

	fixDurationMS(
		"Resolver.Timeout",
		&r.Timeout,
		MinDNSResolverTimeout,
		MaxDNSResolverTimeout,
		def.Timeout,
		&warns,
	)

	fixInt(
		"Resolver.Tries",
		&r.Tries,
		MinDNSResolverTries,
		MaxDNSResolverTries,
		def.Tries,
		&warns,
	)

	fixInt(
		"Resolver.DPI.Tries",
		&r.DPI.Tries,
		MinDNSResolverDPITries,
		MaxDNSResolverDPITries,
		def.DPI.Tries,
		&warns,
	)

	fixDurationMS(
		"Resolver.DPI.Timeout",
		&r.DPI.Timeout,
		MinDNSResolverDPITimeout,
		MaxDNSResolverDPITimeout,
		def.DPI.Timeout,
		&warns,
	)

	fixPrefix(
		"Resolver.OutputPrefix",
		&r.OutputPrefix,
		def.OutputPrefix,
		&warns,
	)

	return warns
}

func normalizeDNSTunneling(d *config.DNSTunneling) []Warning {
	def := config.DefaultDNSConfig().DNSTunneling
	var warns []Warning

	fixInt(
		"DNSTunneling.Workers",
		&d.Workers,
		MinDNSTunnelingWorkers,
		MaxDNSTunnelingWorkers,
		def.Workers,
		&warns,
	)

	fixInt(
		"DNSTunneling.Tries",
		&d.Tries,
		MinDNSResolverTries,
		MaxDNSResolverTries,
		def.Tries,
		&warns,
	)

	fixDurationMS(
		"DNSTunneling.Timeout",
		&d.Timeout,
		MinDNSTunnelingTimeout,
		MaxDNSTunnelingTimeout,
		def.Timeout,
		&warns,
	)

	fixPrefix(
		"DNSTunneling.OutputPrefix",
		&d.OutputPrefix,
		def.OutputPrefix,
		&warns,
	)

	return warns
}
