package validate

import (
	"time"

	"bgscan/internal/core/config"
)

// ValidateICMP validates an ICMPConfig strictly.
// Returns a map of field name → error for every invalid field.
// Used by the UI OnValidate hook and SaveICMPConfig.
func ValidateICMP(cfg *config.ICMPConfig) map[string]error {
	errs := map[string]error{}

	if err := checkInt("Workers", cfg.Workers, 1, 10000); err != nil {
		errs["Workers"] = err
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

// NormalizeICMP auto-fixes invalid fields to their defaults.
// Returns a list of Warnings describing every correction made.
// Used only at TOML load time.
func NormalizeICMP(cfg *config.ICMPConfig) []Warning {
	def := config.DefaultICMPConfig()
	var warns []Warning

	fixInt("Workers", &cfg.Workers, 1, 10000, def.Workers, &warns)

	fixDurationMS("Timeout", &cfg.Timeout,
		100*time.Millisecond, 30*time.Second, def.Timeout, &warns)

	fixUint16("Tries", &cfg.Tries, 1, 10, def.Tries, &warns)

	fixString("PrefixOutput", &cfg.PrefixOutput, def.PrefixOutput, &warns)

	return warns
}
