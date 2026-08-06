package validate

import (
	"testing"
	"time"

	"bgscan/internal/core/config"
)

func getValidGeneralConfig() config.GeneralConfig {
	return config.GeneralConfig{
		StatusInterval:   config.NewDurationMS(5 * time.Second),
		StopAfterFound:   10,
		MaxIPsToTest:     100,
		MaxIPsPerStage:   1000,
		BatchSize:        500,
		PipelineMode:     "sequential",
		MinProbeDuration: config.NewDurationMS(50 * 2 * time.Millisecond),
		ProbePerSec:      25,
		ProbeBurst:       5,
	}
}

func TestValidateGeneral(t *testing.T) {
	tests := []struct {
		name        string
		mutateCfg   func(*config.GeneralConfig)
		wantErrKeys []string
	}{
		{
			name:        "valid config",
			mutateCfg:   func(c *config.GeneralConfig) {},
			wantErrKeys: nil,
		},
		{
			name: "StatusInterval too low",
			mutateCfg: func(c *config.GeneralConfig) {
				c.StatusInterval = config.NewDurationMS(MinGeneralStatusInterval - 2*time.Millisecond)
			},
			wantErrKeys: []string{"StatusInterval"},
		},
		{
			name: "StatusInterval too high",
			mutateCfg: func(c *config.GeneralConfig) {
				c.StatusInterval = config.NewDurationMS(MaxGeneralStatusInterval + time.Second)
			},
			wantErrKeys: []string{"StatusInterval"},
		},
		{
			name: "StopAfterFound negative",
			mutateCfg: func(c *config.GeneralConfig) {
				c.StopAfterFound = -1
			},
			wantErrKeys: []string{"StopAfterFound"},
		},
		{
			name: "MaxIPsToTest negative",
			mutateCfg: func(c *config.GeneralConfig) {
				c.MaxIPsToTest = -1
			},
			wantErrKeys: []string{"MaxIPsToTest"},
		},
		{
			name: "MaxIPsPerStage too low",
			mutateCfg: func(c *config.GeneralConfig) {
				c.MaxIPsPerStage = MinGeneralMaxIPsPerStage - 1
			},
			wantErrKeys: []string{"MaxIPsPerStage"},
		},
		{
			name: "MaxIPsPerStage too high",
			mutateCfg: func(c *config.GeneralConfig) {
				c.MaxIPsPerStage = MaxGeneralMaxIPsPerStage + 1
			},
			wantErrKeys: []string{"MaxIPsPerStage"},
		},
		{
			name: "BatchSize too low",
			mutateCfg: func(c *config.GeneralConfig) {
				c.BatchSize = MinGeneralBatchSize - 1
			},
			wantErrKeys: []string{"BatchSize"},
		},
		{
			name: "BatchSize too high",
			mutateCfg: func(c *config.GeneralConfig) {
				c.BatchSize = MaxGeneralBatchSize + 1
			},
			wantErrKeys: []string{"BatchSize"},
		},
		{
			name: "PipelineMode invalid",
			mutateCfg: func(c *config.GeneralConfig) {
				c.PipelineMode = "invalid_mode"
			},
			wantErrKeys: []string{"PipelineMode"},
		},
		{
			name: "MinProbeDuration too low",
			mutateCfg: func(c *config.GeneralConfig) {
				c.MinProbeDuration = config.NewDurationMS(MinGeneralProbeDuration - 2*time.Millisecond)
			},
			wantErrKeys: []string{"MinProbeDuration"},
		},
		{
			name: "MinProbeDuration too high",
			mutateCfg: func(c *config.GeneralConfig) {
				c.MinProbeDuration = config.NewDurationMS(MaxGeneralProbeDuration + 2*time.Millisecond)
			},
			wantErrKeys: []string{"MinProbeDuration"},
		},
		{
			name: "ProbePerSec too low",
			mutateCfg: func(c *config.GeneralConfig) {
				c.ProbePerSec = MinGeneralProbePerSec - 1
			},
			wantErrKeys: []string{"ProbePerSec"},
		},
		{
			name: "ProbePerSec too high",
			mutateCfg: func(c *config.GeneralConfig) {
				c.ProbePerSec = MaxGeneralProbePerSec + 1
			},
			wantErrKeys: []string{"ProbePerSec"},
		},
		{
			name: "ProbeBurst too low",
			mutateCfg: func(c *config.GeneralConfig) {
				c.ProbeBurst = MinGeneralProbeBurst - 1
			},
			wantErrKeys: []string{"ProbeBurst"},
		},
		{
			name: "ProbeBurst too high",
			mutateCfg: func(c *config.GeneralConfig) {
				c.ProbeBurst = MaxGeneralProbeBurst + 1
			},
			wantErrKeys: []string{"ProbeBurst"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := getValidGeneralConfig()
			tt.mutateCfg(&cfg)

			errs := ValidateGeneral(cfg)

			if len(errs) != len(tt.wantErrKeys) {
				t.Errorf(
					"ValidateGeneral() returned %d errors, want %d. Errors: %v",
					len(errs),
					len(tt.wantErrKeys),
					errs,
				)
			}

			for _, key := range tt.wantErrKeys {
				if _, ok := errs[key]; !ok {
					t.Errorf("ValidateGeneral() missing expected error for key %q", key)
				}
			}
		})
	}
}

func TestNormalizeGeneral(t *testing.T) {
	def := config.DefaultGeneralConfig()

	tests := []struct {
		name          string
		mutateCfg     func(*config.GeneralConfig)
		wantWarnCount int
		checkFixed    func(*testing.T, *config.GeneralConfig)
	}{
		{
			name:          "valid config",
			mutateCfg:     func(c *config.GeneralConfig) {},
			wantWarnCount: 0,
		},
		{
			name: "ProbePerSec invalid",
			mutateCfg: func(c *config.GeneralConfig) {
				c.ProbePerSec = 0
			},
			wantWarnCount: 1,
			checkFixed: func(t *testing.T, c *config.GeneralConfig) {
				if c.ProbePerSec != def.ProbePerSec {
					t.Errorf("ProbePerSec = %d, want %d", c.ProbePerSec, def.ProbePerSec)
				}
			},
		},
		{
			name: "ProbeBurst invalid",
			mutateCfg: func(c *config.GeneralConfig) {
				c.ProbeBurst = 0
			},
			wantWarnCount: 1,
			checkFixed: func(t *testing.T, c *config.GeneralConfig) {
				if c.ProbeBurst != def.ProbeBurst {
					t.Errorf("ProbeBurst = %d, want %d", c.ProbeBurst, def.ProbeBurst)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := getValidGeneralConfig()
			tt.mutateCfg(&cfg)

			warns := NormalizeGeneral(&cfg)

			if len(warns) != tt.wantWarnCount {
				t.Errorf(
					"NormalizeGeneral() returned %d warnings, want %d. Warnings: %v",
					len(warns),
					tt.wantWarnCount,
					warns,
				)
			}

			if tt.checkFixed != nil {
				tt.checkFixed(t, &cfg)
			}
		})
	}
}
