package xray

import (
	"encoding/json"
	"net/netip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsProxyProtocol(t *testing.T) {
	tests := []struct {
		protocol Protocol
		want     bool
	}{
		{ProtocolVLESS, true},
		{ProtocolVMess, true},
		{ProtocolTrojan, true},
		{ProtocolShadowsocks, true},
		{ProtocolSocks, true},
		{ProtocolHTTP, true},
		{ProtocolWireGuard, true},
		{ProtocolHysteria, true},
		{Protocol("masque"), true},
		{ProtocolFreedom, false},
		{ProtocolBlackhole, false},
		{ProtocolDNS, false},
		{ProtocolLoopback, false},
		{"", false},
	}

	for _, tt := range tests {
		if got := isProxyProtocol(tt.protocol); got != tt.want {
			t.Errorf("isProxyProtocol(%q) = %v, want %v", tt.protocol, got, tt.want)
		}
	}
}

func TestSetOutboundAddress(t *testing.T) {
	ip := netip.MustParseAddr("8.8.8.8")

	tests := []struct {
		name string
		in   map[string]any
		port uint16
		want string // exact JSON of settings after replacement (keys sorted)
	}{
		{
			name: "flat vless address",
			in: map[string]any{
				"protocol": "vless",
				"settings": map[string]any{"address": "example.com", "port": 443},
			},
			want: `{"address":"8.8.8.8","port":443}`,
		},
		{
			name: "vnext vmess address",
			in: map[string]any{
				"protocol": "vmess",
				"settings": map[string]any{
					"vnext": []any{
						map[string]any{"address": "example.com", "port": 443},
					},
				},
			},
			want: `{"vnext":[{"address":"8.8.8.8","port":443}]}`,
		},
		{
			name: "servers trojan address",
			in: map[string]any{
				"protocol": "trojan",
				"settings": map[string]any{
					"servers": []any{
						map[string]any{"address": "example.com", "password": "x", "port": 443},
					},
				},
			},
			want: `{"servers":[{"address":"8.8.8.8","password":"x","port":443}]}`,
		},
		{
			name: "hysteria flat address",
			in: map[string]any{
				"protocol": "hysteria",
				"settings": map[string]any{"address": "example.com", "auth": "x", "port": 443},
			},
			want: `{"address":"8.8.8.8","auth":"x","port":443}`,
		},
		{
			name: "xray address drops embedded port",
			in: map[string]any{
				"protocol": "vless",
				"settings": map[string]any{"address": "example.com:8080"},
			},
			want: `{"address":"8.8.8.8"}`,
		},
		{
			name: "flat vless with override port",
			in: map[string]any{
				"protocol": "vless",
				"settings": map[string]any{"address": "example.com", "port": 443},
			},
			port: 8443,
			want: `{"address":"8.8.8.8","port":8443}`,
		},
		{
			name: "override port wins over embedded template port",
			in: map[string]any{
				"protocol": "vless",
				"settings": map[string]any{"address": "example.com:8080"},
			},
			port: 8443,
			want: `{"address":"8.8.8.8","port":8443}`,
		},
		{
			name: "wireguard endpoint keeps port",
			in: map[string]any{
				"protocol": "wireguard",
				"settings": map[string]any{
					"address": []any{"172.16.0.2/32"},
					"peers": []any{
						map[string]any{"publicKey": "abc", "endpoint": "engage.cloudflareclient.com:2408"},
					},
				},
			},
			want: `{"address":["172.16.0.2/32"],"peers":[{"endpoint":"8.8.8.8:2408","publicKey":"abc"}]}`,
		},
		{
			name: "wireguard endpoint override port",
			in: map[string]any{
				"protocol": "wireguard",
				"settings": map[string]any{
					"address": []any{"172.16.0.2/32"},
					"peers": []any{
						map[string]any{"publicKey": "abc", "endpoint": "engage.cloudflareclient.com:2408"},
					},
				},
			},
			port: 51820,
			want: `{"address":["172.16.0.2/32"],"peers":[{"endpoint":"8.8.8.8:51820","publicKey":"abc"}]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := setOutboundAddress(tt.in, ip, tt.port); err != nil {
				t.Fatalf("setOutboundAddress() error = %v", err)
			}

			got, err := json.Marshal(tt.in["settings"])
			if err != nil {
				t.Fatalf("json.Marshal() error = %v", err)
			}
			if string(got) != tt.want {
				t.Fatalf("settings = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestSetOutboundAddress_Errors(t *testing.T) {
	ip := netip.MustParseAddr("8.8.8.8")

	tests := []struct {
		name        string
		in          map[string]any
		invalidIP   bool
		errContains string
	}{
		{
			name:        "freedom is not a proxy",
			in:          map[string]any{"protocol": "freedom", "settings": map[string]any{"address": "1.1.1.1"}},
			errContains: "not a proxy",
		},
		{
			name:        "dns outbound untouched",
			in:          map[string]any{"protocol": "dns", "settings": map[string]any{"address": "1.1.1.1"}},
			errContains: "not a proxy",
		},
		{
			name:        "missing settings",
			in:          map[string]any{"protocol": "vless"},
			errContains: "no settings",
		},
		{
			name:        "no address at all",
			in:          map[string]any{"protocol": "vless", "settings": map[string]any{"port": 443}},
			errContains: "no replaceable",
		},
		{
			name:        "non-string address is not the remote",
			in:          map[string]any{"protocol": "vless", "settings": map[string]any{"address": []any{"172.16.0.2/32"}}},
			errContains: "no replaceable",
		},
		{
			name:        "wireguard without peers",
			in:          map[string]any{"protocol": "wireguard", "settings": map[string]any{"address": []any{"172.16.0.2/32"}}},
			errContains: "no peer endpoint",
		},
		{
			name:        "invalid IP",
			in:          map[string]any{"protocol": "vless", "settings": map[string]any{"address": "example.com"}},
			invalidIP:   true,
			errContains: "invalid IP",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := ip
			if tt.invalidIP {
				target = netip.Addr{}
			}

			err := setOutboundAddress(tt.in, target, 0)
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tt.errContains) {
				t.Fatalf("error = %v, want message containing %q", err, tt.errContains)
			}
		})
	}
}

func TestRequireReplaceableAddress(t *testing.T) {
	proxy := map[string]any{"protocol": "vless", "settings": map[string]any{"address": "example.com"}}
	if err := requireReplaceableAddress(proxy); err != nil {
		t.Fatalf("proxy outbound rejected: %v", err)
	}

	nonProxy := map[string]any{"protocol": "freedom", "settings": map[string]any{}}
	if err := requireReplaceableAddress(nonProxy); err == nil {
		t.Fatal("non-proxy outbound accepted")
	}
}

func TestApplyOutboundTemplate(t *testing.T) {
	tests := []struct {
		name     string
		tmpl     map[string]any
		port     uint16
		wantAddr string // expected address JSON fragment
		wantPort string // expected "port":N JSON fragment
	}{
		{
			name: "vnext vmess",
			tmpl: map[string]any{
				"protocol": "vmess",
				"settings": map[string]any{
					"vnext": []any{
						map[string]any{"address": "example.com", "port": 443},
					},
				},
			},
			wantAddr: `"address":"8.8.8.8"`,
			wantPort: `"port":443`,
		},
		{
			name: "flat vless",
			tmpl: map[string]any{
				"protocol": "vless",
				"settings": map[string]any{"address": "example.com", "port": 443},
			},
			wantAddr: `"address":"8.8.8.8"`,
			wantPort: `"port":443`,
		},
		{
			name: "flat vless override port",
			tmpl: map[string]any{
				"protocol": "vless",
				"settings": map[string]any{"address": "example.com", "port": 443},
			},
			port:     8443,
			wantAddr: `"address":"8.8.8.8"`,
			wantPort: `"port":8443`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "template.json")
			writeJSONFile(t, path, tt.tmpl)

			got, err := applyOutboundTemplate(path, netip.MustParseAddr("8.8.8.8"), tt.port)
			if err != nil {
				t.Fatalf("applyOutboundTemplate() error = %v", err)
			}

			root, ok := got.(map[string]any)
			if !ok {
				t.Fatalf("applyOutboundTemplate() returned %T, want map[string]any", got)
			}

			raw, err := json.Marshal(root["settings"])
			if err != nil {
				t.Fatalf("json.Marshal() error = %v", err)
			}
			if !strings.Contains(string(raw), tt.wantAddr) {
				t.Fatalf("address not replaced: got %s, want %q", raw, tt.wantAddr)
			}
			if !strings.Contains(string(raw), tt.wantPort) {
				t.Fatalf("port wrong: got %s, want %q", raw, tt.wantPort)
			}
		})
	}
}

func TestApplyOutboundTemplate_InvalidIP(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "template.json")
	writeJSONFile(t, path, map[string]any{
		"protocol": "vless",
		"settings": map[string]any{"address": "example.com"},
	})

	// The zero netip.Addr is the only invalid value the type can carry.
	_, err := applyOutboundTemplate(path, netip.Addr{}, 0)
	if err == nil {
		t.Fatal("expected invalid IP error")
	}
	if !strings.Contains(err.Error(), "invalid IP") {
		t.Fatalf("error = %v, want invalid IP message", err)
	}
}

func TestApplyOutboundTemplate_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")

	if err := os.WriteFile(path, []byte(`{"protocol":`), 0o644); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	_, err := applyOutboundTemplate(path, netip.MustParseAddr("1.1.1.1"), 0)
	if err == nil {
		t.Fatal("expected JSON parse error")
	}
}

func TestNormalizeTemplateName(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"node", "node.json"},
		{"node.json", "node.json"},
		{"node.txt", "node.json"},
		{"dir/node", "node.json"},
		{"../sneaky/x", "x.json"},
	}

	for _, tt := range tests {
		got := normalizeTemplateName(tt.in)
		if got != tt.want {
			t.Fatalf("normalizeTemplateName(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestLoadOutboundFileMetadata(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.json")
	writeJSONFile(t, path, map[string]any{
		"protocol": "vless",
		"streamSettings": map[string]any{
			"network":  "ws",
			"security": "tls",
		},
	})

	got, err := loadOutboundFileMetadata(path)
	if err != nil {
		t.Fatalf("loadOutboundFileMetadata() error = %v", err)
	}

	if got.Name != "sample" {
		t.Fatalf("Name = %q, want sample", got.Name)
	}
	if got.Path != path {
		t.Fatalf("Path = %q, want %q", got.Path, path)
	}
	if got.Protocol != "vless" {
		t.Fatalf("Protocol = %q, want vless", got.Protocol)
	}
	if got.Network != "ws" {
		t.Fatalf("Network = %q, want ws", got.Network)
	}
	if got.Security != SecurityTLS {
		t.Fatalf("Security = %q, want tls", got.Security)
	}
	if got.CreatedTime.IsZero() {
		t.Fatal("CreatedTime is zero")
	}
}

func TestSaveOutboundFromFile_Errors(t *testing.T) {
	t.Run("missing source", func(t *testing.T) {
		_, err := SaveOutboundFromFile(nil, filepath.Join(t.TempDir(), "missing.json"), "node")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("source is directory", func(t *testing.T) {
		dir := t.TempDir()
		_, err := SaveOutboundFromFile(nil, dir, "node")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		dir := t.TempDir()
		src := filepath.Join(dir, "bad.json")
		if err := os.WriteFile(src, []byte(`{`), 0o644); err != nil {
			t.Fatalf("os.WriteFile() error = %v", err)
		}

		_, err := SaveOutboundFromFile(nil, src, "node")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("non-proxy outbound rejected", func(t *testing.T) {
		dir := t.TempDir()
		src := filepath.Join(dir, "freedom.json")
		writeJSONFile(t, src, map[string]any{
			"protocol": "freedom",
		})

		_, err := SaveOutboundFromFile(nil, src, "node")
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "not a proxy") {
			t.Fatalf("error = %v, want not-a-proxy message", err)
		}
	})

	t.Run("proxy without remote address rejected", func(t *testing.T) {
		dir := t.TempDir()
		src := filepath.Join(dir, "vless.json")
		writeJSONFile(t, src, map[string]any{
			"protocol": "vless",
			"settings": map[string]any{"port": 443},
		})

		_, err := SaveOutboundFromFile(nil, src, "node")
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "no replaceable") {
			t.Fatalf("error = %v, want no-replaceable message", err)
		}
	})
}

func writeJSONFile(t *testing.T, path string, v any) {
	t.Helper()

	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}
}
