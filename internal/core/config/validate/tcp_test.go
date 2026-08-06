package validate

import (
	"testing"
	"time"

	"bgscan/internal/core/config"
)

func TestTCPConfig(t *testing.T) {
	def := config.DefaultTCPConfig()
	tests := []struct {
		name          string
		mutateCfg     func(*config.TCPConfig)
		wantErrKeys   []string
		wantWarnCount int
		checkFixed    func(t *testing.T, cfg *config.TCPConfig)
	}{
		{
			name:          "valid config",
			mutateCfg:     func(c *config.TCPConfig) {},
			wantErrKeys:   nil,
			wantWarnCount: 0,
			checkFixed:    func(t *testing.T, c *config.TCPConfig) {},
		},
		{
			name:          "Workers too low",
			mutateCfg:     func(c *config.TCPConfig) { c.Workers = MinTCPWorkers - 1 },
			wantErrKeys:   []string{"Workers"},
			wantWarnCount: 1,
			checkFixed: func(t *testing.T, c *config.TCPConfig) {
				if c.Workers != def.Workers {
					t.Errorf("Workers = %d, want %d", c.Workers, def.Workers)
				}
			},
		},
		{
			name: "Workers too high",
			mutateCfg: func(c *config.TCPConfig) {
				c.Workers = MaxTCPWorkers + 1
			},
			wantErrKeys:   []string{"Workers"},
			wantWarnCount: 1,
			checkFixed: func(t *testing.T, c *config.TCPConfig) {
				if c.Workers != def.Workers {
					t.Errorf("Workers = %d, want %d", c.Workers, def.Workers)
				}
			},
		},
		{
			name:          "Port too low",
			mutateCfg:     func(c *config.TCPConfig) { c.Port = MinTCPPort - 1 },
			wantErrKeys:   []string{"Port"},
			wantWarnCount: 1,
			checkFixed: func(t *testing.T, c *config.TCPConfig) {
				if c.Port != def.Port {
					t.Errorf("Port = %d, want %d", c.Port, def.Port)
				}
			},
		},
		{
			name:          "Port too high",
			mutateCfg:     func(c *config.TCPConfig) { c.Port = MaxTCPPort + 1 },
			wantErrKeys:   []string{"Port"},
			wantWarnCount: 1,
			checkFixed: func(t *testing.T, c *config.TCPConfig) {
				if c.Port != def.Port {
					t.Errorf("Port = %d, want %d", c.Port, def.Port)
				}
			},
		},
		{
			name:          "Timeout too high",
			mutateCfg:     func(c *config.TCPConfig) { c.Timeout = config.NewDurationMS(MaxTCPTimeout + time.Second) },
			wantErrKeys:   []string{"Timeout"},
			wantWarnCount: 1,
			checkFixed: func(t *testing.T, c *config.TCPConfig) {
				if c.Timeout.Duration() != def.Timeout.Duration() {
					t.Errorf("Timeout = %v, want %v", c.Timeout.Duration(), def.Timeout.Duration())
				}
			},
		},
		{
			name:          "Tries too low",
			mutateCfg:     func(c *config.TCPConfig) { c.Tries = MinTCPTries - 1 },
			wantErrKeys:   []string{"Tries"},
			wantWarnCount: 1,
			checkFixed: func(t *testing.T, c *config.TCPConfig) {
				if c.Tries != def.Tries {
					t.Errorf("Tries = %d, want %d", c.Tries, def.Tries)
				}
			},
		},
		{
			name:          "PrefixOutput invalid",
			mutateCfg:     func(c *config.TCPConfig) { c.OutputPrefix = "invalid/prefix" },
			wantErrKeys:   []string{"PrefixOutput"},
			wantWarnCount: 1,
			checkFixed: func(t *testing.T, c *config.TCPConfig) {
				if c.OutputPrefix != def.OutputPrefix {
					t.Errorf("PrefixOutput = %q, want %q", c.OutputPrefix, def.OutputPrefix)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := def
			tt.mutateCfg(&cfg)
			errs := ValidateTCP(cfg)
			if len(errs) != len(tt.wantErrKeys) {
				t.Errorf("ValidateTCP() returned %d errors, want %d. Errors: %v", len(errs), len(tt.wantErrKeys), errs)
			}
			for _, key := range tt.wantErrKeys {
				if _, ok := errs[key]; !ok {
					t.Errorf("ValidateTCP() missing expected error for key %q", key)
				}
			}

			cfg = def
			tt.mutateCfg(&cfg)
			warns := NormalizeTCP(&cfg)
			if len(warns) != tt.wantWarnCount {
				t.Errorf("NormalizeTCP() returned %d warnings, want %d. Warnings: %v", len(warns), tt.wantWarnCount, warns)
			}
			tt.checkFixed(t, &cfg)
		})
	}
}
