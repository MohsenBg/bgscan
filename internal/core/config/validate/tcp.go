package validate

import (
	"time"

	"bgscan/internal/core/config"
)

const (
	MinTCPWorkers = 1
	MaxTCPWorkers = 5000

	MinTCPPort = 1
	MaxTCPPort = 65535

	MinTCPTries = 1
	MaxTCPTries = 10
)

const (
	MinTCPTimeout = 100 * time.Millisecond
	MaxTCPTimeout = 30 * time.Second
)

// ValidateTCP strictly validates a TCPConfig and returns errors by field name.
func ValidateTCP(cfg config.TCPConfig) map[string]error {
	errs := map[string]error{}

	if err := checkInt("Workers", cfg.Workers, MinTCPWorkers, MaxTCPWorkers); err != nil {
		errs["Workers"] = err
	}

	if err := checkInt("Port", cfg.Port, MinTCPPort, MaxTCPPort); err != nil {
		errs["Port"] = err
	}

	if err := checkDuration("Timeout", cfg.Timeout.Duration(),
		MinTCPTimeout, MaxTCPTimeout); err != nil {
		errs["Timeout"] = err
	}

	if err := checkUint16("Tries", cfg.Tries, MinTCPTries, MaxTCPTries); err != nil {
		errs["Tries"] = err
	}

	if err := checkPrefix("OutputPrefix", cfg.OutputPrefix); err != nil {
		errs["OutputPrefix"] = err
	}

	return errs
}

// NormalizeTCP replaces invalid TCPConfig fields with defaults and reports each correction.
func NormalizeTCP(cfg *config.TCPConfig) []Warning {
	def := config.DefaultTCPConfig()
	var warns []Warning

	fixInt("Workers", &cfg.Workers,
		MinTCPWorkers, MaxTCPWorkers,
		def.Workers, &warns)

	fixInt("Port", &cfg.Port,
		MinTCPPort, MaxTCPPort,
		def.Port, &warns)

	fixDurationMS("Timeout", &cfg.Timeout,
		MinTCPTimeout, MaxTCPTimeout,
		def.Timeout, &warns)

	fixUint16("Tries", &cfg.Tries,
		MinTCPTries, MaxTCPTries,
		def.Tries, &warns)

	fixPrefix("OutputPrefix",
		&cfg.OutputPrefix,
		def.OutputPrefix,
		&warns)

	return warns
}
