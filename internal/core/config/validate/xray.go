package validate

import (
	"fmt"
	"time"

	"bgscan/internal/core/config"
)

var allowedPreScanTypes = []string{"tcp", "icmp", "none", "http"}

// ValidateXray strictly validates an XrayConfig and returns errors by field name.
func ValidateXray(cfg config.XrayConfig) map[string]error {
	errs := map[string]error{}

	if err := checkInt("Workers", cfg.Workers, 1, 1000); err != nil {
		errs["Workers"] = err
	}

	if !cfg.ConnectivityTestType.IsValid() {
		errs["ConnectivityTestType"] = errInvalidConnectivityTest()
	}

	if err := checkInt("DownloadSpeed", cfg.DownloadSpeed, 0, 100000); err != nil {
		errs["DownloadSpeed"] = err
	}

	if err := checkInt("UploadSpeed", cfg.UploadSpeed, 0, 100000); err != nil {
		errs["UploadSpeed"] = err
	}

	if err := checkDuration("Timeout", cfg.Timeout.Duration(),
		100*time.Millisecond, 60*time.Second); err != nil {
		errs["Timeout"] = err
	}

	if err := checkEnum("PreScanType", cfg.PreScanType, allowedPreScanTypes); err != nil {
		errs["PreScanType"] = err
	}

	if err := checkPrefix("PrefixOutput", cfg.OutputPrefix); err != nil {
		errs["PrefixOutput"] = err
	}

	return errs
}

// NormalizeXray replaces invalid XrayConfig fields with defaults and reports each correction.
func NormalizeXray(cfg *config.XrayConfig) []Warning {
	def := config.DefaultXrayConfig()
	var warns []Warning

	fixInt("Workers", &cfg.Workers, 1, 1000, def.Workers, &warns)

	if !cfg.ConnectivityTestType.IsValid() {
		warns = append(warns, Warning{
			Field:  "ConnectivityTestType",
			OldVal: cfg.ConnectivityTestType,
			NewVal: def.ConnectivityTestType,
			Reason: "invalid → default",
		})
		cfg.ConnectivityTestType = def.ConnectivityTestType
	}

	fixInt("DownloadSpeed", &cfg.DownloadSpeed, 0, 10000, def.DownloadSpeed, &warns)
	fixInt("UploadSpeed", &cfg.UploadSpeed, 0, 10000, def.UploadSpeed, &warns)

	fixDurationMS("Timeout", &cfg.Timeout,
		100*time.Millisecond, 60*time.Second, def.Timeout, &warns)

	fixEnum("PreScanType", &cfg.PreScanType, allowedPreScanTypes, def.PreScanType, &warns)
	fixString("PrefixOutput", &cfg.OutputPrefix, def.OutputPrefix, &warns)

	return warns
}

func errInvalidConnectivityTest() error {
	return fmt.Errorf("must be one of: ConnectivityOnly, DownloadSpeedOnly, UploadSpeedOnly, Both")
}
