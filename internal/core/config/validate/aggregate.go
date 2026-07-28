package validate

import "bgscan/internal/core/config"

// AllWarnings groups normalization warnings by configuration section.
type AllWarnings struct {
	General []Warning
	Writer  []Warning
	ICMP    []Warning
	TCP     []Warning
	HTTP    []Warning
	Xray    []Warning
	DNS     []Warning
}

// HasWarnings reports whether any configuration section produced warnings.
func (a AllWarnings) HasWarnings() bool {
	return len(a.General) > 0 ||
		len(a.Writer) > 0 ||
		len(a.ICMP) > 0 ||
		len(a.TCP) > 0 ||
		len(a.HTTP) > 0 ||
		len(a.Xray) > 0 ||
		len(a.DNS) > 0
}

// NormalizeAll normalizes every configuration section after configuration is loaded.
// The caller is responsible for persisting any corrections.
func NormalizeAll(cfg *config.ScannerConfig) AllWarnings {
	return AllWarnings{
		General: NormalizeGeneral(&cfg.General),
		Writer:  NormalizeWriter(&cfg.Writer),
		ICMP:    NormalizeICMP(&cfg.ICMP),
		TCP:     NormalizeTCP(&cfg.TCP),
		HTTP:    NormalizeHTTP(&cfg.HTTP),
		Xray:    NormalizeXray(&cfg.Xray),
		DNS:     NormalizeDNS(&cfg.DNS),
	}
}

// AllErrors groups strict validation errors by configuration section.
type AllErrors struct {
	General map[string]error
	Writer  map[string]error
	ICMP    map[string]error
	TCP     map[string]error
	HTTP    map[string]error
	Xray    map[string]error
	DNS     map[string]error
}

// HasErrors reports whether any configuration section contains validation errors.
func (a AllErrors) HasErrors() bool {
	return len(a.General) > 0 ||
		len(a.Writer) > 0 ||
		len(a.ICMP) > 0 ||
		len(a.TCP) > 0 ||
		len(a.HTTP) > 0 ||
		len(a.Xray) > 0 ||
		len(a.DNS) > 0
}

// ValidateAll strictly validates every configuration section.
func ValidateAll(cfg config.ScannerConfig) AllErrors {
	return AllErrors{
		General: ValidateGeneral(cfg.General),
		Writer:  ValidateWriter(cfg.Writer),
		ICMP:    ValidateICMP(cfg.ICMP),
		TCP:     ValidateTCP(cfg.TCP),
		HTTP:    ValidateHTTP(cfg.HTTP),
		Xray:    ValidateXray(cfg.Xray),
		DNS:     ValidateDNS(cfg.DNS),
	}
}
