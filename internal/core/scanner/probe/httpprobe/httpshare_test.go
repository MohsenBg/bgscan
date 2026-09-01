package httpprobe

import (
	"crypto/tls"
	"testing"
	"time"

	"github.com/MohsenBg/bgscan/internal/core/config"
)

// --- statusFilter ---

func TestNewStatusFilter_Empty(t *testing.T) {
	f := newStatusFilter(nil, totalHTTPStatusCodes)
	if len(f.accepted) != 0 {
		t.Errorf("expected empty filter for nil codes, got %d entries", len(f.accepted))
	}
}

// TestNewStatusFilter_CoversAll verifies that a filter defaults to accepting
// all codes when the provided list size equals or exceeds the total.
func TestNewStatusFilter_CoversAll(t *testing.T) {
	codes := make([]int, totalHTTPStatusCodes)
	for i := range codes {
		codes[i] = 100 + i
	}
	f := newStatusFilter(codes, totalHTTPStatusCodes)
	if len(f.accepted) != 0 {
		t.Errorf("expected empty filter when codes >= total, got %d entries", len(f.accepted))
	}
}

func TestNewStatusFilter_Subset(t *testing.T) {
	f := newStatusFilter([]int{200, 301, 404}, totalHTTPStatusCodes)
	if len(f.accepted) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(f.accepted))
	}
	for _, code := range []int{200, 301, 404} {
		if _, ok := f.accepted[code]; !ok {
			t.Errorf("code %d missing from filter", code)
		}
	}
}

func TestStatusFilter_IsAccepted(t *testing.T) {
	tests := []struct {
		name  string
		codes []int
		check int
		want  bool
	}{
		{"empty filter accepts anything", nil, 500, true},
		{"empty filter accepts 200", nil, 200, true},
		{"subset accepts listed", []int{200, 201, 204}, 200, true},
		{"subset accepts listed 2", []int{200, 201, 204}, 204, true},
		{"subset rejects unlisted", []int{200, 201, 204}, 404, false},
		{"subset rejects 500", []int{200, 301}, 500, false},
		{"full coverage accepts all", make([]int, totalHTTPStatusCodes), 999, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newStatusFilter(tt.codes, totalHTTPStatusCodes)
			if got := f.isAccepted(tt.check); got != tt.want {
				t.Errorf("isAccepted(%d) = %v, want %v", tt.check, got, tt.want)
			}
		})
	}
}

// --- newTLSConfig ---

func TestNewTLSConfig_NoTLS(t *testing.T) {
	req := HTTPRequest{UseTLS: false}
	if cfg := newTLSConfig(req); cfg != nil {
		t.Errorf("expected nil for non-TLS request, got %+v", cfg)
	}
}

func TestNewTLSConfig_WithTLS(t *testing.T) {
	req := HTTPRequest{
		UseTLS:        true,
		SNI:           "example.com",
		SkipTLSVerify: true,
		MinTLSVersion: tls.VersionTLS12,
		MaxTLSVersion: tls.VersionTLS13,
	}

	cfg := newTLSConfig(req)
	if cfg == nil {
		t.Fatal("expected non-nil tls.Config for TLS request")
	}
	if cfg.ServerName != "example.com" {
		t.Errorf("ServerName = %q, want %q", cfg.ServerName, "example.com")
	}
	if !cfg.InsecureSkipVerify {
		t.Error("InsecureSkipVerify = false, want true")
	}
	if cfg.MinVersion != tls.VersionTLS12 {
		t.Errorf("MinVersion = %x, want %x", cfg.MinVersion, tls.VersionTLS12)
	}
	if cfg.MaxVersion != tls.VersionTLS13 {
		t.Errorf("MaxVersion = %x, want %x", cfg.MaxVersion, tls.VersionTLS13)
	}
}

func TestNewTLSConfig_NoSkipVerify(t *testing.T) {
	req := HTTPRequest{
		UseTLS:        true,
		SNI:           "secure.io",
		SkipTLSVerify: false,
	}

	cfg := newTLSConfig(req)
	if cfg.InsecureSkipVerify {
		t.Error("InsecureSkipVerify = true, want false")
	}
}

