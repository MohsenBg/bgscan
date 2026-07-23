package config

import (
	"sync"
	"testing"
)

// ============================================================================
// Singleton
// ============================================================================

func TestGet_ReturnsSameInstance(t *testing.T) {
	a := Get()
	b := Get()
	if a != b {
		t.Error("Get() returned different instances — singleton broken")
	}
}

func TestGet_NotNil(t *testing.T) {
	cfg := Get()
	if cfg == nil {
		t.Fatal("Get() returned nil")
	}
}

func TestGet_AllSubconfigsInitialized(t *testing.T) {
	cfg := Get()

	if cfg.General == nil {
		t.Error("General is nil")
	}
	if cfg.Writer == nil {
		t.Error("Writer is nil")
	}
	if cfg.ICMP == nil {
		t.Error("ICMP is nil")
	}
	if cfg.TCP == nil {
		t.Error("TCP is nil")
	}
	if cfg.HTTP == nil {
		t.Error("HTTP is nil")
	}
	if cfg.Xray == nil {
		t.Error("Xray is nil")
	}
	if cfg.DNS == nil {
		t.Error("DNS is nil")
	}
}

// ============================================================================
// Thread-safe Getters
// ============================================================================

func TestGetters_ReturnNonNil(t *testing.T) {
	if GetGeneral() == nil {
		t.Error("GetGeneral() returned nil")
	}
	if GetWriter() == nil {
		t.Error("GetWriter() returned nil")
	}
	if GetICMP() == nil {
		t.Error("GetICMP() returned nil")
	}
	if GetTCP() == nil {
		t.Error("GetTCP() returned nil")
	}
	if GetHTTP() == nil {
		t.Error("GetHTTP() returned nil")
	}
	if GetXray() == nil {
		t.Error("GetXray() returned nil")
	}
	if GetDNS() == nil {
		t.Error("GetDNS() returned nil")
	}
}

func TestGetters_MatchSingleton(t *testing.T) {
	cfg := Get()

	if GetGeneral() != cfg.General {
		t.Error("GetGeneral() does not match singleton General")
	}
	if GetWriter() != cfg.Writer {
		t.Error("GetWriter() does not match singleton Writer")
	}
	if GetICMP() != cfg.ICMP {
		t.Error("GetICMP() does not match singleton ICMP")
	}
	if GetTCP() != cfg.TCP {
		t.Error("GetTCP() does not match singleton TCP")
	}
	if GetHTTP() != cfg.HTTP {
		t.Error("GetHTTP() does not match singleton HTTP")
	}
	if GetXray() != cfg.Xray {
		t.Error("GetXray() does not match singleton Xray")
	}
	if GetDNS() != cfg.DNS {
		t.Error("GetDNS() does not match singleton DNS")
	}
}

// ============================================================================
// Internal Setters
// ============================================================================

func TestSetters_UpdateSingleton(t *testing.T) {
	original := GetGeneral()

	newCfg := DefaultGeneralConfig()
	newCfg.BatchSize = 9999
	setGeneral(newCfg)

	if GetGeneral().BatchSize != 9999 {
		t.Errorf("after setGeneral, BatchSize = %d, want 9999", GetGeneral().BatchSize)
	}

	// restore
	setGeneral(original)
}

func TestSetters_AllSubconfigs(t *testing.T) {
	// Writer
	origWriter := GetWriter()
	newWriter := DefaultWriterConfig()
	newWriter.ChanSize = 512
	setWriter(newWriter)
	if GetWriter().ChanSize != 512 {
		t.Errorf("setWriter: ChanSize = %d, want 512", GetWriter().ChanSize)
	}
	setWriter(origWriter)

	// ICMP
	origICMP := GetICMP()
	newICMP := DefaultICMPConfig()
	newICMP.Workers = 42
	setICMP(newICMP)
	if GetICMP().Workers != 42 {
		t.Errorf("setICMP: Workers = %d, want 42", GetICMP().Workers)
	}
	setICMP(origICMP)

	// TCP
	origTCP := GetTCP()
	newTCP := DefaultTCPConfig()
	newTCP.Port = 8080
	setTCP(newTCP)
	if GetTCP().Port != 8080 {
		t.Errorf("setTCP: Port = %d, want 8080", GetTCP().Port)
	}
	setTCP(origTCP)

	// HTTP
	origHTTP := GetHTTP()
	newHTTP := DefaultHTTPConfig()
	newHTTP.Host = "test.local"
	setHTTP(newHTTP)
	if GetHTTP().Host != "test.local" {
		t.Errorf("setHTTP: Host = %q, want %q", GetHTTP().Host, "test.local")
	}
	setHTTP(origHTTP)

	// Xray
	origXray := GetXray()
	newXray := DefaultXrayConfig()
	newXray.Workers = 16
	setXray(newXray)
	if GetXray().Workers != 16 {
		t.Errorf("setXray: Workers = %d, want 16", GetXray().Workers)
	}
	setXray(origXray)

	// DNS
	origDNS := GetDNS()
	newDNS := DefaultDNSConfig()
	newDNS.Resolver.Port = 5353
	setDNS(newDNS)
	if GetDNS().Resolver.Port != 5353 {
		t.Errorf("setDNS: Resolver.Port = %d, want 5353", GetDNS().Resolver.Port)
	}
	setDNS(origDNS)
}

// ============================================================================
// Concurrency — run with: go test -race
// ============================================================================

func TestConcurrentReads(t *testing.T) {
	const goroutines = 100
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for range goroutines {
		go func() {
			defer wg.Done()
			_ = GetGeneral()
			_ = GetWriter()
			_ = GetICMP()
			_ = GetTCP()
			_ = GetHTTP()
			_ = GetXray()
			_ = GetDNS()
		}()
	}

	wg.Wait()
}

func TestConcurrentReadsAndWrites(t *testing.T) {
	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines * 2)

	// readers
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			_ = GetGeneral()
			_ = GetTCP()
		}()
	}

	// writers
	for range goroutines {
		go func() {
			defer wg.Done()
			setGeneral(DefaultGeneralConfig())
			setTCP(DefaultTCPConfig())
		}()
	}

	wg.Wait()
}

func TestConcurrentGet_SingleInstance(t *testing.T) {
	const goroutines = 200
	instances := make([]*ScannerConfig, goroutines)
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := range goroutines {
		go func() {
			defer wg.Done()
			instances[i] = Get()
		}()
	}

	wg.Wait()

	first := instances[0]
	for i, inst := range instances {
		if inst != first {
			t.Errorf("goroutine %d got a different instance — singleton broken under concurrency", i)
		}
	}
}

// ============================================================================
// AppVersion
// ============================================================================

func TestAppVersion_NotEmpty(t *testing.T) {
	if AppVersion == "" {
		t.Error("AppVersion is empty")
	}
}
