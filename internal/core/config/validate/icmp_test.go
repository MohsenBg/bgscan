package validate

import (
	"testing"
	"time"

	"github.com/MohsenBg/bgscan/internal/core/config"
)

func TestICMPConfig(t *testing.T) {
	def := config.DefaultICMPConfig()

	tests := []struct {
		name          string
		mutateCfg     func(*config.ICMPConfig)
		wantErrKeys   []string
		wantWarnCount int
		checkFixed    func(t *testing.T, cfg *config.ICMPConfig)
	}{
		{
			name:          "valid config",
			mutateCfg:     func(c *config.ICMPConfig) {},
			wantErrKeys:   nil,
			wantWarnCount: 0,
			checkFixed:    func(t *testing.T, c *config.ICMPConfig) {},
		},
		{
			name: "Workers too low",
			mutateCfg: func(c *config.ICMPConfig) {
				c.Workers = MinICMPWorkers - 1
			},
			wantErrKeys:   []string{"Workers"},
			wantWarnCount: 1,
			checkFixed: func(t *testing.T, c *config.ICMPConfig) {
				if c.Workers != def.Workers {
					t.Errorf("Workers = %d, want %d", c.Workers, def.Workers)
				}
			},
		},
		{
			name: "Workers too high",
			mutateCfg: func(c *config.ICMPConfig) {
				c.Workers = MaxICMPWorkers + 1
			},
			wantErrKeys:   []string{"Workers"},
			wantWarnCount: 1,
			checkFixed: func(t *testing.T, c *config.ICMPConfig) {
				if c.Workers != def.Workers {
					t.Errorf("Workers = %d, want %d", c.Workers, def.Workers)
				}
			},
		},
		{
			name: "Timeout too low",
			mutateCfg: func(c *config.ICMPConfig) {
				c.Timeout = config.NewDurationMS(MinICMPTimeout - 2*time.Millisecond)
			},
			wantErrKeys:   []string{"Timeout"},
			wantWarnCount: 1,
			checkFixed: func(t *testing.T, c *config.ICMPConfig) {
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
			mutateCfg: func(c *config.ICMPConfig) {
				c.Timeout = config.NewDurationMS(MaxICMPTimeout + 2*time.Millisecond)
			},
			wantErrKeys:   []string{"Timeout"},
			wantWarnCount: 1,
			checkFixed: func(t *testing.T, c *config.ICMPConfig) {
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
			name: "Tries too low",
			mutateCfg: func(c *config.ICMPConfig) {
				c.Tries = MinICMPTries - 1
			},
			wantErrKeys:   []string{"Tries"},
			wantWarnCount: 1,
			checkFixed: func(t *testing.T, c *config.ICMPConfig) {
				if c.Tries != def.Tries {
					t.Errorf("Tries = %d, want %d", c.Tries, def.Tries)
				}
			},
		},
		{
			name: "Tries too high",
			mutateCfg: func(c *config.ICMPConfig) {
				c.Tries = MaxICMPTries + 1
			},
			wantErrKeys:   []string{"Tries"},
			wantWarnCount: 1,
			checkFixed: func(t *testing.T, c *config.ICMPConfig) {
				if c.Tries != def.Tries {
					t.Errorf("Tries = %d, want %d", c.Tries, def.Tries)
				}
			},
		},
		{
			name: "OutputPrefix empty",
			mutateCfg: func(c *config.ICMPConfig) {
				c.OutputPrefix = ""
			},
			wantErrKeys:   []string{"OutputPrefix"},
			wantWarnCount: 1,
			checkFixed: func(t *testing.T, c *config.ICMPConfig) {
				if c.OutputPrefix != def.OutputPrefix {
					t.Errorf(
						"OutputPrefix = %q, want %q",
						c.OutputPrefix,
						def.OutputPrefix,
					)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := def
			tt.mutateCfg(&cfg)

			errs := ValidateICMP(cfg)

			if len(errs) != len(tt.wantErrKeys) {
				t.Errorf(
					"ValidateICMP() returned %d errors, want %d. Errors: %v",
					len(errs),
					len(tt.wantErrKeys),
					errs,
				)
			}

			for _, key := range tt.wantErrKeys {
				if _, ok := errs[key]; !ok {
					t.Errorf(
						"ValidateICMP() missing expected error for key %q",
						key,
					)
				}
			}

			cfg = def
			tt.mutateCfg(&cfg)

			warns := NormalizeICMP(&cfg)

			if len(warns) != tt.wantWarnCount {
				t.Errorf(
					"NormalizeICMP() returned %d warnings, want %d. Warnings: %v",
					len(warns),
					tt.wantWarnCount,
					warns,
				)
			}

			tt.checkFixed(t, &cfg)
		})
	}
}