// --- defaultPort ---

func TestDefaultPort(t *testing.T) {
	tests := []struct {
		name    string
		port    int
		useTLS  bool
		want    uint16
		wantErr bool
	}{
		{"explicit port", 8080, false, 8080, false},
		{"explicit port with TLS", 8443, true, 8443, false},
		{"zero port no TLS → 80", 0, false, 80, false},
		{"zero port TLS → 443", 0, true, 443, false},
		{"port 1", 1, false, 1, false},
		{"max port", 65535, false, 65535, false},
		{"negative port", -1, false, 0, true},
		{"too large port", 65536, false, 0, true},
		{"way too large", 100000, true, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := defaultPort(tt.port, tt.useTLS)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got port %d", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("defaultPort(%d, %v) = %d, want %d", tt.port, tt.useTLS, got, tt.want)
			}
		})
	}
}

// --- resolveSNI ---

// TestResolveSNI verifies that explicit SNI is preserved and empty SNI
// gracefully falls back without error, leaving derivation to the caller.
func TestResolveSNI(t *testing.T) {
	tests := []struct {
		name       string
		serverName string
		useTLS     bool
		want       string
		wantErr    bool
	}{
		{"explicit SNI", "cdn.example.com", true, "cdn.example.com", false},
		{"explicit SNI no TLS", "cdn.example.com", false, "cdn.example.com", false},
		{"empty SNI with TLS", "", true, "", false},
		{"empty SNI no TLS", "", false, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveSNI(tt.serverName, tt.useTLS)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("resolveSNI(%q, %v) = %q, want %q", tt.serverName, tt.useTLS, got, tt.want)
			}
		})
	}
}

// --- resolveHTTPVersion ---

