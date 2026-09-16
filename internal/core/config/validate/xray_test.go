package validate

import (
	"testing"
	"time"

	"github.com/MohsenBg/bgscan/internal/core/config"
)

func TestXrayConfig(t *testing.T) {
	def := config.DefaultXrayConfig()

	tests := []struct {
		name          string
		mutateCfg     func(*config.XrayConfig)
		wantErrKeys   []string
		wantWarnCount int
		checkFixed    func(t *testing.T, cfg *config.XrayConfig)
	}{
		{
			name:          "valid config",
			mutateCfg:     func(c *config.XrayConfig) {},
			wantErrKeys:   nil,
			wantWarnCount: 0,
			checkFixed:    func(t *testing.T, c *config.XrayConfig) {},
		},
		{
			name: "Workers too low",
			mutateCfg: func(c *config.XrayConfig) {
				c.Workers = MinXrayWorkers - 1
			},
			wantErrKeys:   []string{"Workers"},
			wantWarnCount: 1,
			checkFixed: func(t *testing.T, c *config.XrayConfig) {
				if c.Workers != def.Workers {
					t.Errorf("Workers = %d, want %d", c.Workers, def.Workers)
				}
			},
		},
		{
			name: "Workers too high",
			mutateCfg: func(c *config.XrayConfig) {
				c.Workers = MaxXrayWorkers + 1
			},
			wantErrKeys:   []string{"Workers"},
			wantWarnCount: 1,
			checkFixed: func(t *testing.T, c *config.XrayConfig) {
				if c.Workers != def.Workers {
					t.Errorf("Workers = %d, want %d", c.Workers, def.Workers)
				}
			},
		},
		{
			name: "ConnectivityTestType invalid",
			mutateCfg: func(c *config.XrayConfig) {
				c.ConnectivityTestType = 99
			},
			wantErrKeys:   []string{"ConnectivityTestType"},
			wantWarnCount: 1,
			checkFixed: func(t *testing.T, c *config.XrayConfig) {
				if c.ConnectivityTestType != def.ConnectivityTestType {
					t.Errorf(
						"ConnectivityTestType = %d, want %d",
						c.ConnectivityTestType,
						def.ConnectivityTestType,
					)
				}
			},
		},
		{
			name: "DownloadSpeed too low",
			mutateCfg: func(c *config.XrayConfig) {
				c.DownloadSpeed = MinXrayDownloadSpeed - 1
			},
			wantErrKeys:   []string{"DownloadSpeed"},
			wantWarnCount: 1,
			checkFixed: func(t *testing.T, c *config.XrayConfig) {
				if c.DownloadSpeed != def.DownloadSpeed {
					t.Errorf("DownloadSpeed = %d, want %d", c.DownloadSpeed, def.DownloadSpeed)
				}
			},
		},
		{
			name: "DownloadSpeed too high",
			mutateCfg: func(c *config.XrayConfig) {
				c.DownloadSpeed = MaxXrayDownloadSpeed + 1
			},
			wantErrKeys:   []string{"DownloadSpeed"},
			wantWarnCount: 1,
			checkFixed: func(t *testing.T, c *config.XrayConfig) {
				if c.DownloadSpeed != def.DownloadSpeed {
					t.Errorf("DownloadSpeed = %d, want %d", c.DownloadSpeed, def.DownloadSpeed)
				}
			},
		},
		{
			name: "UploadSpeed too low",
			mutateCfg: func(c *config.XrayConfig) {
				c.UploadSpeed = MinXrayUploadSpeed - 1
			},
			wantErrKeys:   []string{"UploadSpeed"},
			wantWarnCount: 1,
			checkFixed: func(t *testing.T, c *config.XrayConfig) {
				if c.UploadSpeed != def.UploadSpeed {
					t.Errorf("UploadSpeed = %d, want %d", c.UploadSpeed, def.UploadSpeed)
				}
			},
		},
		{
			name: "UploadSpeed too high",
			mutateCfg: func(c *config.XrayConfig) {
				c.UploadSpeed = MaxXrayUploadSpeed + 1
			},
			wantErrKeys:   []string{"UploadSpeed"},
			wantWarnCount: 1,
			checkFixed: func(t *testing.T, c *config.XrayConfig) {
				if c.UploadSpeed != def.UploadSpeed {
					t.Errorf("UploadSpeed = %d, want %d", c.UploadSpeed, def.UploadSpeed)
				}
			},
		},
		{
			name: "Timeout too low",
			mutateCfg: func(c *config.XrayConfig) {
				c.Timeout = config.NewDurationMS(MinXrayTimeout - 2*time.Millisecond)
			},
			wantErrKeys:   []string{"Timeout"},
			wantWarnCount: 1,
			checkFixed: func(t *testing.T, c *config.XrayConfig) {
				if c.Timeout.Duration() != def.Timeout.Duration() {
					t.Errorf(
						"Timeout = %v, want %v",
						c.Timeout.Duration(),
						def.Timeout.Duration(),
					)
				}
			},
		},
		{
			name: "Timeout too high",
			mutateCfg: func(c *config.XrayConfig) {
				c.Timeout = config.NewDurationMS(MaxXrayTimeout + time.Second)
			},
			wantErrKeys:   []string{"Timeout"},
			wantWarnCount: 1,
			checkFixed: func(t *testing.T, c *config.XrayConfig) {
				if c.Timeout.Duration() != def.Timeout.Duration() {
					t.Errorf(
						"Timeout = %v, want %v",
						c.Timeout.Duration(),
						def.Timeout.Duration(),
					)
				}
			},
		},
		{
			name: "PreScanType invalid",
			mutateCfg: func(c *config.XrayConfig) {
				c.PreScanType = "invalid"
			},
			wantErrKeys:   []string{"PreScanType"},
			wantWarnCount: 1,
			checkFixed: func(t *testing.T, c *config.XrayConfig) {
				if c.PreScanType != def.PreScanType {
					t.Errorf(
						"PreScanType = %q, want %q",
						c.PreScanType,
						def.PreScanType,
					)
				}
			},
		},
		{
			name: "OutputPrefix empty",
			mutateCfg: func(c *config.XrayConfig) {
				c.OutputPrefix = ""
			},
			wantErrKeys:   []string{"OutputPrefix"},
			wantWarnCount: 1,
			checkFixed: func(t *testing.T, c *config.XrayConfig) {
				if c.OutputPrefix != def.OutputPrefix {
					t.Errorf(
						"OutputPrefix = %q, want %q",
						c.OutputPrefix,
						def.OutputPrefix,
					)
				}
			},
		},
		{
			name: "MaxConcurrentDials too low",
			mutateCfg: func(c *config.XrayConfig) {
				c.MaxConcurrentDials = MinXrayMaxConcurrentDials - 1
			},
			wantErrKeys:   []string{"MaxConcurrentDials"},
			wantWarnCount: 1,
			checkFixed: func(t *testing.T, c *config.XrayConfig) {
				if c.MaxConcurrentDials != def.MaxConcurrentDials {
					t.Errorf("MaxConcurrentDials = %d, want %d", c.MaxConcurrentDials, def.MaxConcurrentDials)
				}
			},
		},
		{
			name: "MaxConcurrentDials too high",
			mutateCfg: func(c *config.XrayConfig) {
				c.MaxConcurrentDials = MaxXrayMaxConcurrentDials + 1
			},
			wantErrKeys:   []string{"MaxConcurrentDials"},
			wantWarnCount: 1,
			checkFixed: func(t *testing.T, c *config.XrayConfig) {
				if c.MaxConcurrentDials != def.MaxConcurrentDials {
					t.Errorf("MaxConcurrentDials = %d, want %d", c.MaxConcurrentDials, def.MaxConcurrentDials)
				}
			},
		},
		{
			name: "DialMaxAttempts too low",
			mutateCfg: func(c *config.XrayConfig) {
				c.DialMaxAttempts = MinXrayDialMaxAttempts - 1
			},
			wantErrKeys:   []string{"DialMaxAttempts"},
			wantWarnCount: 1,
			checkFixed: func(t *testing.T, c *config.XrayConfig) {
				if c.DialMaxAttempts != def.DialMaxAttempts {
					t.Errorf("DialMaxAttempts = %d, want %d", c.DialMaxAttempts, def.DialMaxAttempts)
				}
			},
		},
		{
			name: "MaxConcurrentDials zero means unlimited",
			mutateCfg: func(c *config.XrayConfig) {
				c.MaxConcurrentDials = 0
			},
			wantErrKeys:   nil,
			wantWarnCount: 0,
			checkFixed: func(t *testing.T, c *config.XrayConfig) {
				if c.MaxConcurrentDials != 0 {
					t.Errorf("MaxConcurrentDials = %d, want 0", c.MaxConcurrentDials)
				}
			},
		},
		{
			name: "DialMaxAttempts too high",
			mutateCfg: func(c *config.XrayConfig) {
				c.DialMaxAttempts = MaxXrayDialMaxAttempts + 1
			},
			wantErrKeys:   []string{"DialMaxAttempts"},
			wantWarnCount: 1,
			checkFixed: func(t *testing.T, c *config.XrayConfig) {
				if c.DialMaxAttempts != def.DialMaxAttempts {
					t.Errorf("DialMaxAttempts = %d, want %d", c.DialMaxAttempts, def.DialMaxAttempts)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := def
			tt.mutateCfg(&cfg)

			errs := ValidateXray(cfg)

			if len(errs) != len(tt.wantErrKeys) {
				t.Errorf(
					"ValidateXray() returned %d errors, want %d. Errors: %v",
					len(errs),
					len(tt.wantErrKeys),
					errs,
				)
			}

			for _, key := range tt.wantErrKeys {
				if _, ok := errs[key]; !ok {
					t.Errorf(
						"ValidateXray() missing expected error for key %q",
						key,
					)
				}
			}

			cfg = def
			tt.mutateCfg(&cfg)

			warns := NormalizeXray(&cfg)

			if len(warns) != tt.wantWarnCount {
				t.Errorf(
					"NormalizeXray() returned %d warnings, want %d. Warnings: %v",
					len(warns),
					tt.wantWarnCount,
					warns,
				)
			}

			tt.checkFixed(t, &cfg)
		})
	}
}
