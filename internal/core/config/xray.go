package config

import (
	"time"

	"bgscan/internal/logger"
)

// XrayConfig defines configuration for Xray connectivity testing.
type XrayConfig struct {
	Workers              int              `toml:"workers" comment:"Concurrent Xray test workers. Range: 1-500. Higher = faster but more bandwidth."`
	ConnectivityTestType ConnectivityTest `toml:"connectivity_test_type" comment:"Test type: 0=connectivity only, 1=download, 2=upload, 3=both."`
	DownloadSpeed        int              `toml:"download_speed" comment:"Target download speed in Kbps. Range: 0-100000. Used when test type includes download."`
	UploadSpeed          int              `toml:"upload_speed" comment:"Target upload speed in Kbps. Range: 0-100000. Used when test type includes upload."`
	Timeout              DurationMS       `toml:"timeout" comment:"Max wait for Xray test, in ms. Range: 100-60000."`
	OutputPrefix         string           `toml:"output_prefix" comment:"Filename prefix for result files."`
	PreScanType          string           `toml:"pre_scan_type" comment:"Pre-scan test before Xray: tcp, icmp, http, or none (skip pre-scan)."`
}

func DefaultXrayConfig() XrayConfig {
	platform := DetectPlatform()
	tier := SelectTier(CheckResources())

	tiers, ok := xrayDefaults[platform]
	if !ok {
		logger.CoreWarn("bgscan: no Xray defaults for platform %q, falling back to %q", platform, Desktop)
		platform = Desktop
		tiers = xrayDefaults[platform]
	}

	cfg, ok := tiers[tier]
	if !ok {
		logger.CoreWarn("bgscan: no Xray defaults for platform %q tier %q, falling back to %q", platform, tier, Mid)
		cfg = tiers[Mid]
	}

	return cfg
}

// xrayBase holds fields shared across every platform/tier.
func xrayBase() XrayConfig {
	return XrayConfig{
		ConnectivityTestType: Both,
		OutputPrefix:         "xray_",
		PreScanType:          "tcp",
	}
}

var xrayDefaults = map[Platform]map[Tier]XrayConfig{
	Server: {
		Low:  withXray(xrayBase(), 16, 50, 25, 4*time.Second),
		Mid:  withXray(xrayBase(), 32, 100, 50, 4*time.Second),
		High: withXray(xrayBase(), 64, 200, 100, 3*time.Second),
	},
	Desktop: {
		Low:  withXray(xrayBase(), 8, 50, 25, 4*time.Second),
		Mid:  withXray(xrayBase(), 16, 100, 50, 4*time.Second),
		High: withXray(xrayBase(), 32, 150, 75, 4*time.Second),
	},
	Android: {
		Low:  withXray(xrayBase(), 4, 20, 10, 6*time.Second),
		Mid:  withXray(xrayBase(), 8, 50, 25, 5*time.Second),
		High: withXray(xrayBase(), 16, 80, 40, 5*time.Second),
	},
}

// withXray returns a copy of base with the platform/tier-varying fields overridden.
func withXray(base XrayConfig, workers, downloadSpeed, uploadSpeed int, timeout time.Duration) XrayConfig {
	base.Workers = workers
	base.DownloadSpeed = downloadSpeed
	base.UploadSpeed = uploadSpeed
	base.Timeout = NewDurationMS(timeout)
	return base
}
