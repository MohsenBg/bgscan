// Package config provides centralized configuration management for the scanner.
// It exposes a global thread‑safe singleton, supports protocol‑specific
// settings, and persists configurations using TOML files.
package config

import (
	"fmt"
	"path/filepath"
	"sync"

	"bgscan/internal/core/fileutil"
)

const AppVersion = "2.4.2"

// ============================================================================
// Singleton
// ============================================================================

var (
	instance *ScannerConfig
	once     sync.Once
	mu       sync.RWMutex
)

// Get returns the global thread‑safe ScannerConfig singleton instance.
func Get() *ScannerConfig {
	once.Do(func() {
		instance = &ScannerConfig{
			General: DefaultGeneralConfig(),
			Writer:  DefaultWriterConfig(),
			ICMP:    DefaultICMPConfig(),
			TCP:     DefaultTCPConfig(),
			HTTP:    DefaultHTTPConfig(),
			Xray:    DefaultXrayConfig(),
			DNS:     DefaultDNSConfig(),
		}
	})
	return instance
}

// ============================================================================
// Thread‑safe Accessors and Setters
// ============================================================================

func GetGeneral() *GeneralConfig { mu.RLock(); defer mu.RUnlock(); return Get().General }
func GetWriter() *WriterConfig   { mu.RLock(); defer mu.RUnlock(); return Get().Writer }
func GetICMP() *ICMPConfig       { mu.RLock(); defer mu.RUnlock(); return Get().ICMP }
func GetTCP() *TCPConfig         { mu.RLock(); defer mu.RUnlock(); return Get().TCP }
func GetHTTP() *HTTPConfig       { mu.RLock(); defer mu.RUnlock(); return Get().HTTP }
func GetXray() *XrayConfig       { mu.RLock(); defer mu.RUnlock(); return Get().Xray }
func GetDNS() *DNSConfig         { mu.RLock(); defer mu.RUnlock(); return Get().DNS }

func setGeneral(cfg *GeneralConfig) { mu.Lock(); defer mu.Unlock(); Get().General = cfg }
func setWriter(cfg *WriterConfig)   { mu.Lock(); defer mu.Unlock(); Get().Writer = cfg }
func setICMP(cfg *ICMPConfig)       { mu.Lock(); defer mu.Unlock(); Get().ICMP = cfg }
func setTCP(cfg *TCPConfig)         { mu.Lock(); defer mu.Unlock(); Get().TCP = cfg }
func setHTTP(cfg *HTTPConfig)       { mu.Lock(); defer mu.Unlock(); Get().HTTP = cfg }
func setXray(cfg *XrayConfig)       { mu.Lock(); defer mu.Unlock(); Get().Xray = cfg }
func setDNS(cfg *DNSConfig)         { mu.Lock(); defer mu.Unlock(); Get().DNS = cfg }

// ============================================================================
// File Paths
// ============================================================================

const (
	settingsDir = "settings"

	generalFile = "general_settings.toml"
	writerFile  = "writer_settings.toml"
	icmpFile    = "icmp_settings.toml"
	tcpFile     = "tcp_settings.toml"
	httpFile    = "http_settings.toml"
	xrayFile    = "xray_settings.toml"
	dnsFile     = "dns_settings.toml"
)

// configPath returns the fully‑qualified path for a settings file.
func configPath(filename string) (string, error) {
	base, err := fileutil.GetCurrentPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, settingsDir, filename), nil
}

// ============================================================================
// Generic Load/Save Helpers
// ============================================================================

// loadConfig loads a TOML configuration file into cfg or falls back to defaults.
// After loading, it applies the provided setter to update the global state.
func loadConfig[T any](filename string, cfg *T, def *T, set func(*T)) error {
	path, err := configPath(filename)
	if err != nil {
		return err
	}

	if err := fileutil.GetTOMLFileOrDefault(path, cfg, def); err != nil {
		return fmt.Errorf("load config %s: %w", filename, err)
	}

	set(cfg)
	return nil
}

// saveConfig writes cfg to disk as TOML and updates the global in‑memory state.
func saveConfig[T any](filename string, cfg *T, set func(*T)) error {
	path, err := configPath(filename)
	if err != nil {
		return err
	}

	if err := fileutil.WriteTOMLFile(path, cfg); err != nil {
		return fmt.Errorf("save config %s: %w", filename, err)
	}

	set(cfg)
	return nil
}

// ============================================================================
// Public Load/Save Entry Points
// ============================================================================

// LoadGeneralConfig loads general settings from disk.
func LoadGeneralConfig() error {
	return loadConfig(generalFile, &GeneralConfig{}, DefaultGeneralConfig(), setGeneral)
}

// LoadWriterConfig loads writer settings from disk.
func LoadWriterConfig() error {
	return loadConfig(writerFile, &WriterConfig{}, DefaultWriterConfig(), setWriter)
}

// LoadICMPConfig loads ICMP settings.
func LoadICMPConfig() error {
	return loadConfig(icmpFile, &ICMPConfig{}, DefaultICMPConfig(), setICMP)
}

// LoadTCPConfig loads TCP settings.
func LoadTCPConfig() error {
	return loadConfig(tcpFile, &TCPConfig{}, DefaultTCPConfig(), setTCP)
}

// LoadHTTPConfig loads HTTP settings.
func LoadHTTPConfig() error {
	return loadConfig(httpFile, &HTTPConfig{}, DefaultHTTPConfig(), setHTTP)
}

// LoadXrayConfig loads Xray vulnerability settings.
func LoadXrayConfig() error {
	return loadConfig(xrayFile, &XrayConfig{}, DefaultXrayConfig(), setXray)
}

// LoadDNSConfig loads DNS settings.
func LoadDNSConfig() error {
	return loadConfig(dnsFile, &DNSConfig{}, DefaultDNSConfig(), setDNS)
}

func SaveGeneralConfig(cfg *GeneralConfig) error { return saveConfig(generalFile, cfg, setGeneral) }
func SaveWriterConfig(cfg *WriterConfig) error   { return saveConfig(writerFile, cfg, setWriter) }
func SaveICMPConfig(cfg *ICMPConfig) error       { return saveConfig(icmpFile, cfg, setICMP) }
func SaveTCPConfig(cfg *TCPConfig) error         { return saveConfig(tcpFile, cfg, setTCP) }
func SaveHTTPConfig(cfg *HTTPConfig) error       { return saveConfig(httpFile, cfg, setHTTP) }
func SaveXrayConfig(cfg *XrayConfig) error       { return saveConfig(xrayFile, cfg, setXray) }
func SaveDNSConfig(cfg *DNSConfig) error         { return saveConfig(dnsFile, cfg, setDNS) }

// ============================================================================
// Initialization
// ============================================================================

// Init loads all configuration files into the global singleton instance.
func Init() error {
	loaders := []func() error{
		LoadGeneralConfig,
		LoadWriterConfig,
		LoadICMPConfig,
		LoadTCPConfig,
		LoadHTTPConfig,
		LoadXrayConfig,
		LoadDNSConfig,
	}

	for _, load := range loaders {
		if err := load(); err != nil {
			return err
		}
	}

	return nil
}
