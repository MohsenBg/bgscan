package validate

import (
	"testing"
	"time"

	"bgscan/internal/core/config"
)

// ============================================================================
// ICMP Config Tests
// ============================================================================
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
			name:          "Workers too low",
			mutateCfg:     func(c *config.ICMPConfig) { c.Workers = 0 },
			wantErrKeys:   []string{"Workers"},
			wantWarnCount: 1,
			checkFixed: func(t *testing.T, c *config.ICMPConfig) {
				if c.Workers != def.Workers {
					t.Errorf("Workers = %d, want %d", c.Workers, def.Workers)
				}
			},
		},
		{
			name:          "Timeout too low",
			mutateCfg:     func(c *config.ICMPConfig) { c.Timeout = config.NewDurationMS(50 * time.Millisecond) },
			wantErrKeys:   []string{"Timeout"},
			wantWarnCount: 1,
			checkFixed: func(t *testing.T, c *config.ICMPConfig) {
				if c.Timeout.Duration() != def.Timeout.Duration() {
					t.Errorf("Timeout = %v, want %v", c.Timeout.Duration(), def.Timeout.Duration())
				}
			},
		},
		{
			name:          "Tries too high",
			mutateCfg:     func(c *config.ICMPConfig) { c.Tries = 11 },
			wantErrKeys:   []string{"Tries"},
			wantWarnCount: 1,
			checkFixed: func(t *testing.T, c *config.ICMPConfig) {
				if c.Tries != def.Tries {
					t.Errorf("Tries = %d, want %d", c.Tries, def.Tries)
				}
			},
		},
		{
			name:          "PrefixOutput empty",
			mutateCfg:     func(c *config.ICMPConfig) { c.PrefixOutput = "" },
			wantErrKeys:   []string{"PrefixOutput"},
			wantWarnCount: 1,
			checkFixed: func(t *testing.T, c *config.ICMPConfig) {
				if c.PrefixOutput != def.PrefixOutput {
					t.Errorf("PrefixOutput = %q, want %q", c.PrefixOutput, def.PrefixOutput)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test Validate
			cfg := def
			tt.mutateCfg(cfg)
			errs := ValidateICMP(cfg)
			if len(errs) != len(tt.wantErrKeys) {
				t.Errorf("ValidateICMP() returned %d errors, want %d. Errors: %v", len(errs), len(tt.wantErrKeys), errs)
			}
			for _, key := range tt.wantErrKeys {
				if _, ok := errs[key]; !ok {
					t.Errorf("ValidateICMP() missing expected error for key %q", key)
				}
			}

			// Test Normalize
			cfg = def // Reset to defaults
			tt.mutateCfg(cfg)
			warns := NormalizeICMP(cfg)
			if len(warns) != tt.wantWarnCount {
				t.Errorf("NormalizeICMP() returned %d warnings, want %d. Warnings: %v", len(warns), tt.wantWarnCount, warns)
			}
			tt.checkFixed(t, cfg)
		})
	}
}
