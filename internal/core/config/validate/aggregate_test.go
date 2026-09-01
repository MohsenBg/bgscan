package validate

import (
	"errors"
	"testing"

	"github.com/MohsenBg/bgscan/internal/core/config"
)

func defaultScannerConfig() config.ScannerConfig {
	return config.ScannerConfig{
		General: config.DefaultGeneralConfig(),
		Writer:  config.DefaultWriterConfig(),
		ICMP:    config.DefaultICMPConfig(),
		TCP:     config.DefaultTCPConfig(),
		HTTP:    config.DefaultHTTPConfig(),
		Xray:    config.DefaultXrayConfig(),
		DNS:     config.DefaultDNSConfig(),
	}
}

func TestAllWarnings_HasWarnings(t *testing.T) {
	tests := []struct {
		name     string
		warnings AllWarnings
		want     bool
	}{
		{
			name: "no warnings",
		},
		{
			name: "general warning",
			warnings: AllWarnings{
				General: make([]Warning, 1),
			},
			want: true,
		},
		{
			name: "writer warning",
			warnings: AllWarnings{
				Writer: make([]Warning, 1),
			},
			want: true,
		},
		{
			name: "icmp warning",
			warnings: AllWarnings{
				ICMP: make([]Warning, 1),
			},
			want: true,
		},
		{
			name: "tcp warning",
			warnings: AllWarnings{
				TCP: make([]Warning, 1),
			},
			want: true,
		},
		{
			name: "http warning",
			warnings: AllWarnings{
				HTTP: make([]Warning, 1),
			},
			want: true,
		},
		{
			name: "xray warning",
			warnings: AllWarnings{
				Xray: make([]Warning, 1),
			},
			want: true,
		},
		{
			name: "dns warning",
			warnings: AllWarnings{
				DNS: make([]Warning, 1),
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.warnings.HasWarnings(); got != tt.want {
				t.Fatalf("HasWarnings() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAllErrors_HasErrors(t *testing.T) {
	tests := []struct {
		name string
		errs AllErrors
		want bool
	}{
		{
			name: "no errors",
		},
		{
			name: "general error",
			errs: AllErrors{
				General: map[string]error{"field": errors.New("invalid")},
			},
			want: true,
		},
		{
			name: "writer error",
			errs: AllErrors{
				Writer: map[string]error{"field": errors.New("invalid")},
			},
			want: true,
		},
		{
			name: "icmp error",
			errs: AllErrors{
				ICMP: map[string]error{"field": errors.New("invalid")},
			},
			want: true,
		},
		{
			name: "tcp error",
			errs: AllErrors{
				TCP: map[string]error{"field": errors.New("invalid")},
			},
			want: true,
		},
		{
			name: "http error",
			errs: AllErrors{
				HTTP: map[string]error{"field": errors.New("invalid")},
			},
			want: true,
		},
		{
			name: "xray error",
			errs: AllErrors{
				Xray: map[string]error{"field": errors.New("invalid")},
			},
			want: true,
		},
		{
			name: "dns error",
			errs: AllErrors{
				DNS: map[string]error{"field": errors.New("invalid")},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.errs.HasErrors(); got != tt.want {
				t.Fatalf("HasErrors() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNormalizeAll_DefaultConfigHasNoWarnings(t *testing.T) {
	cfg := defaultScannerConfig()

	warnings := NormalizeAll(&cfg)

	if warnings.HasWarnings() {
		t.Fatalf("default config produced warnings: %+v", warnings)
	}
}

func TestValidateAll_DefaultConfigHasNoErrors(t *testing.T) {
	cfg := defaultScannerConfig()

	errs := ValidateAll(cfg)

	if errs.HasErrors() {
		t.Fatalf("default config produced validation errors: %+v", errs)
	}
}
