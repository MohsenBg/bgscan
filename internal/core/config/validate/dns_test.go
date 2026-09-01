package validate

import (
	"testing"
	"time"

	"github.com/MohsenBg/bgscan/internal/core/config"
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
				cfg.Resolver.Transport = "invalid"
				cfg.Resolver.Domain = ""
				cfg.Resolver.Port = MinDNSResolverPort - 1
				cfg.Resolver.CheckTypes = []string{}
				cfg.Resolver.Timeout = config.NewDurationMS(MinDNSResolverTimeout - 2*time.Millisecond)
				cfg.Resolver.Tries = MinDNSResolverTries - 1
				cfg.Resolver.DPI.Tries = MinDNSResolverDPITries - 1
				cfg.Resolver.DPI.Timeout = config.NewDurationMS(MinDNSResolverDPITimeout - 2*time.Millisecond)
				cfg.Resolver.OutputPrefix = ""
			},
			wantErrKeys: []string{
				"Resolver.Workers",
				"Resolver.Transport",
				"Resolver.Domain",
				"Resolver.Port",
				"Resolver.CheckTypes",
				"Resolver.Timeout",
				"Resolver.Tries",
				"Resolver.DPI.Tries",
				"Resolver.DPI.Timeout",
				"Resolver.OutputPrefix",
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

				if got.Resolver.Transport != want.Resolver.Transport {
					t.Errorf(
						"Resolver.Transport = %q, want %q",
						got.Resolver.Transport,
						want.Resolver.Transport,
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
				// DNSTunneling doesn't have a Domain field
			},
			wantErrKeys: []string{
				"Resolver.Domain",
			},
			wantWarnCount: 1,
			checkFixed: func(t *testing.T, got, want config.DNSConfig) {
				if got.Resolver.Domain != want.Resolver.Domain {
					t.Errorf(
						"Resolver.Domain = %q, want %q",
						got.Resolver.Domain,
						want.Resolver.Domain,
					)
				}
			},
		},
		{
			name: "invalid DNSTunneling fields",
			mutate: func(cfg *config.DNSConfig) {
				cfg.DNSTunneling.Workers = MinDNSTunnelingWorkers - 1
				cfg.DNSTunneling.Timeout = config.NewDurationMS(MinDNSTunnelingTimeout - 2*time.Millisecond)
				cfg.DNSTunneling.Tries = MinDNSResolverTries - 1
			},
			wantErrKeys: []string{
				"DNSTunneling.Workers",
				"DNSTunneling.Timeout",
				"DNSTunneling.Tries",
			},
			wantWarnCount: 3,
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
