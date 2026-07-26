package xray

import (
	"encoding/base64"
	"encoding/json"
	"net/url"
	"reflect"
	"testing"
)

func TestDefaultPort(t *testing.T) {
	tests := []struct {
		name     string
		in       string
		fallback int
		want     int
	}{
		{"explicit", "8443", 443, 8443},
		{"empty uses fallback", "", 443, 443},
		{"invalid uses fallback", "bad", 443, 443},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := defaultPort(tt.in, tt.fallback)
			if got != tt.want {
				t.Fatalf("defaultPort(%q, %d) = %d, want %d", tt.in, tt.fallback, got, tt.want)
			}
		})
	}
}

func TestSplitComma(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want []string
	}{
		{"empty", "", nil},
		{"single", "a", []string{"a"}},
		{"multi", "a,b,c", []string{"a", "b", "c"}},
		{"trim", " a, b , c ", []string{"a", "b", "c"}},
		{"skip empty", "a,,c", []string{"a", "c"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitComma(tt.in)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("splitComma(%q) = %#v, want %#v", tt.in, got, tt.want)
			}
		})
	}
}

func TestNum(t *testing.T) {
	tests := []struct {
		name string
		in   any
		want int
	}{
		{"string", "443", 443},
		{"int", 80, 80},
		{"float64", float64(8080), 8080},
		{"bad string", "abc", 0},
		{"nil", nil, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := num(tt.in)
			if got != tt.want {
				t.Fatalf("num(%v) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}

func TestCanonicalQuery(t *testing.T) {
	v := url.Values{}
	v.Set("b", "2")
	v.Set("a", "1")
	v.Add("a", "0")

	got := canonicalQuery(v)
	want := "a=1&a=0&b=2"

	if got != want {
		t.Fatalf("canonicalQuery() = %q, want %q", got, want)
	}
}

func TestSplitMethodPass(t *testing.T) {
	tests := []struct {
		name       string
		in         string
		wantMethod string
		wantPass   string
	}{
		{"normal", "aes-256-gcm:secret", "aes-256-gcm", "secret"},
		{"fallback", "secret-only", "2022-blake3-aes-128-gcm", "secret-only"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			method, pass := splitMethodPass(tt.in)
			if method != tt.wantMethod || pass != tt.wantPass {
				t.Fatalf("splitMethodPass(%q) = (%q, %q), want (%q, %q)",
					tt.in, method, pass, tt.wantMethod, tt.wantPass)
			}
		})
	}
}

func TestBase64DecodeFlexible(t *testing.T) {
	raw := `{"v":"2","ps":"test","add":"example.com","port":"443","id":"uuid","aid":"0","net":"ws","type":"none","host":"example.com","path":"/ws","tls":"tls"}`
	std := base64.StdEncoding.EncodeToString([]byte(raw))
	urlSafe := base64.URLEncoding.EncodeToString([]byte(raw))

	tests := []struct {
		name string
		in   string
	}{
		{"std", std},
		{"url-safe", urlSafe},
		{"std-no-padding", trimBase64Padding(std)},
		{"url-safe-no-padding", trimBase64Padding(urlSafe)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := base64DecodeFlexible(tt.in)
			if err != nil {
				t.Fatalf("base64DecodeFlexible() error = %v", err)
			}
			if string(got) != raw {
				t.Fatalf("decoded mismatch\ngot:  %s\nwant: %s", string(got), raw)
			}
		})
	}
}

func trimBase64Padding(s string) string {
	for len(s) > 0 && s[len(s)-1] == '=' {
		s = s[:len(s)-1]
	}
	return s
}

func TestCanonicalQuery_SortsKeys(t *testing.T) {
	in, _ := url.ParseQuery("b=2&a=1&c=3")
	got := canonicalQuery(in)

	want := "a=1&b=2&c=3"
	if got != want {
		t.Fatalf("canonicalQuery(%q) = %q, want %q", in, got, want)
	}
}

func TestParseLink_UnsupportedScheme(t *testing.T) {
	_, err := ParseLink("ftp://example.com")
	if err == nil {
		t.Fatal("expected error for unsupported scheme, got nil")
	}
}

func TestParseVmess(t *testing.T) {
	j := map[string]any{
		"v":    "2",
		"ps":   "vmess-node",
		"add":  "example.com",
		"port": "443",
		"id":   "11111111-1111-1111-1111-111111111111",
		"net":  "ws",
		"host": "ws.example.com",
		"path": "/ws",
		"tls":  "tls",
		"scy":  "auto",
	}

	raw, err := json.Marshal(j)
	if err != nil {
		t.Fatalf("json.Marshal error: %v", err)
	}

	link := "vmess://" + base64.StdEncoding.EncodeToString(raw)

	got, err := parseVmess(link)
	if err != nil {
		t.Fatalf("parseVmess error = %v", err)
	}
	if got == nil {
		t.Fatal("parseVmess returned nil")
	}

	if got.Identity == "" {
		t.Fatal("Identity is empty")
	}

	if got.Outbound["protocol"] != "vmess" {
		t.Fatalf("protocol = %v, want vmess", got.Outbound["protocol"])
	}
	if got.Outbound["tag"] != "vmess-node" {
		t.Fatalf("tag = %v, want vmess-node", got.Outbound["tag"])
	}

	settings := got.Outbound["settings"].(map[string]any)
	vnext := settings["vnext"].([]any)
	server := vnext[0].(map[string]any)

	if server["address"] != addressPlaceholder {
		t.Fatalf("address = %v, want %v", server["address"], addressPlaceholder)
	}
	if server["port"] != 443 {
		t.Fatalf("port = %v, want 443", server["port"])
	}

	users := server["users"].([]any)
	user := users[0].(map[string]any)

	if user["id"] != "11111111-1111-1111-1111-111111111111" {
		t.Fatalf("id = %v", user["id"])
	}
	if user["security"] != "auto" {
		t.Fatalf("security = %v, want auto", user["security"])
	}

	stream := got.Outbound["streamSettings"].(map[string]any)
	if stream["network"] != "ws" {
		t.Fatalf("network = %v, want ws", stream["network"])
	}
	if stream["security"] != "tls" {
		t.Fatalf("security = %v, want tls", stream["security"])
	}
}

func TestVmessIdentity_IgnoresPS(t *testing.T) {
	j1 := map[string]any{
		"add":  "example.com",
		"port": "443",
		"id":   "uuid",
		"ps":   "node-1",
	}
	j2 := map[string]any{
		"add":  "example.com",
		"port": "443",
		"id":   "uuid",
		"ps":   "node-2",
	}

	id1 := vmessIdentity(j1)
	id2 := vmessIdentity(j2)

	if id1 != id2 {
		t.Fatalf("vmessIdentity should ignore ps, got %q != %q", id1, id2)
	}
}

func TestParseVless(t *testing.T) {
	link := "vless://11111111-1111-1111-1111-111111111111@example.com:443?type=ws&security=tls&host=cdn.example.com&path=%2Fws&encryption=none#my-vless"

	got, err := parseVless(link)
	if err != nil {
		t.Fatalf("parseVless error = %v", err)
	}
	if got == nil {
		t.Fatal("parseVless returned nil")
	}

	if got.Outbound["protocol"] != "vless" {
		t.Fatalf("protocol = %v, want vless", got.Outbound["protocol"])
	}
	if got.Outbound["tag"] != "my-vless" {
		t.Fatalf("tag = %v, want my-vless", got.Outbound["tag"])
	}

	settings := got.Outbound["settings"].(map[string]any)
	vnext := settings["vnext"].([]any)
	server := vnext[0].(map[string]any)

	if server["address"] != addressPlaceholder {
		t.Fatalf("address = %v, want %v", server["address"], addressPlaceholder)
	}
	if server["port"] != 443 {
		t.Fatalf("port = %v, want 443", server["port"])
	}

	users := server["users"].([]any)
	user := users[0].(map[string]any)

	if user["id"] != "11111111-1111-1111-1111-111111111111" {
		t.Fatalf("id = %v", user["id"])
	}
	if user["encryption"] != "none" {
		t.Fatalf("encryption = %v, want none", user["encryption"])
	}

	if got.Identity == "" {
		t.Fatal("identity is empty")
	}
}

func TestParseTrojan(t *testing.T) {
	link := "trojan://secret@example.com:443?type=tcp&security=tls#my-trojan"

	got, err := parseTrojan(link)
	if err != nil {
		t.Fatalf("parseTrojan error = %v", err)
	}
	if got == nil {
		t.Fatal("parseTrojan returned nil")
	}

	if got.Outbound["protocol"] != "trojan" {
		t.Fatalf("protocol = %v, want trojan", got.Outbound["protocol"])
	}
	if got.Outbound["tag"] != "my-trojan" {
		t.Fatalf("tag = %v, want my-trojan", got.Outbound["tag"])
	}

	settings := got.Outbound["settings"].(map[string]any)
	servers := settings["servers"].([]any)
	server := servers[0].(map[string]any)

	if server["address"] != addressPlaceholder {
		t.Fatalf("address = %v, want %v", server["address"], addressPlaceholder)
	}
	if server["port"] != 443 {
		t.Fatalf("port = %v, want 443", server["port"])
	}
	if server["password"] != "secret" {
		t.Fatalf("password = %v, want secret", server["password"])
	}
}

func TestParseShadowsocks_Modern(t *testing.T) {
	userinfo := base64.StdEncoding.EncodeToString([]byte("aes-256-gcm:secret"))
	link := "ss://" + userinfo + "@example.com:8388#ss-node"

	got, err := parseShadowsocks(link)
	if err != nil {
		t.Fatalf("parseShadowsocks error = %v", err)
	}
	if got == nil {
		t.Fatal("parseShadowsocks returned nil")
	}

	if got.Outbound["protocol"] != "shadowsocks" {
		t.Fatalf("protocol = %v, want shadowsocks", got.Outbound["protocol"])
	}
	if got.Outbound["tag"] != "ss-node" {
		t.Fatalf("tag = %v, want ss-node", got.Outbound["tag"])
	}
	if got.Identity != "ss:aes-256-gcm:secret@example.com:8388" {
		t.Fatalf("identity = %q", got.Identity)
	}

	settings := got.Outbound["settings"].(map[string]any)
	servers := settings["servers"].([]any)
	server := servers[0].(map[string]any)

	if server["address"] != addressPlaceholder {
		t.Fatalf("address = %v, want %v", server["address"], addressPlaceholder)
	}
	if server["port"] != 8388 {
		t.Fatalf("port = %v, want 8388", server["port"])
	}
	if server["method"] != "aes-256-gcm" {
		t.Fatalf("method = %v", server["method"])
	}
	if server["password"] != "secret" {
		t.Fatalf("password = %v", server["password"])
	}
}

func TestParseShadowsocks_Legacy(t *testing.T) {
	raw := "aes-256-gcm:secret@example.com:8388"
	link := "ss://" + base64.StdEncoding.EncodeToString([]byte(raw)) + "#legacy-ss"

	got, err := parseShadowsocks(link)
	if err != nil {
		t.Fatalf("parseShadowsocks error = %v", err)
	}
	if got == nil {
		t.Fatal("parseShadowsocks returned nil")
	}

	if got.Outbound["tag"] != "legacy-ss" {
		t.Fatalf("tag = %v, want legacy-ss", got.Outbound["tag"])
	}
	if got.Identity != "ss:aes-256-gcm:secret@example.com:8388" {
		t.Fatalf("identity = %q", got.Identity)
	}
}

func TestParseHysteria2(t *testing.T) {
	link := "hy2://mypassword@example.com:8443?sni=example.com&alpn=h3,h3-29&fp=chrome#hy2-node"

	got, err := parseHysteria2(link)
	if err != nil {
		t.Fatalf("parseHysteria2 error = %v", err)
	}
	if got == nil {
		t.Fatal("parseHysteria2 returned nil")
	}

	if got.Outbound["protocol"] != "hysteria" {
		t.Fatalf("protocol = %v, want hysteria", got.Outbound["protocol"])
	}
	if got.Outbound["tag"] != "hy2-node" {
		t.Fatalf("tag = %v, want hy2-node", got.Outbound["tag"])
	}

	settings := got.Outbound["settings"].(map[string]any)
	if settings["address"] != addressPlaceholder {
		t.Fatalf("address = %v, want %v", settings["address"], addressPlaceholder)
	}
	if settings["port"] != 8443 {
		t.Fatalf("port = %v, want 8443", settings["port"])
	}
	if settings["version"] != 2 {
		t.Fatalf("version = %v, want 2", settings["version"])
	}

	stream := got.Outbound["streamSettings"].(map[string]any)
	if stream["network"] != "hysteria" {
		t.Fatalf("network = %v, want hysteria", stream["network"])
	}
	if stream["security"] != "tls" {
		t.Fatalf("security = %v, want tls", stream["security"])
	}
}

func TestParseWireguard(t *testing.T) {
	link := "wireguard://secret@example.com:51820?publickey=pubkey123&address=10.0.0.2%2F32,fd00::2%2F128&allowedips=0.0.0.0%2F0,::%2F0&mtu=1280&reserved=1,2,3#wg-node"

	got, err := parseWireguard(link)
	if err != nil {
		t.Fatalf("parseWireguard error = %v", err)
	}
	if got == nil {
		t.Fatal("parseWireguard returned nil")
	}

	if got.Outbound["protocol"] != "wireguard" {
		t.Fatalf("protocol = %v, want wireguard", got.Outbound["protocol"])
	}
	if got.Outbound["tag"] != "wg-node" {
		t.Fatalf("tag = %v, want wg-node", got.Outbound["tag"])
	}

	settings := got.Outbound["settings"].(map[string]any)
	if settings["secretKey"] != "secret" {
		t.Fatalf("secretKey = %v, want secret", settings["secretKey"])
	}
	if settings["mtu"] != 1280 {
		t.Fatalf("mtu = %v, want 1280", settings["mtu"])
	}

	peers := settings["peers"].([]any)
	peer := peers[0].(map[string]any)

	if peer["publicKey"] != "pubkey123" {
		t.Fatalf("publicKey = %v", peer["publicKey"])
	}
	if peer["endpoint"] != "example.com:51820" {
		t.Fatalf("endpoint = %v, want example.com:51820", peer["endpoint"])
	}
}
