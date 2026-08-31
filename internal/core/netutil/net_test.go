package netutil

import (
	"crypto/tls"
	"net"
	"testing"
)

func TestNormalizeHostWithSuffix(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"valid with path", "example.com/path", "example.com/path", false},
		{"valid with spaces", "  example.com/path  ", "example.com/path", false},
		{"valid with scheme and port", "https://example.com:443/index.html", "example.com/index.html", false},
		{"valid IDN", "münchen.de/path", "xn--mnchen-3ya.de/path", false},
		{"valid IP", "192.168.1.1/path", "192.168.1.1/path", false},
		{"valid localhost", "localhost/path", "localhost/path", false},
		{"valid query", "example.com?query=1", "example.com?query=1", false},
		{"invalid host trailing dot", "example.com./path", "", true},
		{"invalid host underscore", "invalid_host.com/path", "", true},
		{"invalid host label too long", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.com/path", "", true},
		{"invalid TLD length", "example.c/path", "", true}, // Fails regex: TLD must be 2-63 chars
		{"invalid single label", "com/path", "", true},     // Fails regex: re
		{"empty host", "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeHostWithSuffix(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NormalizeHostWithSuffix() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("NormalizeHostWithSuffix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractTLSServerName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"valid url", "https://example.com:443/path", "example.com", false},
		{"valid ip", "http://192.168.1.1", "192.168.1.1", false},
		{"invalid", "http://invalid_host", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractTLSServerName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractTLSServerName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ExtractTLSServerName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProtocolToScheme(t *testing.T) {
	tests := []struct {
		protocol string
		want     string
	}{
		{"https", "https://"},
		{"HTTPS", "https://"},
		{"https://", "https://"},
		{"http", "http://"},
		{"ftp", "http://"}, // Defaults to http
	}
	for _, tt := range tests {
		t.Run(tt.protocol, func(t *testing.T) {
			if got := ProtocolToScheme(tt.protocol); got != tt.want {
				t.Errorf("ProtocolToScheme() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsHTTPS(t *testing.T) {
	tests := []struct {
		protocol string
		want     bool
	}{
		{"https", true},
		{"HTTPS", true},
		{"https://", true},
		{"http", false},
		{"ftp", false},
	}
	for _, tt := range tests {
		t.Run(tt.protocol, func(t *testing.T) {
			if got := IsHTTPS(tt.protocol); got != tt.want {
				t.Errorf("IsHTTPS() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParsePortOrDefault(t *testing.T) {
	tests := []struct {
		port        int
		defaultPort uint16
		want        uint16
	}{
		{80, 443, 80},
		{0, 443, 0},
		{65535, 443, 65535},
		{-1, 443, 443},
		{65536, 443, 443},
	}
	for _, tt := range tests {
		if got := ParsePortOrDefault(tt.port, tt.defaultPort); got != tt.want {
			t.Errorf("ParsePortOrDefault(%d, %d) = %v, want %v", tt.port, tt.defaultPort, got, tt.want)
		}
	}
}

func TestParseTLSVersion(t *testing.T) {
	tests := []struct {
		v       string
		want    uint16
		wantErr bool
	}{
		{"tls1.0", tls.VersionTLS10, false},
		{"1.0", tls.VersionTLS10, false},
		{"tls1.1", tls.VersionTLS11, false},
		{"1.1", tls.VersionTLS11, false},
		{"tls1.2", tls.VersionTLS12, false},
		{"1.2", tls.VersionTLS12, false},
		{"tls1.3", tls.VersionTLS13, false},
		{"1.3", tls.VersionTLS13, false},
		{"invalid", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.v, func(t *testing.T) {
			got, err := ParseTLSVersion(tt.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTLSVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseTLSVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsPortAvailable_Available(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen() error = %v", err)
	}

	port := ln.Addr().(*net.TCPAddr).Port

	_ = ln.Close()

	if !IsPortAvailable(port) {
		t.Fatalf(
			"IsPortAvailable(%d) = false, want true",
			port,
		)
	}
}

func TestValidateDomain(t *testing.T) {
	tests := []struct {
		name    string
		domain  string
		wantErr bool
	}{
		{"valid domain", "example.com", false},
		{"valid subdomain", "sub.example.com", false},
		{"empty", "", true},
		{"with scheme", "http://example.com", true},
		{"with uppercase scheme", "HTTPS://example.com", true},
		{"with ftp scheme", "ftp://example.com", true},
		{"with path", "example.com/path", true},
		{"with port", "example.com:443", true},
		{"leading dot", ".example.com", true},
		{"trailing dot", "example.com.", true},
		{"consecutive dots", "example..com", true},
		{"single label", "example", true},
		{"label starting with hyphen", "-example.com", true},
		{"label ending with hyphen", "example-.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDomain(tt.domain)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDomain(%q) error = %v, wantErr %v", tt.domain, err, tt.wantErr)
			}
		})
	}
}

func TestHasScheme(t *testing.T) {
	tests := []struct {
		value string
		want  bool
	}{
		{"http://example.com", true},
		{"HTTPS://example.com", true},
		{"ftp://example.com", true},
		{"example.com", false},
		{"example.com/http://x", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			if got := HasScheme(tt.value); got != tt.want {
				t.Errorf("HasScheme(%q) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}
