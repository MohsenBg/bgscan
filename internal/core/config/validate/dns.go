package validate

import (
	"math"
	"time"

	"bgscan/internal/core/config"
)

var allowedDNSProtocols = []string{
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
	// DNSTT limits
	MinDNSTTWorkers = 1
	MaxDNSTTWorkers = 500

	MinDNSTTTimeout = 100 * time.Millisecond
	MaxDNSTTTimeout = 60 * time.Second
)

const (
	// SlipStream limits
	MinSlipStreamWorkers = 1
	MaxSlipStreamWorkers = 500

	MinSlipStreamTimeout = 100 * time.Millisecond
	MaxSlipStreamTimeout = 60 * time.Second
)

// ValidateDNS strictly validates a DNSConfig and returns errors by field name.
// Nested field names use dot notation, such as "Resolver.Workers".
func ValidateDNS(cfg config.DNSConfig) map[string]error {
	errs := map[string]error{}

	for k, v := range validateResolver(cfg.Resolver) {
		errs["Resolver."+k] = v
	}

	for k, v := range validateDNSTT(cfg.DNSTT) {
		errs["DNSTT."+k] = v
	}

	for k, v := range validateSlipStream(cfg.SlipStream) {
		errs["SlipStream."+k] = v
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

	if err := checkEnum("Protocol", r.Protocol, allowedDNSProtocols); err != nil {
		errs["Protocol"] = err
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
		r.DPITries,
		MinDNSResolverDPITries,
		MaxDNSResolverDPITries,
	); err != nil {
		errs["DPITries"] = err
	}

	if err := checkDuration(
		"DPITimeout",
		r.DPITimeout.Duration(),
		MinDNSResolverDPITimeout,
		MaxDNSResolverDPITimeout,
	); err != nil {
		errs["DPITimeout"] = err
	}

	if err := checkPrefix("PrefixOutput", r.PrefixOutput); err != nil {
		errs["PrefixOutput"] = err
	}

	return errs
}

func validateDNSTT(d config.DNSTTConfig) map[string]error {
	errs := map[string]error{}

	if err := checkInt(
		"Workers",
		d.Workers,
		MinDNSTTWorkers,
		MaxDNSTTWorkers,
	); err != nil {
		errs["Workers"] = err
	}

	if err := checkDomain("Domain", d.Domain); err != nil {
		errs["Domain"] = err
	}

	if err := checkPubKey("PublicKey", d.PublicKey); err != nil {
		errs["PublicKey"] = err
	}

	if err := checkDuration(
		"Timeout",
		d.Timeout.Duration(),
		MinDNSTTTimeout,
		MaxDNSTTTimeout,
	); err != nil {
		errs["Timeout"] = err
	}

	if err := checkPrefix("PrefixOutput", d.OutputPrefix); err != nil {
		errs["PrefixOutput"] = err
	}

	return errs
}

func validateSlipStream(s config.SlipStreamConfig) map[string]error {
	errs := map[string]error{}

	if err := checkInt(
		"Workers",
		s.Workers,
		MinSlipStreamWorkers,
		MaxSlipStreamWorkers,
	); err != nil {
		errs["Workers"] = err
	}

	if err := checkDomain("Domain", s.Domain); err != nil {
		errs["Domain"] = err
	}

	if err := checkDuration(
		"Timeout",
		s.Timeout.Duration(),
		MinSlipStreamTimeout,
		MaxSlipStreamTimeout,
	); err != nil {
		errs["Timeout"] = err
	}

	if err := checkPrefix("PrefixOutput", s.OutputPrefix); err != nil {
		errs["PrefixOutput"] = err
	}

	return errs
}

// NormalizeDNS replaces invalid DNSConfig fields with defaults and reports each correction.
func NormalizeDNS(cfg *config.DNSConfig) []Warning {
	var warns []Warning

	warns = append(warns, normalizeResolver(&cfg.Resolver)...)
	warns = append(warns, normalizeDNSTT(&cfg.DNSTT)...)
	warns = append(warns, normalizeSlipStream(&cfg.SlipStream)...)

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
		"Resolver.Protocol",
		&r.Protocol,
		allowedDNSProtocols,
		def.Protocol,
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
		"Resolver.DPITries",
		&r.DPITries,
		MinDNSResolverDPITries,
		MaxDNSResolverDPITries,
		def.DPITries,
		&warns,
	)

	fixDurationMS(
		"Resolver.DPITimeout",
		&r.DPITimeout,
		MinDNSResolverDPITimeout,
		MaxDNSResolverDPITimeout,
		def.DPITimeout,
		&warns,
	)

	fixPrefix(
		"Resolver.PrefixOutput",
		&r.PrefixOutput,
		def.PrefixOutput,
		&warns,
	)

	return warns
}

func normalizeDNSTT(d *config.DNSTTConfig) []Warning {
	def := config.DefaultDNSConfig().DNSTT
	var warns []Warning

	fixInt(
		"DNSTT.Workers",
		&d.Workers,
		MinDNSTTWorkers,
		MaxDNSTTWorkers,
		def.Workers,
		&warns,
	)

	fixDomain(
		"DNSTT.Domain",
		&d.Domain,
		def.Domain,
		&warns,
	)

	fixPubKey(
		"DNSTT.PublicKey",
		&d.PublicKey,
		def.PublicKey,
		&warns,
	)

	fixDurationMS(
		"DNSTT.Timeout",
		&d.Timeout,
		MinDNSTTTimeout,
		MaxDNSTTTimeout,
		def.Timeout,
		&warns,
	)

	fixPrefix(
		"DNSTT.PrefixOutput",
		&d.OutputPrefix,
		def.OutputPrefix,
		&warns,
	)

	return warns
}

func normalizeSlipStream(s *config.SlipStreamConfig) []Warning {
	def := config.DefaultDNSConfig().SlipStream
	var warns []Warning

	fixInt(
		"SlipStream.Workers",
		&s.Workers,
		MinSlipStreamWorkers,
		MaxSlipStreamWorkers,
		def.Workers,
		&warns,
	)

	fixDomain(
		"SlipStream.Domain",
		&s.Domain,
		def.Domain,
		&warns,
	)

	fixDurationMS(
		"SlipStream.Timeout",
		&s.Timeout,
		MinSlipStreamTimeout,
		MaxSlipStreamTimeout,
		def.Timeout,
		&warns,
	)

	fixPrefix(
		"SlipStream.PrefixOutput",
		&s.OutputPrefix,
		def.OutputPrefix,
		&warns,
	)

	return warns
}
