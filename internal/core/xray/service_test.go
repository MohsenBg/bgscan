package xray

import (
	"context"
	"encoding/json"
	"net/netip"
	"strings"
	"testing"
)

func TestServiceVersion_NonEmpty(t *testing.T) {
	svc := NewXrayService()

	if got := svc.Version(); got == "" {
		t.Fatal("Version() = empty string")
	}
}

func TestGenerateConfig_InvalidIP(t *testing.T) {
	svc := NewXrayService()

	if _, err := svc.GenerateConfig("anything", netip.Addr{}, 1080); err == nil {
		t.Fatal("expected error for invalid IP")
	}
}

func TestGenerateConfig_MissingTemplate(t *testing.T) {
	svc := NewXrayService()

	ip := netip.MustParseAddr("1.2.3.4")
	if _, err := svc.GenerateConfig("does-not-exist-xyz", ip, 1080); err == nil {
		t.Fatal("expected error for missing outbound template")
	}
}

func TestValidateConfig_Nil(t *testing.T) {
	svc := NewXrayService()

	if err := svc.ValidateConfig(context.Background(), nil); err == nil {
		t.Fatal("expected error for nil config")
	}
}

func TestValidateConfig_NoInboundsOrOutbounds(t *testing.T) {
	svc := NewXrayService()

	if err := svc.ValidateConfig(context.Background(), &XrayConfig{}); err == nil {
		t.Fatal("expected an error for a config with no inbound/outbound")
	}
}

func TestValidateConfig_NoOutbounds(t *testing.T) {
	svc := NewXrayService()

	cfg := &XrayConfig{Inbound: []Inbound{getInbound(18080)}}
	if err := svc.ValidateConfig(context.Background(), cfg); err == nil {
		t.Fatal("expected an error for a config with no outbounds")
	}
}

func TestStart_NilConfig(t *testing.T) {
	svc := NewXrayService()

	if _, err := svc.Start(context.Background(), nil); err == nil {
		t.Fatal("expected error for nil config")
	}
}

func TestStart_CanceledContext(t *testing.T) {
	svc := NewXrayService()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	cfg := &XrayConfig{
		Inbound:  []Inbound{getInbound(18081)},
		Outbound: []any{map[string]any{"protocol": "freedom"}},
	}
	if _, err := svc.Start(ctx, cfg); err == nil {
		t.Fatal("expected error for canceled context")
	}
}

// Guards the core JSON shape: plural "inbounds"/"outbounds" keys.
func TestToCoreJSON_UsesPluralKeys(t *testing.T) {
	cfg := &XrayConfig{
		Inbound:  []Inbound{getInbound(1080)},
		Outbound: []any{map[string]any{"protocol": "freedom"}},
	}

	raw, err := toCoreJSON(cfg)
	if err != nil {
		t.Fatalf("toCoreJSON() error = %v", err)
	}

	var decoded map[string]json.RawMessage
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if _, ok := decoded["inbounds"]; !ok {
		t.Fatalf("converted JSON keys = %v, want plural \"inbounds\"", keys(decoded))
	}
	if _, ok := decoded["outbounds"]; !ok {
		t.Fatalf("converted JSON keys = %v, want plural \"outbounds\"", keys(decoded))
	}
}

func keys(m map[string]json.RawMessage) string {
	var b strings.Builder
	first := true
	for k := range m {
		if !first {
			b.WriteString(", ")
		}
		b.WriteString(k)
		first = false
	}
	return b.String()
}
