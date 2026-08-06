package validate

import (
	"testing"
	"time"

	"bgscan/internal/core/config"
)

func TestDNSConfig(t *testing.T) {
	tests := []struct {
		name          string
		mutate        func(*config.DNSConfig)
		wantErrKeys   []string
		wantWarnCount int
		checkFixed    func(*testing.T, config.DNSConfig, config.DNSConfig)
	}{
		{
			name:          "valid config",
			mutate:        func(*config.DNSConfig) {},
			wantWarnCount: 0,
		},
		{
			name: "invalid resolver fields",
			mutate: func(cfg *config.DNSConfig) {
				cfg.Resolver.Workers = MinDNSResolverWorkers - 1
				cfg.Resolver.Protocol = "invalid"
				cfg.Resolver.Domain = ""
				cfg.Resolver.Port = MinDNSResolverPort - 1
				cfg.Resolver.CheckTypes = []string{}
				cfg.Resolver.Timeout = config.NewDurationMS(MinDNSResolverTimeout - 2*time.Millisecond)
				cfg.Resolver.Tries = MinDNSResolverTries - 1
				cfg.Resolver.DPITries = MinDNSResolverDPITries - 1
				cfg.Resolver.DPITimeout = config.NewDurationMS(MinDNSResolverDPITimeout - 2*time.Millisecond)
				cfg.Resolver.PrefixOutput = ""
			},
			wantErrKeys: []string{
				"Resolver.Workers",
				"Resolver.Protocol",
				"Resolver.Domain",
				"Resolver.Port",
				"Resolver.CheckTypes",
				"Resolver.Timeout",
				"Resolver.Tries",
				"Resolver.DPITries",
				"Resolver.DPITimeout",
				"Resolver.PrefixOutput",
			},
			wantWarnCount: 10,
			checkFixed: func(t *testing.T, got, want config.DNSConfig) {
				if got.Resolver.Workers != want.Resolver.Workers {
					t.Errorf(
						"Resolver.Workers = %d, want %d",
						got.Resolver.Workers,
						want.Resolver.Workers,
					)
				}

				if got.Resolver.Protocol != want.Resolver.Protocol {
					t.Errorf(
						"Resolver.Protocol = %q, want %q",
						got.Resolver.Protocol,
						want.Resolver.Protocol,
					)
				}

				if got.Resolver.Domain != want.Resolver.Domain {
					t.Errorf(
						"Resolver.Domain = %q, want %q",
						got.Resolver.Domain,
						want.Resolver.Domain,
					)
				}

				if got.Resolver.Port != want.Resolver.Port {
					t.Errorf(
						"Resolver.Port = %d, want %d",
						got.Resolver.Port,
						want.Resolver.Port,
					)
				}
			},
		},
		{
			name: "invalid domains",
			mutate: func(cfg *config.DNSConfig) {
				cfg.Resolver.Domain = "invalid domain!"
				cfg.DNSTT.Domain = "invalid domain!"
				cfg.SlipStream.Domain = "-invalid.com"
			},
			wantErrKeys: []string{
				"Resolver.Domain",
				"DNSTT.Domain",
				"SlipStream.Domain",
			},
			wantWarnCount: 3,
			checkFixed: func(t *testing.T, got, want config.DNSConfig) {
				if got.Resolver.Domain != want.Resolver.Domain {
					t.Errorf(
						"Resolver.Domain = %q, want %q",
						got.Resolver.Domain,
						want.Resolver.Domain,
					)
				}

				if got.DNSTT.Domain != want.DNSTT.Domain {
					t.Errorf(
						"DNSTT.Domain = %q, want %q",
						got.DNSTT.Domain,
						want.DNSTT.Domain,
					)
				}

				if got.SlipStream.Domain != want.SlipStream.Domain {
					t.Errorf(
						"SlipStream.Domain = %q, want %q",
						got.SlipStream.Domain,
						want.SlipStream.Domain,
					)
				}
			},
		},
		{
			name: "invalid DNSTT fields",
			mutate: func(cfg *config.DNSConfig) {
				cfg.DNSTT.Workers = MinDNSTTWorkers - 1
				cfg.DNSTT.Domain = "invalid domain!"
				cfg.DNSTT.PublicKey = "short"
				cfg.DNSTT.Timeout = config.NewDurationMS(MinDNSTTTimeout - 2*time.Millisecond)
				cfg.DNSTT.OutputPrefix = ""
			},
			wantErrKeys: []string{
				"DNSTT.Workers",
				"DNSTT.Domain",
				"DNSTT.PublicKey",
				"DNSTT.Timeout",
				"DNSTT.PrefixOutput",
			},
			wantWarnCount: 5,
		},
		{
			name: "invalid SlipStream fields",
			mutate: func(cfg *config.DNSConfig) {
				cfg.SlipStream.Workers = MinSlipStreamWorkers - 1
				cfg.SlipStream.Domain = "invalid domain!"
				cfg.SlipStream.Timeout = config.NewDurationMS(MinSlipStreamTimeout - 2*time.Millisecond)
				cfg.SlipStream.OutputPrefix = ""
			},
			wantErrKeys: []string{
				"SlipStream.Workers",
				"SlipStream.Domain",
				"SlipStream.Timeout",
				"SlipStream.PrefixOutput",
			},
			wantWarnCount: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.DefaultDNSConfig()
			tt.mutate(&cfg)

			errs := ValidateDNS(cfg)

			if len(errs) != len(tt.wantErrKeys) {
				t.Errorf(
					"ValidateDNS() returned %d errors, want %d: %v",
					len(errs),
					len(tt.wantErrKeys),
					errs,
				)
			}

			for _, key := range tt.wantErrKeys {
				if _, ok := errs[key]; !ok {
					t.Errorf(
						"ValidateDNS() missing error for %q",
						key,
					)
				}
			}

			cfg = config.DefaultDNSConfig()
			tt.mutate(&cfg)

			warnings := NormalizeDNS(&cfg)

			if len(warnings) != tt.wantWarnCount {
				t.Errorf(
					"NormalizeDNS() returned %d warnings, want %d: %v",
					len(warnings),
					tt.wantWarnCount,
					warnings,
				)
			}

			if errs := ValidateDNS(cfg); len(errs) != 0 {
				t.Errorf(
					"normalized config is still invalid: %v",
					errs,
				)
			}

			if tt.checkFixed != nil {
				tt.checkFixed(
					t,
					cfg,
					config.DefaultDNSConfig(),
				)
			}
		})
	}
}
