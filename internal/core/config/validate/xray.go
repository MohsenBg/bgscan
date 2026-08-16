package validate

import (
	"fmt"
	"time"

	"bgscan/internal/core/config"
)

var allowedPreScanTypes = []string{
	"tcp",
	"icmp",
	"none",
	"http",
}

const (
	MinXrayWorkers = 1
	MaxXrayWorkers = 500

	MinXrayDownloadSpeed = 0
	MaxXrayDownloadSpeed = 100000

	MinXrayUploadSpeed = 0
	MaxXrayUploadSpeed = 100000
)

const (
	MinXrayTimeout = 100 * time.Millisecond
	MaxXrayTimeout = 60 * time.Second
)

// ValidateXray strictly validates an XrayConfig and returns errors by field name.
func ValidateXray(cfg config.XrayConfig) map[string]error {
	errs := map[string]error{}

	if err := checkInt(
		"Workers",
		cfg.Workers,
		MinXrayWorkers,
		MaxXrayWorkers,
	); err != nil {
		errs["Workers"] = err
	}

	if !cfg.ConnectivityTestType.IsValid() {
		errs["ConnectivityTestType"] = errInvalidConnectivityTest()
	}

	if err := checkInt(
		"DownloadSpeed",
		cfg.DownloadSpeed,
		MinXrayDownloadSpeed,
		MaxXrayDownloadSpeed,
	); err != nil {
		errs["DownloadSpeed"] = err
	}

	if err := checkInt(
		"UploadSpeed",
		cfg.UploadSpeed,
		MinXrayUploadSpeed,
		MaxXrayUploadSpeed,
	); err != nil {
		errs["UploadSpeed"] = err
	}

	if err := checkDuration(
		"Timeout",
		cfg.Timeout.Duration(),
		MinXrayTimeout,
		MaxXrayTimeout,
	); err != nil {
		errs["Timeout"] = err
	}

	if err := checkEnum(
		"PreScanType",
		cfg.PreScanType,
		allowedPreScanTypes,
	); err != nil {
		errs["PreScanType"] = err
	}

	if err := checkPrefix(
		"OutputPrefix",
		cfg.OutputPrefix,
	); err != nil {
		errs["OutputPrefix"] = err
	}

	return errs
}

// NormalizeXray replaces invalid XrayConfig fields with defaults and reports each correction.
func NormalizeXray(cfg *config.XrayConfig) []Warning {
	def := config.DefaultXrayConfig()
	var warns []Warning

	fixInt(
		"Workers",
		&cfg.Workers,
		MinXrayWorkers,
		MaxXrayWorkers,
		def.Workers,
		&warns,
	)

	if !cfg.ConnectivityTestType.IsValid() {
		warns = append(warns, Warning{
			Field:  "ConnectivityTestType",
			OldVal: cfg.ConnectivityTestType,
			NewVal: def.ConnectivityTestType,
			Reason: "invalid → default",
		})

		cfg.ConnectivityTestType = def.ConnectivityTestType
	}

	fixInt(
		"DownloadSpeed",
		&cfg.DownloadSpeed,
		MinXrayDownloadSpeed,
		MaxXrayDownloadSpeed,
		def.DownloadSpeed,
		&warns,
	)

	fixInt(
		"UploadSpeed",
		&cfg.UploadSpeed,
		MinXrayUploadSpeed,
		MaxXrayUploadSpeed,
		def.UploadSpeed,
		&warns,
	)

	fixDurationMS(
		"Timeout",
		&cfg.Timeout,
		MinXrayTimeout,
		MaxXrayTimeout,
		def.Timeout,
		&warns,
	)

	fixEnum(
		"PreScanType",
		&cfg.PreScanType,
		allowedPreScanTypes,
		def.PreScanType,
		&warns,
	)

	fixString(
		"OutputPrefix",
		&cfg.OutputPrefix,
		def.OutputPrefix,
		&warns,
	)

	return warns
}

func errInvalidConnectivityTest() error {
	return fmt.Errorf(
		"must be one of: ConnectivityOnly, DownloadSpeedOnly, UploadSpeedOnly, Both",
	)
}
