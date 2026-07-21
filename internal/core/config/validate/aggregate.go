package validate

import "bgscan/internal/core/config"

// AllWarnings holds normalization warnings grouped by config section.
// Returned by NormalizeAll after a TOML load.
type AllWarnings struct {
	General []Warning
	Writer  []Warning
	ICMP    []Warning
	TCP     []Warning
	HTTP    []Warning
	Xray    []Warning
	DNS     []Warning
}

// HasWarnings reports whether any section produced warnings.
func (a AllWarnings) HasWarnings() bool {
	return len(a.General) > 0 ||
		len(a.Writer) > 0 ||
		len(a.ICMP) > 0 ||
		len(a.TCP) > 0 ||
		len(a.HTTP) > 0 ||
		len(a.Xray) > 0 ||
		len(a.DNS) > 0
}

// NormalizeAll runs Normalize* on every live config section.
// Call this immediately after config.Init() at startup.
// Any corrected values must be saved back to disk by the caller.
func NormalizeAll() AllWarnings {
	return AllWarnings{
		General: NormalizeGeneral(config.GetGeneral()),
		Writer:  NormalizeWriter(config.GetWriter()),
		ICMP:    NormalizeICMP(config.GetICMP()),
		TCP:     NormalizeTCP(config.GetTCP()),
		HTTP:    NormalizeHTTP(config.GetHTTP()),
		Xray:    NormalizeXray(config.GetXray()),
		DNS:     NormalizeDNS(config.GetDNS()),
	}
}

// AllErrors holds strict validation errors grouped by config section.
// Each inner map is field name → error.
type AllErrors struct {
	General map[string]error
	Writer  map[string]error
	ICMP    map[string]error
	TCP     map[string]error
	HTTP    map[string]error
	Xray    map[string]error
	DNS     map[string]error
}

// HasErrors reports whether any section has validation errors.
func (a AllErrors) HasErrors() bool {
	return len(a.General) > 0 ||
		len(a.Writer) > 0 ||
		len(a.ICMP) > 0 ||
		len(a.TCP) > 0 ||
		len(a.HTTP) > 0 ||
		len(a.Xray) > 0 ||
		len(a.DNS) > 0
}

// ValidateAll runs Validate* on every live config section.
// Useful for a runtime health check endpoint.
func ValidateAll() AllErrors {
	return AllErrors{
		General: ValidateGeneral(config.GetGeneral()),
		Writer:  ValidateWriter(config.GetWriter()),
		ICMP:    ValidateICMP(config.GetICMP()),
		TCP:     ValidateTCP(config.GetTCP()),
		HTTP:    ValidateHTTP(config.GetHTTP()),
		Xray:    ValidateXray(config.GetXray()),
		DNS:     ValidateDNS(config.GetDNS()),
	}
}
