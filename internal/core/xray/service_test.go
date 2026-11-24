package xray

import (
	"context"
	"encoding/json"
	"net/netip"
	"strings"
	"testing"
)

// TestServiceVersion_NonEmpty: the service reports the embedded core version.
func TestServiceVersion_NonEmpty(t *testing.T) {
	svc := NewXrayService()

	if got := svc.Version(); got == "" {
		t.Fatal("Version() = empty string")
	}
}

// TestGenerateConfig_InvalidIP: an empty address is rejected.
func TestGenerateConfig_InvalidIP(t *testing.T) {
	svc := NewXrayService()

	if _, err := svc.GenerateConfig("anything", "", 1080, 0); err == nil {
		t.Fatal("expected error for invalid IP")
	}
}

// TestGenerateConfig_MissingTemplate: an unknown template name returns an error.
func TestGenerateConfig_MissingTemplate(t *testing.T) {
	svc := NewXrayService()

	ip := netip.MustParseAddr("1.2.3.4")
	if _, err := svc.GenerateConfig("does-not-exist-xyz", ip.String(), 1080, 0); err == nil {
		t.Fatal("expected error for missing outbound template")
	}
}

// TestValidateConfig_Nil: a nil config is rejected.
func TestValidateConfig_Nil(t *testing.T) {
	svc := NewXrayService()

	if err := svc.ValidateConfig(context.Background(), nil); err == nil {
		t.Fatal("expected error for nil config")
	}
}

// TestValidateConfig_NoInboundsOrOutbounds: a config with neither inbounds nor outbounds is rejected.
func TestValidateConfig_NoInboundsOrOutbounds(t *testing.T) {
	svc := NewXrayService()

	if err := svc.ValidateConfig(context.Background(), &Config{}); err == nil {
		t.Fatal("expected an error for a config with no inbound/outbound")
	}
}

// TestValidateConfig_NoOutbounds: a config without outbounds is rejected.
func TestValidateConfig_NoOutbounds(t *testing.T) {
	svc := NewXrayService()

	cfg := &Config{Inbounds: []Inbound{getInbound(18080)}}
	if err := svc.ValidateConfig(context.Background(), cfg); err == nil {
		t.Fatal("expected an error for a config with no outbounds")
	}
}

// TestStart_NilConfig: Start rejects a nil config.
func TestStart_NilConfig(t *testing.T) {
	svc := NewXrayService()

	if _, err := svc.Start(context.Background(), nil); err == nil {
		t.Fatal("expected error for nil config")
	}
}

// TestStart_CanceledContext: Start fails on an already cancelled context.
func TestStart_CanceledContext(t *testing.T) {
	svc := NewXrayService()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	cfg := &Config{
		Inbounds:  []Inbound{getInbound(18081)},
		Outbounds: []json.RawMessage{json.RawMessage(`{"protocol": "freedom"}`)},
	}
	if _, err := svc.Start(ctx, cfg); err == nil {
		t.Fatal("expected error for canceled context")
	}
}

// Guards the core JSON shape: plural "inbounds"/"outbounds" keys.
func TestToCoreJSON_UsesPluralKeys(t *testing.T) {
	cfg := &Config{
		Inbounds:  []Inbound{getInbound(1080)},
		Outbounds: []json.RawMessage{json.RawMessage(`{"protocol": "freedom"}`)},
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
