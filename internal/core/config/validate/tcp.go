package validate

import (
	"time"

	"bgscan/internal/core/config"
)

// ValidateTCP validates a TCPConfig strictly.
// Returns a map of field name → error for every invalid field.
// Used by the UI OnValidate hook and SaveTCPConfig.
func ValidateTCP(cfg *config.TCPConfig) map[string]error {
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

	if err := checkPrefix("PrefixOutput", cfg.PrefixOutput); err != nil {
		errs["PrefixOutput"] = err
	}

	return errs
}

// NormalizeTCP auto-fixes invalid fields to their defaults.
// Returns a list of Warnings describing every correction made.
// Used only at TOML load time.
func NormalizeTCP(cfg *config.TCPConfig) []Warning {
	def := config.DefaultTCPConfig()
	var warns []Warning

	fixInt("Workers", &cfg.Workers, 1, 10000, def.Workers, &warns)
	fixInt("Port", &cfg.Port, 1, 65535, def.Port, &warns)

	fixDurationMS("Timeout", &cfg.Timeout,
		100*time.Millisecond, 30*time.Second, def.Timeout, &warns)

	fixUint16("Tries", &cfg.Tries, 1, 10, def.Tries, &warns)

	fixPrefix("PrefixOutput", &cfg.PrefixOutput, def.PrefixOutput, &warns)

	return warns
}