// TestResolveHTTPVersion ensures case-insensitive matching and correct
// fallback to H1H2 for unrecognized or mixed protocol strings.
func TestResolveHTTPVersion(t *testing.T) {
	tests := []struct {
		input string
		want  HTTPVersion
	}{
		// HTTP/1 variants
		{"h1", HTTPVersionH1},
		{"H1", HTTPVersionH1},
		{"http/1", HTTPVersionH1},
		{"HTTP/1", HTTPVersionH1},
		{"http/1.1", HTTPVersionH1},
		{"HTTP/1.1", HTTPVersionH1},
		{"http1", HTTPVersionH1},

		// HTTP/2 variants
		{"h2", HTTPVersionH2},
		{"H2", HTTPVersionH2},
		{"http/2", HTTPVersionH2},
		{"HTTP/2", HTTPVersionH2},
		{"http2", HTTPVersionH2},

		// Default / H1H2
		{"", HTTPVersionH1H2},
		{"h1,h2", HTTPVersionH1H2},
		{"h3", HTTPVersionH1H2},
		{"quic", HTTPVersionH1H2},
		{"unknown", HTTPVersionH1H2},

		// Whitespace handling
		{"  h1  ", HTTPVersionH1},
		{" h2 ", HTTPVersionH2},
		{"  ", HTTPVersionH1H2},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := resolveHTTPVersion(tt.input); got != tt.want {
				t.Errorf("resolveHTTPVersion(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

// --- resolveTLSVersions ---

func TestResolveTLSVersions_Valid(t *testing.T) {
	cfg := config.HTTPConfig{
		MinTLSVersion: "tls1.2",
		MaxTLSVersion: "tls1.3",
	}

	minV, maxV, err := resolveTLSVersions(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if minV != tls.VersionTLS12 {
		t.Errorf("min = %x, want TLS1.2 (%x)", minV, tls.VersionTLS12)
	}
	if maxV != tls.VersionTLS13 {
		t.Errorf("max = %x, want TLS1.3 (%x)", maxV, tls.VersionTLS13)
	}
}

func TestResolveTLSVersions_Equal(t *testing.T) {
	cfg := config.HTTPConfig{
		MinTLSVersion: "tls1.2",
		MaxTLSVersion: "tls1.2",
	}

	minV, maxV, err := resolveTLSVersions(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if minV != maxV {
		t.Errorf("min (%x) != max (%x), want equal", minV, maxV)
	}
}

// TestResolveTLSVersions_MinGreaterThanMax ensures validation catches
// inverted TLS version constraints.
func TestResolveTLSVersions_MinGreaterThanMax(t *testing.T) {
	cfg := config.HTTPConfig{
		MinTLSVersion: "tls1.3",
		MaxTLSVersion: "tls1.1",
	}

	_, _, err := resolveTLSVersions(cfg)
	if err == nil {
		t.Fatal("expected error when min > max")
	}
}

func TestResolveTLSVersions_InvalidMin(t *testing.T) {
	cfg := config.HTTPConfig{
		MinTLSVersion: "ssl3.0",
		MaxTLSVersion: "tls1.3",
	}

	_, _, err := resolveTLSVersions(cfg)
	if err == nil {
		t.Fatal("expected error for invalid min TLS version")
	}
}

func TestResolveTLSVersions_InvalidMax(t *testing.T) {
	cfg := config.HTTPConfig{
		MinTLSVersion: "tls1.2",
		MaxTLSVersion: "garbage",
	}

	_, _, err := resolveTLSVersions(cfg)
	if err == nil {
		t.Fatal("expected error for invalid max TLS version")
	}
}

// --- NewHTTPRequestFromConfig ---

func TestNewHTTPRequestFromConfig_HTTPS(t *testing.T) {
	cfg := config.HTTPConfig{
		Host:          "example.com",
		Port:          443,
		Protocol:      "https",
		Version:       "h2",
		TLSValidation: true,
		MinTLSVersion: "tls1.2",
		MaxTLSVersion: "tls1.3",
		Timeout:       config.NewDurationMS(5 * time.Second),
	}

	req, err := NewHTTPRequestFromConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !req.UseTLS {
		t.Error("UseTLS = false, want true for https")
	}
	if req.Version != HTTPVersionH2 {
		t.Errorf("Version = %d, want HTTPVersionH2", req.Version)
	}
	if req.SkipTLSVerify {
		t.Error("SkipTLSVerify = true, want false (TLSValidation=true)")
	}
	if req.Timeout != 5*time.Second {
		t.Errorf("Timeout = %v, want 5s", req.Timeout)
	}
	if req.SNI != "example.com" {
		t.Errorf("SNI = %q, want %q", req.SNI, "example.com")
	}
	if req.Host != "example.com" {
		t.Errorf("Host = %q, want %q", req.Host, "example.com")
	}
	if req.MinTLSVersion != tls.VersionTLS12 {
		t.Errorf("MinTLSVersion = %x, want %x", req.MinTLSVersion, tls.VersionTLS12)
	}
	if req.MaxTLSVersion != tls.VersionTLS13 {
		t.Errorf("MaxTLSVersion = %x, want %x", req.MaxTLSVersion, tls.VersionTLS13)
	}
}

func TestNewHTTPRequestFromConfig_HTTP(t *testing.T) {
	cfg := config.HTTPConfig{
		Host:          "plain.io",
		Port:          8080,
		Protocol:      "http",
		Version:       "h1",
		TLSValidation: false,
		MinTLSVersion: "tls1.0",
		MaxTLSVersion: "tls1.3",
		Timeout:       config.NewDurationMS(3 * time.Second),
	}

	req, err := NewHTTPRequestFromConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if req.UseTLS {
		t.Error("UseTLS = true, want false for http")
	}
	if req.Version != HTTPVersionH1 {
		t.Errorf("Version = %d, want HTTPVersionH1", req.Version)
	}
	if req.Timeout != 3*time.Second {
		t.Errorf("Timeout = %v, want 3s", req.Timeout)
	}
}

// TestNewHTTPRequestFromConfig_DefaultPort_HTTPS verifies that a zero port
// correctly defaults to 443 in the generated URL for HTTPS.
func TestNewHTTPRequestFromConfig_DefaultPort_HTTPS(t *testing.T) {
	cfg := config.HTTPConfig{
		Host:          "secure.io",
		Port:          0,
		Protocol:      "https",
		TLSValidation: true,
		MinTLSVersion: "tls1.2",
		MaxTLSVersion: "tls1.3",
		Timeout:       config.NewDurationMS(4 * time.Second),
	}

	req, err := NewHTTPRequestFromConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if want := "https://secure.io:443"; req.URL != want {
		t.Errorf("URL = %q, want %q", req.URL, want)
	}
}

// TestNewHTTPRequestFromConfig_DefaultPort_HTTP verifies that a zero port
// correctly defaults to 80 in the generated URL for plain HTTP.
func TestNewHTTPRequestFromConfig_DefaultPort_HTTP(t *testing.T) {
	cfg := config.HTTPConfig{
		Host:          "plain.io",
		Port:          0,
		Protocol:      "http",
		MinTLSVersion: "tls1.0",
		MaxTLSVersion: "tls1.3",
		Timeout:       config.NewDurationMS(4 * time.Second),
	}

	req, err := NewHTTPRequestFromConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if want := "http://plain.io:80"; req.URL != want {
		t.Errorf("URL = %q, want %q", req.URL, want)
	}
}

func TestNewHTTPRequestFromConfig_CustomServerName(t *testing.T) {
	cfg := config.HTTPConfig{
		Host:          "1.2.3.4",
		ServerName:    "real.example.com",
		Port:          443,
		Protocol:      "https",
		TLSValidation: true,
		MinTLSVersion: "tls1.2",
		MaxTLSVersion: "tls1.3",
		Timeout:       config.NewDurationMS(4 * time.Second),
	}

	req, err := NewHTTPRequestFromConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if req.SNI != "real.example.com" {
		t.Errorf("SNI = %q, want %q", req.SNI, "real.example.com")
	}
}

// TestNewHTTPRequestFromConfig_SkipVerify confirms that disabling TLSValidation
// correctly sets SkipTLSVerify to true.
func TestNewHTTPRequestFromConfig_SkipVerify(t *testing.T) {
	cfg := config.HTTPConfig{
		Host:          "self-signed.local",
		Port:          443,
		Protocol:      "https",
		TLSValidation: false,
		MinTLSVersion: "tls1.2",
		MaxTLSVersion: "tls1.3",
		Timeout:       config.NewDurationMS(4 * time.Second),
	}

	req, err := NewHTTPRequestFromConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !req.SkipTLSVerify {
		t.Error("SkipTLSVerify = false, want true when TLSValidation=false")
	}
}

func TestNewHTTPRequestFromConfig_InvalidPort(t *testing.T) {
	cfg := config.HTTPConfig{
		Host:          "example.com",
		Port:          -1,
		Protocol:      "https",
		MinTLSVersion: "tls1.2",
		MaxTLSVersion: "tls1.3",
		Timeout:       config.NewDurationMS(4 * time.Second),
	}

	_, err := NewHTTPRequestFromConfig(cfg)
	if err == nil {
		t.Fatal("expected error for invalid port")
	}
}

func TestNewHTTPRequestFromConfig_InvalidTLSVersions(t *testing.T) {
	cfg := config.HTTPConfig{
		Host:          "example.com",
		Port:          443,
		Protocol:      "https",
		MinTLSVersion: "tls1.3",
		MaxTLSVersion: "tls1.0",
		Timeout:       config.NewDurationMS(4 * time.Second),
	}

	_, err := NewHTTPRequestFromConfig(cfg)
	if err == nil {
		t.Fatal("expected error when min TLS > max TLS")
	}
}

// --- isHTTPS (assumed helper) ---

func TestIsHTTPS(t *testing.T) {
	tests := []struct {
		protocol string
		want     bool
	}{
		{"https", true},
		{"HTTPS", true},
		{"http", false},
		{"HTTP", false},
		{"h2", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.protocol, func(t *testing.T) {
			if got := isHTTPS(tt.protocol); got != tt.want {
				t.Errorf("isHTTPS(%q) = %v, want %v", tt.protocol, got, tt.want)
			}
		})
	}
}
