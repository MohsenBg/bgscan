package xray

import (
	"context"
	"os"
	"testing"
)

func xrayBin(t *testing.T) string {
	t.Helper()
	bin, err := FindXrayBinary()
	if err != nil {
		t.Skip("xray binary not found, skipping integration test")
	}
	return bin
}

func TestFindXrayBinary_NotFound(t *testing.T) {
	_, err := FindXrayBinary()
	if err == nil {
		t.Log("binary found — skipping absence check")
	}
}

func TestValidateConfig_FileNotExist(t *testing.T) {
	xrayBin(t) // skip if no binary
	err := ValidateConfig(context.Background(), "/nonexistent/path/config.json")
	if err == nil {
		t.Fatal("expected error for missing config file")
	}
}

func TestValidateConfig_InvalidConfig(t *testing.T) {
	xrayBin(t)
	f, err := os.CreateTemp(t.TempDir(), "*.json")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = f.WriteString(`{"invalid": true}`)
	_ = f.Close()

	err = ValidateConfig(context.Background(), f.Name())
	if err == nil {
		t.Fatal("expected validation error for invalid config")
	}
}

func TestStartXray_FileNotExist(t *testing.T) {
	xrayBin(t)
	_, err := StartXray(context.Background(), "/nonexistent/config.json")
	if err == nil {
		t.Fatal("expected error for missing config file")
	}
}
