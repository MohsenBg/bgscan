package xray

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestReplacePlaceholders(t *testing.T) {
	in := map[string]any{
		"protocol": "vless",
		"settings": map[string]any{
			"vnext": []any{
				map[string]any{
					"address": addressPlaceholder,
					"port":    443,
				},
			},
		},
		"list": []any{
			addressPlaceholder,
			map[string]any{"address": addressPlaceholder},
			"keep",
		},
	}

	got := replacePlaceholders(in, map[string]string{
		addressPlaceholder: "1.2.3.4",
	})

	want := map[string]any{
		"protocol": "vless",
		"settings": map[string]any{
			"vnext": []any{
				map[string]any{
					"address": "1.2.3.4",
					"port":    443,
				},
			},
		},
		"list": []any{
			"1.2.3.4",
			map[string]any{"address": "1.2.3.4"},
			"keep",
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("replacePlaceholders() = %#v, want %#v", got, want)
	}
}

func TestApplyOutboundTemplate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "template.json")
	writeJSONFile(t, path, map[string]any{
		"protocol": "vmess",
		"settings": map[string]any{
			"vnext": []any{
				map[string]any{
					"address": addressPlaceholder,
					"port":    443,
				},
			},
		},
	})

	got, err := applyOutboundTemplate(path, "8.8.8.8")
	if err != nil {
		t.Fatalf("applyOutboundTemplate() error = %v", err)
	}

	root, ok := got.(map[string]any)
	if !ok {
		t.Fatalf("applyOutboundTemplate() returned %T, want map[string]any", got)
	}

	settings := root["settings"].(map[string]any)
	vnext := settings["vnext"].([]any)
	node := vnext[0].(map[string]any)

	if node["address"] != "8.8.8.8" {
		t.Fatalf("address = %#v, want 8.8.8.8", node["address"])
	}
	if node["port"] != float64(443) {
		t.Fatalf("port = %#v, want 443", node["port"])
	}
}

func TestApplyOutboundTemplate_InvalidIP(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "template.json")
	writeJSONFile(t, path, map[string]any{
		"address": addressPlaceholder,
	})

	_, err := applyOutboundTemplate(path, "not-an-ip")
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

	_, err := applyOutboundTemplate(path, "1.1.1.1")
	if err == nil {
		t.Fatal("expected JSON parse error")
	}
}

func TestContainsAddressPlaceholder(t *testing.T) {
	tests := []struct {
		name string
		in   any
		want bool
	}{
		{
			name: "direct field",
			in: map[string]any{
				"address": addressPlaceholder,
			},
			want: true,
		},
		{
			name: "nested",
			in: map[string]any{
				"settings": map[string]any{
					"vnext": []any{
						map[string]any{"address": addressPlaceholder},
					},
				},
			},
			want: true,
		},
		{
			name: "same token different key",
			in: map[string]any{
				"note": addressPlaceholder,
			},
			want: false,
		},
		{
			name: "missing",
			in: map[string]any{
				"address": "1.1.1.1",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsAddressPlaceholder(tt.in)
			if got != tt.want {
				t.Fatalf("containsAddressPlaceholder() = %v, want %v", got, tt.want)
			}
		})
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
		{"dir/node", "dir/node.json"},
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
	if !got.UseTLS {
		t.Fatal("UseTLS = false, want true")
	}
	if got.CreatedTime.IsZero() {
		t.Fatal("CreatedTime is zero")
	}
}

func TestSaveOutboundFromFile_Errors(t *testing.T) {
	t.Run("missing source", func(t *testing.T) {
		_, err := SaveOutboundFromFile(filepath.Join(t.TempDir(), "missing.json"), "node")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("source is directory", func(t *testing.T) {
		dir := t.TempDir()
		_, err := SaveOutboundFromFile(dir, "node")
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

		_, err := SaveOutboundFromFile(src, "node")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("missing address placeholder", func(t *testing.T) {
		dir := t.TempDir()
		src := filepath.Join(dir, "plain.json")
		writeJSONFile(t, src, map[string]any{
			"protocol": "vless",
			"settings": map[string]any{
				"vnext": []any{
					map[string]any{
						"address": "1.1.1.1",
						"port":    443,
					},
				},
			},
		})

		_, err := SaveOutboundFromFile(src, "node")
		if err == nil {
			t.Fatal("expected error")
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
