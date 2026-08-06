package config

import (
	"testing"
	"time"
)

func TestDefaultGeneralConfig(t *testing.T) {
	cfg := DefaultGeneralConfig()

	if cfg.StatusInterval != NewDurationMS(1*time.Second) {
		t.Errorf("StatusInterval = %v, want 1s", cfg.StatusInterval)
	}
	if cfg.StopAfterFound != 0 {
		t.Errorf("StopAfterFound = %d, want 0", cfg.StopAfterFound)
	}
	if cfg.MaxIPsToTest != 0 {
		t.Errorf("MaxIPsToTest = %d, want 0", cfg.MaxIPsToTest)
	}
	if cfg.MaxIPsPerStage != 100_000 {
		t.Errorf("MaxIPsPerStage = %d, want 100000", cfg.MaxIPsPerStage)
	}
	if cfg.BatchSize != 5_000 {
		t.Errorf("BatchSize = %d, want 5000", cfg.BatchSize)
	}
	if cfg.Shuffled == false {
		t.Errorf("Shuffled = %v, want true", cfg.Shuffled)
	}
	if cfg.PipelineMode != "streaming" {
		t.Errorf("PipelineMode = %q, want %q", cfg.PipelineMode, "streaming")
	}
}

func TestDefaultWriterConfig(t *testing.T) {
	cfg := DefaultWriterConfig()

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

func TestDefaultICMPConfig(t *testing.T) {
	cfg := DefaultICMPConfig()

	if cfg.Workers != 200 {
		t.Errorf("Workers = %d, want 200", cfg.Workers)
	}
	if cfg.Timeout != NewDurationMS(2*time.Second) {
		t.Errorf("Timeout = %v, want 2s", cfg.Timeout)
	}
	if cfg.Tries != 1 {
		t.Errorf("Tries = %d, want 1", cfg.Tries)
	}
	if cfg.OutputPrefix != "icmp_" {
		t.Errorf("PrefixOutput = %q, want %q", cfg.OutputPrefix, "icmp_")
	}
}

func TestDefaultTCPConfig(t *testing.T) {
	cfg := DefaultTCPConfig()

	if cfg.Workers != 400 {
		t.Errorf("Workers = %d, want 400", cfg.Workers)
	}
	if cfg.Port != 443 {
		t.Errorf("Port = %d, want 443", cfg.Port)
	}
	if cfg.Timeout != NewDurationMS(2*time.Second) {
		t.Errorf("Timeout = %v, want 2s", cfg.Timeout)
	}
	if cfg.Tries != 1 {
		t.Errorf("Tries = %d, want 1", cfg.Tries)
	}
	if cfg.OutputPrefix != "tcp_" {
		t.Errorf("PrefixOutput = %q, want %q", cfg.OutputPrefix, "tcp_")
	}
}

func TestDefaultHTTPConfig(t *testing.T) {
	cfg := DefaultHTTPConfig()

	if cfg.Workers != 50 {
		t.Errorf("Workers = %d, want 50", cfg.Workers)
	}
	if cfg.Host != "example.com" {
		t.Errorf("Host = %q, want %q", cfg.Host, "example.com")
	}
	if cfg.Port != 443 {
		t.Errorf("Port = %d, want 443", cfg.Port)
	}
	if cfg.Protocol != "https" {
		t.Errorf("Protocol = %q, want %q", cfg.Protocol, "https")
	}
	if !cfg.TLSValidation {
		t.Error("TLSValidation = false, want true")
	}
	if cfg.Version != "h1,h2" {
		t.Errorf("Version = %q, want %q", cfg.Version, "h1,h2")
	}
	if len(cfg.AcceptedStatusCodes) != 0 {
		t.Errorf("AcceptedStatusCodes = %v, want empty", cfg.AcceptedStatusCodes)
	}
	if cfg.MinTLSVersion != "tls1.1" {
		t.Errorf("MinTLSVersion = %q, want %q", cfg.MinTLSVersion, "tls1.1")
	}
	if cfg.MaxTLSVersion != "tls1.3" {
		t.Errorf("MaxTLSVersion = %q, want %q", cfg.MaxTLSVersion, "tls1.3")
	}
	if cfg.Timeout != NewDurationMS(4*time.Second) {
		t.Errorf("Timeout = %v, want 4s", cfg.Timeout)
	}
	if cfg.OutputPrefix != "http_" {
		t.Errorf("PrefixOutput = %q, want %q", cfg.OutputPrefix, "http_")
	}
}

func TestDefaultXrayConfig(t *testing.T) {
	cfg := DefaultXrayConfig()

	if cfg.Workers != 32 {
		t.Errorf("Workers = %d, want 32", cfg.Workers)
	}
	if cfg.ConnectivityTestType != Both {
		t.Errorf("ConnectivityTestType = %v, want ConnectivityOnly", cfg.ConnectivityTestType)
	}
	if cfg.DownloadSpeed != 100 {
		t.Errorf("DownloadSpeed = %d, want 100", cfg.DownloadSpeed)
	}
	if cfg.UploadSpeed != 50 {
		t.Errorf("UploadSpeed = %d, want 50", cfg.UploadSpeed)
	}
	if cfg.PreScanType != "tcp" {
		t.Errorf("PreScanType = %q, want %q", cfg.PreScanType, "none")
	}
	if cfg.Timeout != NewDurationMS(4*time.Second) {
		t.Errorf("Timeout = %v, want 4s", cfg.Timeout)
	}
	if cfg.OutputPrefix != "xray_" {
		t.Errorf("PrefixOutput = %q, want %q", cfg.OutputPrefix, "xray_")
	}
}

func TestDefaultDNSConfig(t *testing.T) {
	cfg := DefaultDNSConfig()

	r := cfg.Resolver

	if r.Workers != 100 {
		t.Errorf("Resolver.Workers = %d, want 100", r.Workers)
	}
	if r.Protocol != "udp" {
		t.Errorf("Resolver.Protocol = %q, want %q", r.Protocol, "udp")
	}
	if r.Domain != "google.com" {
		t.Errorf("Resolver.Domain = %q, want %q", r.Domain, "google.com")
	}
	if r.Port != 53 {
		t.Errorf("Resolver.Port = %d, want 53", r.Port)
	}
	if len(r.CheckTypes) != 1 || r.CheckTypes[0] != "A" {
		t.Errorf("Resolver.CheckTypes = %v, want [A]", r.CheckTypes)
	}
	if r.EDNSBufSize != 1234 {
		t.Errorf("Resolver.EDNSBufSize = %d, want 1234", r.EDNSBufSize)
	}
	if r.Timeout != NewDurationMS(2*time.Second) {
		t.Errorf("Resolver.Timeout = %v, want 2s", r.Timeout)
	}
	if r.Tries != 1 {
		t.Errorf("Resolver.Tries = %d, want 1", r.Tries)
	}
	if !r.RandomSubdomain {
		t.Error("Resolver.RandomSubdomain = false, want true")
	}
	if len(r.AcceptedRCodes) != 2 {
		t.Errorf("Resolver.AcceptedRCodes = %v, want [noerror nxdomain]", r.AcceptedRCodes)
	}
	if !r.CheckDPI {
		t.Error("Resolver.CheckDPI = false, want true")
	}
	if r.DPITimeout != NewDurationMS(500*time.Millisecond) {
		t.Errorf("Resolver.DPITimeout = %v, want 500ms", r.DPITimeout)
	}
	if r.DPITries != 2 {
		t.Errorf("Resolver.DPITries = %d, want 2", r.DPITries)
	}
	if r.PrefixOutput != "dns_resolver_" {
		t.Errorf("Resolver.PrefixOutput = %q, want %q", r.PrefixOutput, "dns_resolver_")
	}

	// DNSTT
	d := cfg.DNSTT

	if d.Enabled {
		t.Error("DNSTT.Enabled = true, want false")
	}
	if d.Workers != 20 {
		t.Errorf("DNSTT.Workers = %d, want 20", d.Workers)
	}
	if d.Domain != "ns.example.com" {
		t.Errorf("DNSTT.Domain = %q, want %q", d.Domain, "ns.example.com")
	}
	if d.OutputPrefix != "dns_dnstt_" {
		t.Errorf("DNSTT.PrefixOutput = %q, want %q", d.OutputPrefix, "dns_dnstt_")
	}

	// SlipStream
	s := cfg.SlipStream

	if s.Enabled {
		t.Error("SlipStream.Enabled = true, want false")
	}
	if s.Workers != 20 {
		t.Errorf("SlipStream.Workers = %d, want 20", s.Workers)
	}
	if s.Domain != "ns.example.com" {
		t.Errorf("SlipStream.Domain = %q, want %q", s.Domain, "ns.example.com")
	}
	if s.OutputPrefix != "dns_slipstream_" {
		t.Errorf("SlipStream.PrefixOutput = %q, want %q", s.OutputPrefix, "dns_slipstream_")
	}
}

func TestDefaultConfigsReturnNewInstances(t *testing.T) {
	a := DefaultGeneralConfig()
	b := DefaultGeneralConfig()
	if &a == &b {
		t.Error("DefaultGeneralConfig() returned same pointer on two calls")
	}

	c := DefaultDNSConfig()
	d := DefaultDNSConfig()
	if &c == &d {
		t.Error("DefaultDNSConfig() returned same pointer on two calls")
	}
	if &c.Resolver == &d.Resolver {
		t.Error("DefaultDNSConfig().Resolver returned same pointer on two calls")
	}
}
