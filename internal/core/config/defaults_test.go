package config

import (
	"testing"
	"time"
)

var (
	allPlatforms = []Platform{Android, Desktop, Server}
	allTiers     = []Tier{Low, Mid, High}
)

func TestGeneralDefaultsComplete(t *testing.T) {
	for _, p := range allPlatforms {
		for _, tr := range allTiers {
			if _, ok := generalDefaults[p][tr]; !ok {
				t.Errorf("missing generalDefaults entry for platform=%s tier=%s", p, tr)
			}
		}
	}
}

func TestWriterDefaultsComplete(t *testing.T) {
	for _, p := range allPlatforms {
		for _, tr := range allTiers {
			if _, ok := writerDefaults[p][tr]; !ok {
				t.Errorf("missing writerDefaults entry for platform=%s tier=%s", p, tr)
			}
		}
	}
}

func TestICMPDefaultsComplete(t *testing.T) {
	for _, p := range allPlatforms {
		for _, tr := range allTiers {
			if _, ok := icmpDefaults[p][tr]; !ok {
				t.Errorf("missing icmpDefaults entry for platform=%s tier=%s", p, tr)
			}
		}
	}
}

func TestTCPDefaultsComplete(t *testing.T) {
	for _, p := range allPlatforms {
		for _, tr := range allTiers {
			if _, ok := tcpDefaults[p][tr]; !ok {
				t.Errorf("missing tcpDefaults entry for platform=%s tier=%s", p, tr)
			}
		}
	}
}

func TestHTTPDefaultsComplete(t *testing.T) {
	for _, p := range allPlatforms {
		for _, tr := range allTiers {
			if _, ok := httpDefaults[p][tr]; !ok {
				t.Errorf("missing httpDefaults entry for platform=%s tier=%s", p, tr)
			}
		}
	}
}

func TestXrayDefaultsComplete(t *testing.T) {
	for _, p := range allPlatforms {
		for _, tr := range allTiers {
			if _, ok := xrayDefaults[p][tr]; !ok {
				t.Errorf("missing xrayDefaults entry for platform=%s tier=%s", p, tr)
			}
		}
	}
}

func TestDNSDefaultsComplete(t *testing.T) {
	for _, p := range allPlatforms {
		for _, tr := range allTiers {
			if _, ok := dnsDefaults[p][tr]; !ok {
				t.Errorf("missing dnsDefaults entry for platform=%s tier=%s", p, tr)
			}
		}
	}
}

func TestGeneralDefaultsServerHighValues(t *testing.T) {
	cfg := generalDefaults[Server][High]

	if cfg.StatusInterval != NewDurationMS(1*time.Second) {
		t.Errorf("StatusInterval = %v, want 1s", cfg.StatusInterval)
	}
	if cfg.StopAfterFound != 0 {
		t.Errorf("StopAfterFound = %d, want 0", cfg.StopAfterFound)
	}
	if cfg.MaxIPsToTest != 0 {
		t.Errorf("MaxIPsToTest = %d, want 0", cfg.MaxIPsToTest)
	}
	if cfg.MaxIPsPerStage != 500_000 {
		t.Errorf("MaxIPsPerStage = %d, want 500000", cfg.MaxIPsPerStage)
	}
	if cfg.BatchSize != 20_000 {
		t.Errorf("BatchSize = %d, want 20000", cfg.BatchSize)
	}
	if !cfg.Shuffled {
		t.Error("Shuffled = false, want true")
	}
	if cfg.PipelineMode != "streaming" {
		t.Errorf("PipelineMode = %q, want %q", cfg.PipelineMode, "streaming")
	}
}

func TestWriterDefaultsDesktopMidValues(t *testing.T) {
	cfg := writerDefaults[Desktop][Mid]

	if cfg.MergeFlushInterval != NewDurationMS(2*time.Second) {
		t.Errorf("MergeFlushInterval = %v, want 2s", cfg.MergeFlushInterval)
	}
	if cfg.ChanSize != 1024 {
		t.Errorf("ChanSize = %d, want 1024", cfg.ChanSize)
	}
	if cfg.BatchSize != 4096 {
		t.Errorf("BatchSize = %d, want 4096", cfg.BatchSize)
	}
	if cfg.ResultBaseDir != "result" {
		t.Errorf("ResultBaseDir = %q, want %q", cfg.ResultBaseDir, "result")
	}
}

func TestICMPDefaultsServerMidValues(t *testing.T) {
	cfg := icmpDefaults[Server][Mid]

	if cfg.Workers != 400 {
		t.Errorf("Workers = %d, want 400", cfg.Workers)
	}
	if cfg.Timeout != NewDurationMS(2*time.Second) {
		t.Errorf("Timeout = %v, want 2s", cfg.Timeout)
	}
	if cfg.Tries != 1 {
		t.Errorf("Tries = %d, want 1", cfg.Tries)
	}
	if cfg.OutputPrefix != "icmp_" {
		t.Errorf("OutputPrefix = %q, want %q", cfg.OutputPrefix, "icmp_")
	}
}

