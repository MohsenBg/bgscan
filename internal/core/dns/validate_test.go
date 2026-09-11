package dns

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/crypto/ssh"
)

// testPEMKey is a throwaway SSH private key generated once per test run.
var testPEMKey = func() string {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}

	block, err := ssh.MarshalPrivateKey(priv, "bgscan test")
	if err != nil {
		panic(err)
	}

	return string(pem.EncodeToMemory(block))
}()

func TestValidateProxyAuth(t *testing.T) {
	tests := []struct {
		name   string
		mut    func(*proxyAuthFields)
		field  string
		wantOK bool
	}{
		{
			name:   "valid socks no auth",
			wantOK: true,
		},
		{
			name: "socks with password auth",
			mut: func(f *proxyAuthFields) {
				f.authMethod = AuthPassword
				f.username = "user"
				f.password = "pass"
			},
			wantOK: true,
		},
		{
			name: "ssh with key auth",
			mut: func(f *proxyAuthFields) {
				f.proxyType = ResolverProxySSH
				f.authMethod = AuthKey
				f.username = "user"
				f.privateKey = testPEMKey
			},
			wantOK: true,
		},
		{
			name: "proxy port zero",
			mut: func(f *proxyAuthFields) {
				f.proxyPort = 0
			},
			field: "proxy_port",
		},
		{
			name: "ssh without auth",
			mut: func(f *proxyAuthFields) {
				f.proxyType = ResolverProxySSH
			},
			field: "auth_method",
		},
		{
			name: "socks with key auth",
			mut: func(f *proxyAuthFields) {
				f.authMethod = AuthKey
				f.username = "user"
				f.privateKey = testPEMKey
			},
			field: "auth_method",
		},
		{
			name: "unknown proxy type",
			mut: func(f *proxyAuthFields) {
				f.proxyType = "http"
			},
			field: "proxy_type",
		},
		{
			name: "password auth missing username",
			mut: func(f *proxyAuthFields) {
				f.authMethod = AuthPassword
				f.password = "pass"
			},
			field: "username",
		},
		{
			name: "password auth missing password",
			mut: func(f *proxyAuthFields) {
				f.authMethod = AuthPassword
				f.username = "user"
			},
			field: "password",
		},
		{
			name: "key auth missing username",
			mut: func(f *proxyAuthFields) {
				f.authMethod = AuthKey
				f.privateKey = testPEMKey
			},
			field: "username",
		},
		{
			name: "key auth invalid private key",
			mut: func(f *proxyAuthFields) {
				f.authMethod = AuthKey
				f.username = "user"
				f.privateKey = "not a key"
			},
			field: "private_key",
		},
		{
			name: "key auth missing known hosts file",
			mut: func(f *proxyAuthFields) {
				f.authMethod = AuthKey
				f.username = "user"
				f.privateKey = testPEMKey
				f.knownHostsFile = filepath.Join(t.TempDir(), "missing")
			},
			field: "known_hosts_file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fields := proxyAuthFields{
				proxyPort:  1080,
				proxyType:  ResolverProxySOCKS,
				authMethod: AuthNone,
			}
			if tt.mut != nil {
				tt.mut(&fields)
			}

			errs := validateProxyAuth(
				fields.proxyType,
				fields.proxyPort,
				fields.authMethod,
				fields.username,
				fields.password,
				fields.privateKey,
				fields.knownHostsFile,
			)

			if tt.wantOK {
				if len(errs) != 0 {
					t.Fatalf("Validate() = %v, want no errors", errs)
				}
				return
			}

			if _, ok := errs[tt.field]; !ok {
				t.Fatalf("Validate() = %v, want error for field %q", errs, tt.field)
			}
		})
	}
}

// proxyAuthFields mirrors validateProxyAuth's arguments so the test
// table above stays readable.
type proxyAuthFields struct {
	proxyType      ResolverProxyType
	proxyPort      uint16
	authMethod     AuthMethod
	username       string
	password       string
	privateKey     string
	knownHostsFile string
}

func TestValidateRPS(t *testing.T) {
	tests := []struct {
		rps    float64
		wantOK bool
	}{
		{0, true},
		{1, true},
		{500, true},
		{-1, false},
		{501, false},
	}

	for _, tt := range tests {
		err := validateRPS(tt.rps)

		if tt.wantOK && err != nil {
			t.Errorf("validateRPS(%v) = %v, want nil", tt.rps, err)
		}
		if !tt.wantOK && err == nil {
			t.Errorf("validateRPS(%v) = nil, want error", tt.rps)
		}
	}
}

func TestValidateFingerprint(t *testing.T) {
	if err := validateFingerprint(""); err == nil {
		t.Error("empty fingerprint should fail")
	}

	if err := validateFingerprint("definitely-invalid"); err == nil {
		t.Error("unknown fingerprint should fail")
	}

	if err := validateFingerprint("chrome"); err != nil {
		t.Errorf("case-insensitive fingerprint should pass: %v", err)
	}
}

func TestValidatePubKey(t *testing.T) {
	valid := strings.Repeat("a", 64)

	tests := []struct {
		key    string
		wantOK bool
	}{
		{valid, true},
		{"", false},
		{strings.Repeat("a", 63), false},
		{strings.Repeat("a", 65), false},
		{strings.Repeat("g", 64), false},
	}

	for _, tt := range tests {
		err := validatePubKey(tt.key)

		if tt.wantOK && err != nil {
			t.Errorf("validatePubKey(%q) = %v, want nil", tt.key, err)
		}
		if !tt.wantOK && err == nil {
			t.Errorf("validatePubKey(%q) = nil, want error", tt.key)
		}
	}
}

func TestValidateKnownHostsFile(t *testing.T) {
	if err := validateKnownHostsFile(""); err == nil {
		t.Error("empty path should fail")
	}

	if err := validateKnownHostsFile(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Error("missing file should fail")
	}

	if err := validateKnownHostsFile(t.TempDir()); err == nil {
		t.Error("directory should fail")
	}

	valid := filepath.Join(t.TempDir(), "known_hosts")
	if err := os.WriteFile(valid, nil, 0o600); err != nil {
		t.Fatalf("write known hosts: %v", err)
	}

	if err := validateKnownHostsFile(valid); err != nil {
		t.Errorf("empty known hosts file should pass: %v", err)
	}
}
