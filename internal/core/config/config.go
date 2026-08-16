// Package config defines scanner settings, defaults, persistent storage,
// and platform/tier detection for automatic configuration tuning.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"bgscan/internal/core/fileutil"
)

// AppVersion is the current application version.
const AppVersion = "2.9.1"

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

// ScannerConfig holds all configuration settings used by the scanner.
type ScannerConfig struct {
	General GeneralConfig
	Writer  WriterConfig
	ICMP    ICMPConfig
	TCP     TCPConfig
	HTTP    HTTPConfig
	Xray    XrayConfig
	DNS     DNSConfig
}

// Store reads and writes scanner settings in a specific directory.
type Store struct {
	dir string
}

// StoreOption configures a Store during creation.
type StoreOption func(*Store)

// WithSettingsDir sets the directory used to read and write settings.
func WithSettingsDir(dir string) StoreOption {
	return func(s *Store) {
		s.dir = dir
	}
}

// NewStore creates a Store rooted at the application base directory's
// "settings" subdirectory by default.
func NewStore(opts ...StoreOption) Store {
	s := Store{dir: settingsDir}

	if base, err := fileutil.BasePath(); err == nil {
		s = Store{dir: filepath.Join(base, settingsDir)}
	}

	for _, opt := range opts {
		opt(&s)
	}

	return s
}

// Load reads all configuration files from the store directory.
// Missing files are created with defaults; malformed files return an error.
func (s Store) Load() (ScannerConfig, error) {
	general, err := loadTOML(s.path(generalFile), DefaultGeneralConfig())
	if err != nil {
		return ScannerConfig{}, err
	}

	writer, err := loadTOML(s.path(writerFile), DefaultWriterConfig())
	if err != nil {
		return ScannerConfig{}, err
	}

	icmp, err := loadTOML(s.path(icmpFile), DefaultICMPConfig())
	if err != nil {
		return ScannerConfig{}, err
	}

	tcp, err := loadTOML(s.path(tcpFile), DefaultTCPConfig())
	if err != nil {
		return ScannerConfig{}, err
	}

	http, err := loadTOML(s.path(httpFile), DefaultHTTPConfig())
	if err != nil {
		return ScannerConfig{}, err
	}

	xray, err := loadTOML(s.path(xrayFile), DefaultXrayConfig())
	if err != nil {
		return ScannerConfig{}, err
	}

	dns, err := loadTOML(s.path(dnsFile), DefaultDNSConfig())
	if err != nil {
		return ScannerConfig{}, err
	}

	return ScannerConfig{
		General: general,
		Writer:  writer,
		ICMP:    icmp,
		TCP:     tcp,
		HTTP:    http,
		Xray:    xray,
		DNS:     dns,
	}, nil
}

// SaveGeneral writes the general configuration to disk.
func (s Store) SaveGeneral(cfg GeneralConfig) error {
	return saveTOML(s.path(generalFile), cfg)
}

// SaveWriter writes the writer configuration to disk.
func (s Store) SaveWriter(cfg WriterConfig) error {
	return saveTOML(s.path(writerFile), cfg)
}

// SaveICMP writes the ICMP configuration to disk.
func (s Store) SaveICMP(cfg ICMPConfig) error {
	return saveTOML(s.path(icmpFile), cfg)
}

// SaveTCP writes the TCP configuration to disk.
func (s Store) SaveTCP(cfg TCPConfig) error {
	return saveTOML(s.path(tcpFile), cfg)
}

// SaveHTTP writes the HTTP configuration to disk.
func (s Store) SaveHTTP(cfg HTTPConfig) error {
	return saveTOML(s.path(httpFile), cfg)
}

// SaveXray writes the Xray configuration to disk.
func (s Store) SaveXray(cfg XrayConfig) error {
	return saveTOML(s.path(xrayFile), cfg)
}

// SaveDNS writes the DNS configuration to disk.
func (s Store) SaveDNS(cfg DNSConfig) error {
	return saveTOML(s.path(dnsFile), cfg)
}

func (s Store) path(filename string) string {
	return filepath.Join(s.dir, filename)
}

func loadTOML[T any](path string, defaultValue T) (T, error) {
	value, err := fileutil.ReadTOMLFile[T](path)
	if err == nil {
		return value, nil
	}

	if !errors.Is(err, os.ErrNotExist) {
		return defaultValue, fmt.Errorf("load settings %q: %w", path, err)
	}

	if err := fileutil.WriteTOMLFile(path, defaultValue); err != nil {
		return defaultValue, fmt.Errorf("create default settings %q: %w", path, err)
	}

	return defaultValue, nil
}

func saveTOML[T any](path string, value T) error {
	if err := fileutil.WriteTOMLFile(path, value); err != nil {
		return fmt.Errorf("save settings %q: %w", path, err)
	}

	return nil
}
