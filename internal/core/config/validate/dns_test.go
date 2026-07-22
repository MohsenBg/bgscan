package validate

import (
	"testing"
	"time"

	"bgscan/internal/core/config"
)

// ============================================================================
// DNS Config Tests
// ============================================================================
func TestDNSConfig(t *testing.T) {
	def := config.DefaultDNSConfig()

	makeValidDNS := func() config.DNSConfig {
		return config.DNSConfig{
			Resolver: &config.ResolverConfig{
				Workers:      100,
				Protocol:     "udp",
				Domain:       "google.com",
				Port:         53,
				CheckTypes:   []string{"A", "AAAA"},
				Timeout:      config.NewDurationMS(2 * time.Second),
				Tries:        3,
				DPITries:     2,
				DPITimeout:   config.NewDurationMS(1 * time.Second),
				PrefixOutput: "dns-resolver",
			},
			DNSTT: &config.DNSTTConfig{
				Enabled:      true,
				Workers:      100,
				Domain:       "dnstt.com",
				PublicKey:    "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
				Timeout:      config.NewDurationMS(5 * time.Second),
				PrefixOutput: "dnstt",
			},
			SlipStream: &config.SlipStreamConfig{
				Enabled:      true,
				Workers:      100,
				Domain:       "slip.com",
				Timeout:      config.NewDurationMS(5 * time.Second),
				PrefixOutput: "slip",
			},
		}
	}

	tests := []struct {
		name          string
		mutateCfg     func(*config.DNSConfig)
		wantErrKeys   []string
		wantWarnCount int
		checkFixed    func(t *testing.T, cfg *config.DNSConfig)
	}{
		{
			name:          "valid config",
			mutateCfg:     func(c *config.DNSConfig) {},
			wantErrKeys:   nil,
			wantWarnCount: 0,
			checkFixed:    func(t *testing.T, c *config.DNSConfig) {},
		},
		{
			name: "Resolver invalid fields",
			mutateCfg: func(c *config.DNSConfig) {
				c.Resolver.Workers = 0
				c.Resolver.Protocol = "invalid"
				c.Resolver.Domain = ""
				c.Resolver.Port = 0
				c.Resolver.CheckTypes = []string{}
				c.Resolver.Timeout = config.NewDurationMS(50 * time.Millisecond)
				c.Resolver.Tries = 0
				c.Resolver.DPITries = 0
				c.Resolver.DPITimeout = config.NewDurationMS(50 * time.Millisecond)
				c.Resolver.PrefixOutput = ""
			},
			wantErrKeys: []string{
				"Resolver.Workers", "Resolver.Protocol", "Resolver.Domain", "Resolver.Port",
				"Resolver.CheckTypes", "Resolver.Timeout", "Resolver.Tries", "Resolver.DPITries",
				"Resolver.DPITimeout", "Resolver.PrefixOutput",
			},
			wantWarnCount: 10,
			checkFixed: func(t *testing.T, c *config.DNSConfig) {
				if c.Resolver.Workers != def.Resolver.Workers {
					t.Errorf("Resolver.Workers = %d, want %d", c.Resolver.Workers, def.Resolver.Workers)
				}
				if c.Resolver.Domain != def.Resolver.Domain {
					t.Errorf("Resolver.Domain = %q, want %q", c.Resolver.Domain, def.Resolver.Domain)
				}
			},
		},
		{
			name: "Invalid domains in Resolver, DNSTT, and SlipStream",
			mutateCfg: func(c *config.DNSConfig) {
				c.Resolver.Domain = "invalid domain!"
				c.DNSTT.Domain = "bad.domain.123"
				c.SlipStream.Domain = "-invalid.com"
			},
			wantErrKeys: []string{
				"Resolver.Domain",
				"DNSTT.Domain",
				"SlipStream.Domain",
			},
			wantWarnCount: 3,
			checkFixed: func(t *testing.T, c *config.DNSConfig) {
				if c.Resolver.Domain != def.Resolver.Domain {
					t.Errorf("Resolver.Domain = %q, want %q", c.Resolver.Domain, def.Resolver.Domain)
				}
				if c.DNSTT.Domain != def.DNSTT.Domain {
					t.Errorf("DNSTT.Domain = %q, want %q", c.DNSTT.Domain, def.DNSTT.Domain)
				}
				if c.SlipStream.Domain != def.SlipStream.Domain {
					t.Errorf("SlipStream.Domain = %q, want %q", c.SlipStream.Domain, def.SlipStream.Domain)
				}
			},
		},
		{
			name: "empty domains in Resolver, DNSTT, and SlipStream",
			mutateCfg: func(c *config.DNSConfig) {
				c.Resolver.Domain = ""
				c.DNSTT.Domain = ""
				c.SlipStream.Domain = ""
			},
			wantErrKeys: []string{
				"Resolver.Domain",
				"DNSTT.Domain",
				"SlipStream.Domain",
			},
			wantWarnCount: 3,
			checkFixed: func(t *testing.T, c *config.DNSConfig) {
				if c.Resolver.Domain != def.Resolver.Domain {
					t.Errorf("Resolver.Domain = %q, want %q", c.Resolver.Domain, def.Resolver.Domain)
				}
				if c.DNSTT.Domain != def.DNSTT.Domain {
					t.Errorf("DNSTT.Domain = %q, want %q", c.DNSTT.Domain, def.DNSTT.Domain)
				}
				if c.SlipStream.Domain != def.SlipStream.Domain {
					t.Errorf("SlipStream.Domain = %q, want %q", c.SlipStream.Domain, def.SlipStream.Domain)
				}
			},
		},

		{
			name: "DNSTT invalid fields",
			mutateCfg: func(c *config.DNSConfig) {
				c.DNSTT.Workers = 0
				c.DNSTT.Domain = "invalid domain!"
				c.DNSTT.PublicKey = "short"
				c.DNSTT.Timeout = config.NewDurationMS(50 * time.Millisecond)
				c.DNSTT.PrefixOutput = ""
			},
			wantErrKeys: []string{
				"DNSTT.Workers", "DNSTT.Domain", "DNSTT.PublicKey", "DNSTT.Timeout", "DNSTT.PrefixOutput",
			},
			wantWarnCount: 5,
			checkFixed: func(t *testing.T, c *config.DNSConfig) {
				if c.DNSTT.Domain != def.DNSTT.Domain {
					t.Errorf("DNSTT.Domain = %q, want %q", c.DNSTT.Domain, def.DNSTT.Domain)
				}
			},
		},
		{
			name: "SlipStream invalid fields",
			mutateCfg: func(c *config.DNSConfig) {
				c.SlipStream.Workers = 0
				c.SlipStream.Domain = "invalid domain!"
				c.SlipStream.Timeout = config.NewDurationMS(50 * time.Millisecond)
				c.SlipStream.PrefixOutput = ""
			},
			wantErrKeys: []string{
				"SlipStream.Workers", "SlipStream.Domain", "SlipStream.Timeout", "SlipStream.PrefixOutput",
			},
			wantWarnCount: 4,
			checkFixed: func(t *testing.T, c *config.DNSConfig) {
				if c.SlipStream.Domain != def.SlipStream.Domain {
					t.Errorf("SlipStream.Domain = %q, want %q", c.SlipStream.Domain, def.SlipStream.Domain)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := makeValidDNS()
			tt.mutateCfg(&cfg)

			errs := ValidateDNS(&cfg)
			if len(errs) != len(tt.wantErrKeys) {
				t.Errorf("ValidateDNS() returned %d errors, want %d. Errors: %v", len(errs), len(tt.wantErrKeys), errs)
			}
			for _, key := range tt.wantErrKeys {
				if _, ok := errs[key]; !ok {
					t.Errorf("ValidateDNS() missing expected error for key %q", key)
				}
			}

			cfg = makeValidDNS()
			tt.mutateCfg(&cfg)
			warns := NormalizeDNS(&cfg)
			if len(warns) != tt.wantWarnCount {
				t.Errorf("NormalizeDNS() returned %d warnings, want %d. Warnings: %v", len(warns), tt.wantWarnCount, warns)
			}
			tt.checkFixed(t, &cfg)
		})
	}
}