func TestTCPDefaultsDesktopMidValues(t *testing.T) {
	cfg := tcpDefaults[Desktop][Mid]

	if cfg.Port != 443 {
		t.Errorf("Port = %d, want 443", cfg.Port)
	}
	if cfg.Tries != 1 {
		t.Errorf("Tries = %d, want 1", cfg.Tries)
	}
	if cfg.OutputPrefix != "tcp_" {
		t.Errorf("OutputPrefix = %q, want %q", cfg.OutputPrefix, "tcp_")
	}
}

func TestHTTPDefaultsFixedFieldsConsistentAcrossTable(t *testing.T) {
	// Fields that should be identical across every platform/tier, since
	// httpBase() only varies Workers/Timeout per entry.
	for _, p := range allPlatforms {
		for _, tr := range allTiers {
			cfg := httpDefaults[p][tr]
			if cfg.Host != "example.com" {
				t.Errorf("[%s/%s] Host = %q, want %q", p, tr, cfg.Host, "example.com")
			}
			if cfg.Port != 443 {
				t.Errorf("[%s/%s] Port = %d, want 443", p, tr, cfg.Port)
			}
			if cfg.Protocol != "https" {
				t.Errorf("[%s/%s] Protocol = %q, want %q", p, tr, cfg.Protocol, "https")
			}
			if !cfg.TLSValidation {
				t.Errorf("[%s/%s] TLSValidation = false, want true", p, tr)
			}
			if cfg.Version != "h1,h2" {
				t.Errorf("[%s/%s] Version = %q, want %q", p, tr, cfg.Version, "h1,h2")
			}
			if len(cfg.AcceptedStatusCodes) != 0 {
				t.Errorf("[%s/%s] AcceptedStatusCodes = %v, want empty", p, tr, cfg.AcceptedStatusCodes)
			}
			if cfg.MinTLSVersion != "tls1.1" {
				t.Errorf("[%s/%s] MinTLSVersion = %q, want %q", p, tr, cfg.MinTLSVersion, "tls1.1")
			}
			if cfg.MaxTLSVersion != "tls1.3" {
				t.Errorf("[%s/%s] MaxTLSVersion = %q, want %q", p, tr, cfg.MaxTLSVersion, "tls1.3")
			}
			if cfg.OutputPrefix != "http_" {
				t.Errorf("[%s/%s] OutputPrefix = %q, want %q", p, tr, cfg.OutputPrefix, "http_")
			}
		}
	}
}

func TestHTTPDefaultsDesktopMidValues(t *testing.T) {
	cfg := httpDefaults[Desktop][Mid]

	if cfg.Workers != 50 {
		t.Errorf("Workers = %d, want 50", cfg.Workers)
	}
	if cfg.Timeout != NewDurationMS(4*time.Second) {
		t.Errorf("Timeout = %v, want 4s", cfg.Timeout)
	}
}

func TestXrayDefaultsDesktopMidValues(t *testing.T) {
	cfg := xrayDefaults[Desktop][Mid]

	if cfg.Workers != 16 {
		t.Errorf("Workers = %d, want 16", cfg.Workers)
	}
	if cfg.ConnectivityTestType != Both {
		t.Errorf("ConnectivityTestType = %v, want Both", cfg.ConnectivityTestType)
	}
	if cfg.DownloadSpeed != 100 {
		t.Errorf("DownloadSpeed = %d, want 100", cfg.DownloadSpeed)
	}
	if cfg.UploadSpeed != 50 {
		t.Errorf("UploadSpeed = %d, want 50", cfg.UploadSpeed)
	}
	if cfg.PreScanType != "tcp" {
		t.Errorf("PreScanType = %q, want %q", cfg.PreScanType, "tcp")
	}
	if cfg.OutputPrefix != "xray_" {
		t.Errorf("OutputPrefix = %q, want %q", cfg.OutputPrefix, "xray_")
	}
}

