package validate

import (
	"time"

	"bgscan/internal/core/config"
)

// ValidateTCP strictly validates a TCPConfig and returns errors by field name.
func ValidateTCP(cfg config.TCPConfig) map[string]error {
	errs := map[string]error{}

	if err := checkInt("Workers", cfg.Workers, 1, 10000); err != nil {
		errs["Workers"] = err
	}

	if err := checkInt("Port", cfg.Port, 1, 65535); err != nil {
		errs["Port"] = err
	}

	if err := checkDuration("Timeout", cfg.Timeout.Duration(),
		100*time.Millisecond, 30*time.Second); err != nil {
		errs["Timeout"] = err
	}

	if err := checkUint16("Tries", cfg.Tries, 1, 10); err != nil {
		errs["Tries"] = err
	}

	if err := checkPrefix("PrefixOutput", cfg.OutputPrefix); err != nil {
		errs["PrefixOutput"] = err
	}

	return errs
}

// NormalizeTCP replaces invalid TCPConfig fields with defaults and reports each correction.
func NormalizeTCP(cfg *config.TCPConfig) []Warning {
	def := config.DefaultTCPConfig()
	var warns []Warning

	fixInt("Workers", &cfg.Workers, 1, 10000, def.Workers, &warns)
	fixInt("Port", &cfg.Port, 1, 65535, def.Port, &warns)

	fixDurationMS("Timeout", &cfg.Timeout,
		100*time.Millisecond, 30*time.Second, def.Timeout, &warns)

	fixUint16("Tries", &cfg.Tries, 1, 10, def.Tries, &warns)

	fixPrefix("PrefixOutput", &cfg.OutputPrefix, def.OutputPrefix, &warns)

	return warns
}
