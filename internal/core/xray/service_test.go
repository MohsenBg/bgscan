package xray

import (
	"context"
	"os"
	"testing"
)

func xrayBin(t *testing.T) XrayService {
	t.Helper()
	svc, err := NewXrayService()
	if err != nil {
		t.Skip("xray binary not found, skipping integration test")
	}
	return svc
}

func TestFindXrayBinary_NotFound(t *testing.T) {
	_, err := FindXrayBinary()
	if err == nil {
		t.Log("binary found — skipping absence check")
	}
}

func TestServiceVersion(t *testing.T) {
	svc := xrayBin(t)

	version, err := svc.Version()
	if err != nil {
		t.Fatalf("Version() error: %v", err)
	}

	if version == "" {
		t.Fatal("Version() = empty string")
	}
}

func TestServiceValidateConfig_FileNotExist(t *testing.T) {
	svc := xrayBin(t)
	err := svc.ValidateConfig(context.Background(), "/nonexistent/path/config.json")
	if err == nil {
		t.Fatal("expected error for missing config file")
	}
}

func TestServiceValidateConfig_InvalidConfig(t *testing.T) {
	svc := xrayBin(t)
	f, err := os.CreateTemp(t.TempDir(), "*.json")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = f.WriteString(`{"invalid": true}`)
	_ = f.Close()

	err = svc.ValidateConfig(context.Background(), f.Name())
	if err == nil {
		t.Fatal("expected validation error for invalid config")
	}
}

func TestServiceStart_FileNotExist(t *testing.T) {
	svc := xrayBin(t)
	_, err := svc.Start(context.Background(), "/nonexistent/config.json")
	if err == nil {
		t.Fatal("expected error for missing config file")
	}
}
