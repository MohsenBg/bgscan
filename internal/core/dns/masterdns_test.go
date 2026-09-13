package dns

import (
	"context"
	"testing"
)

func validMasterConfig() MasterDNSConfig {
	cfg := DefaultMasterDNSConfig()
	cfg.Domain = "c1.example.net"
	cfg.EncryptionKey = "cd6d78e954f48f62cb74cdcf8a2459d3"
	return cfg
}

func TestMasterDNSValidate(t *testing.T) {
	if errs := validMasterConfig().Validate(); len(errs) != 0 {
		t.Fatalf("valid config errors = %v", errs)
	}

	empty := validMasterConfig()
	empty.Domain = ""
	if errs := empty.Validate(); errs["domain"] == nil {
		t.Error("expected domain error")
	}
}

func TestMasterDNSConfigStoreRoundTrip(t *testing.T) {
	svc := NewMasterDNSService(WithMasterDNSDir(t.TempDir()))

	cfg := validMasterConfig()
	if err := svc.SaveConfig(cfg, "e2e"); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}

	loaded, err := svc.LoadConfig("e2e")
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if loaded.Domain != cfg.Domain || loaded.DataEncMethod != cfg.DataEncMethod {
		t.Fatalf("round trip mismatch: %+v", loaded)
	}

	if err := svc.RenameConfig("e2e", "e2e2"); err != nil {
		t.Fatalf("RenameConfig: %v", err)
	}
	if _, err := svc.LoadConfig("e2e2"); err != nil {
		t.Fatalf("LoadConfig after rename: %v", err)
	}
}

func TestMasterDNSRunTunnelArgValidation(t *testing.T) {
	svc := NewMasterDNSService(WithMasterDNSDir(t.TempDir()))
	cfg := validMasterConfig()
	ctx := context.Background()

	if _, err := svc.RunTunnel(ctx, cfg, "not-an-ip", 18080); err == nil {
		t.Error("expected error for invalid resolver IP")
	}
	if _, err := svc.RunTunnel(ctx, cfg, "1.1.1.1", 0); err == nil {
		t.Error("expected error for zero listen port")
	}

	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := svc.RunTunnel(canceled, cfg, "1.1.1.1", 18080); err == nil {
		t.Error("expected error for canceled context")
	}

	bad := cfg
	bad.EncryptionKey = ""
	if _, err := svc.RunTunnel(ctx, bad, "1.1.1.1", 18080); err == nil {
		t.Error("expected error for invalid config")
	}
}
