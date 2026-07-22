package validate

import (
	"math"
	"time"

	"bgscan/internal/core/config"
)

var allowedDNSProtocols = []string{"udp", "tcp", "dot", "doh"}

// ValidateDNS validates a DNSConfig strictly.
// Returns a map of field name → error for every invalid field.
// Nested fields use dot notation: "Resolver.Workers", "DNSTT.Domain", etc.
// Used by the UI OnValidate hook and SaveDNSConfig.
func ValidateDNS(cfg *config.DNSConfig) map[string]error {
	errs := map[string]error{}

	if cfg.Resolver != nil {
		for k, v := range validateResolver(cfg.Resolver) {
			errs["Resolver."+k] = v
		}
	}

	if cfg.DNSTT != nil && cfg.DNSTT.Enabled {
		for k, v := range validateDNSTT(cfg.DNSTT) {
			errs["DNSTT."+k] = v
		}
	}

	if cfg.SlipStream != nil && cfg.SlipStream.Enabled {
		for k, v := range validateSlipStream(cfg.SlipStream) {
			errs["SlipStream."+k] = v
		}
	}

	return errs
}

func validateResolver(r *config.ResolverConfig) map[string]error {
	errs := map[string]error{}

	if err := checkInt("Workers", r.Workers, 1, 2500); err != nil {
		errs["Workers"] = err
	}

	if err := checkEnum("Protocol", r.Protocol, allowedDNSProtocols); err != nil {
		errs["Protocol"] = err
	}

	if err := checkDomain("Domain", r.Domain); err != nil {
		errs["Domain"] = err
	}

	if err := checkUint16("Port", r.Port, 1, math.MaxUint16); err != nil {
		errs["Port"] = err
	}

	if err := checkStringSlice("CheckTypes", r.CheckTypes); err != nil {
		errs["CheckTypes"] = err
	}

	if err := checkDuration("Timeout", r.Timeout.Duration(),
		100*time.Millisecond, 30*time.Second); err != nil {
		errs["Timeout"] = err
	}

	if err := checkInt("Tries", r.Tries, 1, 10); err != nil {
		errs["Tries"] = err
	}

	if err := checkInt("DPITries", r.DPITries, 1, 10); err != nil {
		errs["DPITries"] = err
	}

	if err := checkDuration("DPITimeout", r.DPITimeout.Duration(),
		100*time.Millisecond, 10*time.Second); err != nil {
		errs["DPITimeout"] = err
	}

	if err := checkPrefix("PrefixOutput", r.PrefixOutput); err != nil {
		errs["PrefixOutput"] = err
	}

	return errs
}

func validateDNSTT(d *config.DNSTTConfig) map[string]error {
	errs := map[string]error{}

	if err := checkInt("Workers", d.Workers, 1, 500); err != nil {
		errs["Workers"] = err
	}

	if err := checkString("Domain", d.Domain); err != nil {
		errs["Domain"] = err
	}

	if err := checkDomain("Domain", d.Domain); err != nil {
		errs["Domain"] = err
	}

	// PublicKey must be non-empty when DNSTT is enabled
	if err := checkPubKey("PublicKey", d.PublicKey); err != nil {
		errs["PublicKey"] = err
	}

	if err := checkDuration("Timeout", d.Timeout.Duration(),
		100*time.Millisecond, 60*time.Second); err != nil {
		errs["Timeout"] = err
	}

	if err := checkPrefix("PrefixOutput", d.PrefixOutput); err != nil {
		errs["PrefixOutput"] = err
	}

	return errs
}

func validateSlipStream(s *config.SlipStreamConfig) map[string]error {
	errs := map[string]error{}

	if err := checkInt("Workers", s.Workers, 1, 500); err != nil {
		errs["Workers"] = err
	}

	if err := checkString("Domain", s.Domain); err != nil {
		errs["Domain"] = err
	}

	if err := checkDomain("Domain", s.Domain); err != nil {
		errs["Domain"] = err
	}

	if err := checkDuration("Timeout", s.Timeout.Duration(),
		100*time.Millisecond, 60*time.Second); err != nil {
		errs["Timeout"] = err
	}

	if err := checkPrefix("PrefixOutput", s.PrefixOutput); err != nil {
		errs["PrefixOutput"] = err
	}

	return errs
}

