package validate

import (
	"testing"
	"time"

	"bgscan/internal/core/config"
)

func TestHTTPConfig(t *testing.T) {
	def := config.DefaultHTTPConfig()

	makeValidHTTP := func() config.HTTPConfig {
		return config.HTTPConfig{
			Workers:             100,
			Host:                "example.com",
			ServerName:          "example.com",
			Port:                443,
			Protocol:            "https",
			Timeout:             config.NewDurationMS(5 * time.Second),
			Version:             "http2",
			MinTLSVersion:       "tls1.2",
			MaxTLSVersion:       "tls1.3",
			OutputPrefix:        "http-scan",
			AcceptedStatusCodes: []int{200, 301, 302},
		}
	}

	tests := []struct {
		name          string
		mutateCfg     func(*config.HTTPConfig)
		wantErrKeys   []string
		wantWarnCount int
		checkFixed    func(t *testing.T, cfg *config.HTTPConfig)
	}{
		{
			name:          "valid config",
			mutateCfg:     func(c *config.HTTPConfig) {},
			wantErrKeys:   nil,
			wantWarnCount: 0,
			checkFixed:    func(t *testing.T, c *config.HTTPConfig) {},
		},
		{
			name: "Workers and Port invalid",
			mutateCfg: func(c *config.HTTPConfig) {
				c.Workers = 0
				c.Port = 70000
			},
			wantErrKeys:   []string{"Workers", "Port"},
			wantWarnCount: 2,
			checkFixed: func(t *testing.T, c *config.HTTPConfig) {
				if c.Workers != def.Workers || c.Port != def.Port {
					t.Errorf("Workers or Port not fixed correctly")
				}
			},
		},
		{
			name: "Host and ServerName invalid",
			mutateCfg: func(c *config.HTTPConfig) {
				c.Host = "invalid_host!"
				c.ServerName = "invalid_sni:123"
			},
			wantErrKeys:   []string{"Host", "ServerName"},
			wantWarnCount: 2,
			checkFixed: func(t *testing.T, c *config.HTTPConfig) {
				if c.Host != def.Host || c.ServerName != def.ServerName {
					t.Errorf("Host or ServerName not fixed correctly")
				}
			},
		},
		{
			name: "Protocol and Version invalid",
			mutateCfg: func(c *config.HTTPConfig) {
				c.Protocol = "ftp"
				c.Version = "http4"
			},
			wantErrKeys:   []string{"Protocol", "Version"},
			wantWarnCount: 2,
			checkFixed: func(t *testing.T, c *config.HTTPConfig) {
				if c.Protocol != def.Protocol || c.Version != def.Version {
					t.Errorf("Protocol or Version not fixed correctly")
				}
			},
		},
		{
			name: "Timeout invalid",
			mutateCfg: func(c *config.HTTPConfig) {
				c.Timeout = config.NewDurationMS(time.Hour)
			},
			wantErrKeys:   []string{"Timeout"},
			wantWarnCount: 1,
			checkFixed: func(t *testing.T, c *config.HTTPConfig) {
				if c.Timeout != def.Timeout {
					t.Errorf("Timeout not fixed correctly")
				}
			},
		},
		{
			name: "TLS Max & Min Version invalid",
			mutateCfg: func(c *config.HTTPConfig) {
				c.MinTLSVersion = "min-tls-invalid"
				c.MaxTLSVersion = "max-tls-invalid"
			},
			wantErrKeys:   []string{"MinTLSVersion", "MaxTLSVersion"},
			wantWarnCount: 2,
			checkFixed: func(t *testing.T, c *config.HTTPConfig) {
				if c.MinTLSVersion != def.MinTLSVersion {
					t.Errorf("MinTLSVersion = %q, want %q", c.MinTLSVersion, def.MinTLSVersion)
				}
				if c.MaxTLSVersion != def.MaxTLSVersion {
					t.Errorf("MaxTLSVersion = %q, want %q", c.MaxTLSVersion, def.MaxTLSVersion)
				}
			},
		},

		{
			name: "TLS Version order invalid",
			mutateCfg: func(c *config.HTTPConfig) {
				c.MinTLSVersion = "tls1.3"
				c.MaxTLSVersion = "tls1.0"
			},
			wantErrKeys:   []string{"MinTLSVersion", "MaxTLSVersion"},
			wantWarnCount: 1,
			checkFixed: func(t *testing.T, c *config.HTTPConfig) {
				if c.MinTLSVersion != def.MinTLSVersion {
					t.Errorf("MinTLSVersion = %q, want %q", c.MinTLSVersion, def.MinTLSVersion)
				}
				if c.MaxTLSVersion != def.MaxTLSVersion {
					t.Errorf("MaxTLSVersion = %q, want %q", c.MaxTLSVersion, def.MaxTLSVersion)
				}
			},
		},
		{
			name: "PrefixOutput and AcceptedStatusCodes invalid",
			mutateCfg: func(c *config.HTTPConfig) {
				c.OutputPrefix = ""
				c.AcceptedStatusCodes = []int{999}
			},
			wantErrKeys:   []string{"PrefixOutput", "AcceptedStatusCodes"},
			wantWarnCount: 2,
			checkFixed: func(t *testing.T, c *config.HTTPConfig) {
				if c.OutputPrefix != def.OutputPrefix {
					t.Errorf("PrefixOutput not fixed")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := makeValidHTTP()
			tt.mutateCfg(&cfg)

			errs := ValidateHTTP(cfg)
			if len(errs) != len(tt.wantErrKeys) {
				t.Errorf("ValidateHTTP() returned %d errors, want %d. Errors: %v", len(errs), len(tt.wantErrKeys), errs)
			}
			for _, key := range tt.wantErrKeys {
				if _, ok := errs[key]; !ok {
					t.Errorf("ValidateHTTP() missing expected error for key %q", key)
				}
			}

			cfg = makeValidHTTP()
			tt.mutateCfg(&cfg)
			warns := NormalizeHTTP(&cfg)
			if len(warns) != tt.wantWarnCount {
				t.Errorf("NormalizeHTTP() returned %d warnings, want %d. Warnings: %v", len(warns), tt.wantWarnCount, warns)
			}
			tt.checkFixed(t, &cfg)
		})
	}
}
