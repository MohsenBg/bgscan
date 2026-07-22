package validate

import (
	"testing"
	"time"

	"bgscan/internal/core/config"
)

// getValidGeneralConfig returns a known-good configuration to use as a baseline for tests.
func getValidGeneralConfig() config.GeneralConfig {
	return config.GeneralConfig{
		StatusInterval: config.NewDurationMS(5 * time.Second),
		StopAfterFound: 10,
		MaxIPsToTest:   100,
		MaxIPsPerStage: 1000,
		BatchSize:      500,
		PipelineMode:   "sequential",
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
			name:        "StatusInterval too low",
			mutateCfg:   func(c *config.GeneralConfig) { c.StatusInterval = config.NewDurationMS(50 * time.Millisecond) },
			wantErrKeys: []string{"StatusInterval"},
		},
		{
			name:        "StatusInterval too high",
			mutateCfg:   func(c *config.GeneralConfig) { c.StatusInterval = config.NewDurationMS(2 * time.Minute) },
			wantErrKeys: []string{"StatusInterval"},
		},
		{
			name:        "StopAfterFound negative",
			mutateCfg:   func(c *config.GeneralConfig) { c.StopAfterFound = -1 },
			wantErrKeys: []string{"StopAfterFound"},
		},
		{
			name:        "MaxIPsToTest negative",
			mutateCfg:   func(c *config.GeneralConfig) { c.MaxIPsToTest = -1 },
			wantErrKeys: []string{"MaxIPsToTest"},
		},
		{
			name:        "MaxIPsPerStage too low",
			mutateCfg:   func(c *config.GeneralConfig) { c.MaxIPsPerStage = 0 },
			wantErrKeys: []string{"MaxIPsPerStage"},
		},
		{
			name:        "MaxIPsPerStage too high",
			mutateCfg:   func(c *config.GeneralConfig) { c.MaxIPsPerStage = 10_000_001 },
			wantErrKeys: []string{"MaxIPsPerStage"},
		},
		{
			name:        "BatchSize too low",
			mutateCfg:   func(c *config.GeneralConfig) { c.BatchSize = 0 },
			wantErrKeys: []string{"BatchSize"},
		},
		{
			name:        "BatchSize too high",
			mutateCfg:   func(c *config.GeneralConfig) { c.BatchSize = 10_000_001 },
			wantErrKeys: []string{"BatchSize"},
		},
		{
			name:        "PipelineMode invalid",
			mutateCfg:   func(c *config.GeneralConfig) { c.PipelineMode = "invalid_mode" },
			wantErrKeys: []string{"PipelineMode"},
		},
		{
			name: "multiple errors at once",
			mutateCfg: func(c *config.GeneralConfig) {
				c.StopAfterFound = -1
				c.MaxIPsToTest = -1
				c.PipelineMode = "invalid"
				c.BatchSize = 0
			},
			wantErrKeys: []string{"StopAfterFound", "MaxIPsToTest", "PipelineMode", "BatchSize"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := getValidGeneralConfig() // Start with a valid copy
			tt.mutateCfg(&cfg)

			errs := ValidateGeneral(&cfg)

			if len(errs) != len(tt.wantErrKeys) {
				t.Errorf("ValidateGeneral() returned %d errors, want %d. Errors: %v", len(errs), len(tt.wantErrKeys), errs)
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
		checkFixed    func(t *testing.T, cfg *config.GeneralConfig)
	}{
		{
			name:          "valid config",
			mutateCfg:     func(c *config.GeneralConfig) {},
			wantWarnCount: 0,
			checkFixed:    func(t *testing.T, c *config.GeneralConfig) {},
		},
		{
			name:          "StatusInterval too low",
			mutateCfg:     func(c *config.GeneralConfig) { c.StatusInterval = config.NewDurationMS(50 * time.Millisecond) },
			wantWarnCount: 1,
			checkFixed: func(t *testing.T, c *config.GeneralConfig) {
				if c.StatusInterval.Duration() != def.StatusInterval.Duration() {
					t.Errorf("StatusInterval not fixed, got %v, want %v", c.StatusInterval.Duration(), def.StatusInterval.Duration())
				}
			},
		},
		{
			name:          "StopAfterFound negative",
			mutateCfg:     func(c *config.GeneralConfig) { c.StopAfterFound = -5 },
			wantWarnCount: 1,
			checkFixed: func(t *testing.T, c *config.GeneralConfig) {
				if c.StopAfterFound != def.StopAfterFound {
					t.Errorf("StopAfterFound not fixed, got %d, want %d", c.StopAfterFound, def.StopAfterFound)
				}
			},
		},
		{
			name:          "MaxIPsToTest negative",
			mutateCfg:     func(c *config.GeneralConfig) { c.MaxIPsToTest = -5 },
			wantWarnCount: 1,
			checkFixed: func(t *testing.T, c *config.GeneralConfig) {
				if c.MaxIPsToTest != def.MaxIPsToTest {
					t.Errorf("MaxIPsToTest not fixed, got %d, want %d", c.MaxIPsToTest, def.MaxIPsToTest)
				}
			},
		},
		{
			name:          "MaxIPsPerStage invalid",
			mutateCfg:     func(c *config.GeneralConfig) { c.MaxIPsPerStage = 0 },
			wantWarnCount: 1,
			checkFixed: func(t *testing.T, c *config.GeneralConfig) {
				if c.MaxIPsPerStage != def.MaxIPsPerStage {
					t.Errorf("MaxIPsPerStage not fixed, got %d, want %d", c.MaxIPsPerStage, def.MaxIPsPerStage)
				}
			},
		},
		{
			name:          "BatchSize invalid",
			mutateCfg:     func(c *config.GeneralConfig) { c.BatchSize = 0 },
			wantWarnCount: 1,
			checkFixed: func(t *testing.T, c *config.GeneralConfig) {
				if c.BatchSize != def.BatchSize {
					t.Errorf("BatchSize not fixed, got %d, want %d", c.BatchSize, def.BatchSize)
				}
			},
		},
		{
			name:          "PipelineMode invalid",
			mutateCfg:     func(c *config.GeneralConfig) { c.PipelineMode = "invalid_mode" },
			wantWarnCount: 1,
			checkFixed: func(t *testing.T, c *config.GeneralConfig) {
				if c.PipelineMode != def.PipelineMode {
					t.Errorf("PipelineMode not fixed, got %q, want %q", c.PipelineMode, def.PipelineMode)
				}
			},
		},
		{
			name: "multiple fixes at once",
			mutateCfg: func(c *config.GeneralConfig) {
				c.StopAfterFound = -1
				c.MaxIPsToTest = -1
				c.PipelineMode = "invalid"
			},
			wantWarnCount: 3,
			checkFixed: func(t *testing.T, c *config.GeneralConfig) {
				if c.StopAfterFound != def.StopAfterFound || c.MaxIPsToTest != def.MaxIPsToTest || c.PipelineMode != def.PipelineMode {
					t.Errorf("Multiple fields were not fixed to their defaults correctly")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := getValidGeneralConfig() // Start with a valid copy
			tt.mutateCfg(&cfg)

			warns := NormalizeGeneral(&cfg)

			if len(warns) != tt.wantWarnCount {
				t.Errorf("NormalizeGeneral() returned %d warnings, want %d. Warnings: %v", len(warns), tt.wantWarnCount, warns)
			}

			// Verify the config was actually mutated to the expected default
			tt.checkFixed(t, &cfg)
		})
	}
}