// NormalizeDNS auto-fixes invalid fields to their defaults.
// Returns a list of Warnings describing every correction made.
// Used only at TOML load time.
func NormalizeDNS(cfg *config.DNSConfig) []Warning {
	var warns []Warning

	if cfg.Resolver != nil {
		warns = append(warns, normalizeResolver(cfg.Resolver)...)
	}

	if cfg.DNSTT != nil && cfg.DNSTT.Enabled {
		warns = append(warns, normalizeDNSTT(cfg.DNSTT)...)
	}

	if cfg.SlipStream != nil && cfg.SlipStream.Enabled {
		warns = append(warns, normalizeSlipStream(cfg.SlipStream)...)
	}

	return warns
}

func normalizeResolver(r *config.ResolverConfig) []Warning {
	def := config.DefaultDNSConfig().Resolver
	var warns []Warning

	fixInt("Resolver.Workers", &r.Workers, 1, 2500, def.Workers, &warns)
	fixEnum("Resolver.Protocol", &r.Protocol, allowedDNSProtocols, def.Protocol, &warns)
	fixDomain("Resolver.Domain", &r.Domain, def.Domain, &warns)

	fixUint16("Resolver.Port", &r.Port, 1, math.MaxUint16, def.Port, &warns)
	fixStringSlice("Resolver.CheckTypes", &r.CheckTypes, def.CheckTypes, &warns)

	fixDurationMS("Resolver.Timeout", &r.Timeout,
		100*time.Millisecond, 30*time.Second, def.Timeout, &warns)

	fixInt("Resolver.Tries", &r.Tries, 1, 10, def.Tries, &warns)
	fixInt("Resolver.DPITries", &r.DPITries, 1, 10, def.DPITries, &warns)

	fixDurationMS("Resolver.DPITimeout", &r.DPITimeout,
		100*time.Millisecond, 10*time.Second, def.DPITimeout, &warns)

	fixPrefix("Resolver.PrefixOutput", &r.PrefixOutput, def.PrefixOutput, &warns)

	return warns
}

func normalizeDNSTT(d *config.DNSTTConfig) []Warning {
	def := config.DefaultDNSConfig().DNSTT
	var warns []Warning

	fixInt("DNSTT.Workers", &d.Workers, 1, 500, def.Workers, &warns)
	fixString("DNSTT.Domain", &d.Domain, def.Domain, &warns)
	fixDomain("DNSTT.Domain", &d.Domain, def.Domain, &warns)
	fixPubKey("DNSTT.PublicKey", &d.PublicKey, def.PublicKey, &warns)

	fixDurationMS("DNSTT.Timeout", &d.Timeout,
		100*time.Millisecond, 60*time.Second, def.Timeout, &warns)

	fixPrefix("DNSTT.PrefixOutput", &d.PrefixOutput, def.PrefixOutput, &warns)

	return warns
}

func normalizeSlipStream(s *config.SlipStreamConfig) []Warning {
	def := config.DefaultDNSConfig().SlipStream
	var warns []Warning

	fixInt("SlipStream.Workers", &s.Workers, 1, 500, def.Workers, &warns)
	fixDomain("SlipStream.Domain", &s.Domain, def.Domain, &warns)

	// CertPath intentionally skipped empty is valid

	fixDurationMS("SlipStream.Timeout", &s.Timeout,
		100*time.Millisecond, 60*time.Second, def.Timeout, &warns)

	fixPrefix("SlipStream.PrefixOutput", &s.PrefixOutput, def.PrefixOutput, &warns)

	return warns
}