func TestDNSDefaultsDesktopMidValues(t *testing.T) {
	cfg := dnsDefaults[Desktop][Mid]
	r := cfg.Resolver

	if r.Workers != 150 {
		t.Errorf("Resolver.Workers = %d, want 150", r.Workers)
	}
	if r.Transport != "udp" {
		t.Errorf("Resolver.Protocol = %q, want %q", r.Transport, "udp")
	}
	if r.Domain != "example.com" {
		t.Errorf("Resolver.Domain = %q, want %q", r.Domain, "example.com")
	}
	if r.Port != 53 {
		t.Errorf("Resolver.Port = %d, want 53", r.Port)
	}
	if len(r.CheckTypes) != 1 || r.CheckTypes[0] != "TXT" {
		t.Errorf("Resolver.CheckTypes = %v, want [TXT]", r.CheckTypes)
	}
	if r.EDNSBufSize != 1232 {
		t.Errorf("Resolver.EDNSBufSize = %d, want 1232", r.EDNSBufSize)
	}
	if !r.RandomSubdomain {
		t.Error("Resolver.RandomSubdomain = false, want true")
	}
	if len(r.AcceptedRCodes) != 3 {
		t.Errorf("Resolver.AcceptedRCodes = %v, want 3 entries", r.AcceptedRCodes)
	}
	if !r.DPI.Enabled {
		t.Error("Resolver.DPI.Enabled = false, want true")
	}
	if r.DPI.Timeout != NewDurationMS(2*time.Second) {
		t.Errorf("Resolver.DPI.Timeout = %v, want 2s", r.DPI.Timeout)
	}
	if r.DPI.Tries != 1 {
		t.Errorf("Resolver.DPI.Tries = %d, want 1", r.DPI.Tries)
	}
	if r.OutputPrefix != "dns_" {
		t.Errorf("Resolver.PrefixOutput = %q, want %q", r.OutputPrefix, "dns_")
	}

	d := cfg.DNSTunneling
	if d.Workers != 16 {
		t.Errorf("DNSTunneling.Workers = %d, want 16", d.Workers)
	}
	if !d.CheckDNSResolver {
		t.Error("DNSTunneling.CheckDNSResolver = false, want true")
	}
}

func TestHTTPDefaultsSlicesAreIndependent(t *testing.T) {
	a := httpDefaults[Desktop][Low]
	b := httpDefaults[Desktop][Mid]
	a.AcceptedStatusCodes = append(a.AcceptedStatusCodes, 200)
	if len(b.AcceptedStatusCodes) != 0 {
		t.Error("AcceptedStatusCodes slice is aliased across table entries")
	}
}

func TestDNSDefaultsSlicesAreIndependent(t *testing.T) {
	a := dnsDefaults[Desktop][Low].Resolver
	b := dnsDefaults[Desktop][Mid].Resolver
	a.CheckTypes = append(a.CheckTypes, "AAAA")
	if len(b.CheckTypes) != 1 {
		t.Error("CheckTypes slice is aliased across table entries")
	}
}

// ---- Smoke tests: Default*Config() must not panic and must return a
// valid (non-zero) entry from *some* row of its table, on any machine ----

func TestDefaultConfigsDoNotPanicAndAreNonZero(t *testing.T) {
	g := DefaultGeneralConfig()
	if g.MaxIPsPerStage == 0 || g.BatchSize == 0 {
		t.Error("DefaultGeneralConfig() returned zero-value fields")
	}

	w := DefaultWriterConfig()
	if w.ChanSize == 0 || w.BatchSize == 0 {
		t.Error("DefaultWriterConfig() returned zero-value fields")
	}

	i := DefaultICMPConfig()
	if i.Workers == 0 || i.OutputPrefix == "" {
		t.Error("DefaultICMPConfig() returned zero-value fields")
	}

	tc := DefaultTCPConfig()
	if tc.Workers == 0 || tc.OutputPrefix == "" {
		t.Error("DefaultTCPConfig() returned zero-value fields")
	}

	h := DefaultHTTPConfig()
	if h.Workers == 0 || h.Host == "" {
		t.Error("DefaultHTTPConfig() returned zero-value fields")
	}

	x := DefaultXrayConfig()
	if x.Workers == 0 || x.OutputPrefix == "" {
		t.Error("DefaultXrayConfig() returned zero-value fields")
	}

	d := DefaultDNSConfig()
	if d.Resolver.Workers == 0 || d.Resolver.OutputPrefix == "" {
		t.Error("DefaultDNSConfig() returned zero-value fields")
	}
}

func TestDefaultConfigsReturnIndependentSlices(t *testing.T) {
	a := DefaultHTTPConfig()
	b := DefaultHTTPConfig()
	a.AcceptedStatusCodes = append(a.AcceptedStatusCodes, 200)
	if len(b.AcceptedStatusCodes) != 0 {
		t.Error("DefaultHTTPConfig() shares AcceptedStatusCodes slice across calls")
	}

	c := DefaultDNSConfig()
	d := DefaultDNSConfig()
	c.Resolver.CheckTypes = append(c.Resolver.CheckTypes, "AAAA")
	if len(d.Resolver.CheckTypes) != 1 {
		t.Error("DefaultDNSConfig() shares Resolver.CheckTypes slice across calls")
	}
}
