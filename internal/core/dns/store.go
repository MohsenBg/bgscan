package dns

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/MohsenBg/bgscan/internal/core/fileutil"
)

// dnsTunnelingDir is the base directory holding per-protocol tunnel
// configuration files, relative to the application base path.
const dnsTunnelingDir = "assets/dns-tunneling"

// tunnelConfigDir returns the configuration directory for the named
// tunnel protocol, resolved against the application base path.
func tunnelConfigDir(name string) string {
	rel := filepath.Join(dnsTunnelingDir, name)

	base, err := fileutil.BasePath()
	if err != nil {
		return rel
	}

	return filepath.Join(base, rel)
}

// normalizeConfigName strips a .toml extension (case-insensitive) if present.
func normalizeConfigName(name string) string {
	if ext := filepath.Ext(name); strings.EqualFold(ext, ".toml") {
		return strings.TrimSuffix(name, ext)
	}

	return name
}

// tunnelConfig is the validation contract for tunnel configurations.
type tunnelConfig interface {
	Validate() map[string]error
}

// ConfigFile pairs a tunnel configuration with its on-disk identity.
// Each protocol exposes it under its historical name via a type alias.
type ConfigFile[C any] struct {
	Name      string
	Path      string
	CreatedAt time.Time
	Config    C
}

// configStore is a TOML-backed configuration directory shared by all
// tunnel services. C is the protocol configuration type.
type configStore[C tunnelConfig] struct {
	dir  string
	kind string
}

func newConfigStore[C tunnelConfig](dir, kind string) configStore[C] {
	return configStore[C]{dir: dir, kind: kind}
}

func (s *configStore[C]) configPath(name string) string {
	return filepath.Join(s.dir, normalizeConfigName(name)+".toml")
}

func (s *configStore[C]) SaveConfig(config C, name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("config name is required")
	}

	path := s.configPath(name)

	if fileutil.CheckFileExists(path) {
		return fmt.Errorf("config %q already exists", name)
	}

	if errs := config.Validate(); len(errs) > 0 {
		return fmt.Errorf("invalid %s config: %v", s.kind, errs)
	}

	if err := fileutil.WriteTOMLFile(path, config); err != nil {
		return fmt.Errorf("save %s config %q: %w", s.kind, name, err)
	}

	return nil
}

func (s *configStore[C]) EditConfig(config C, originalName string) error {
	if strings.TrimSpace(originalName) == "" {
		return fmt.Errorf("original config name is required")
	}

	path := s.configPath(originalName)

	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("config %q does not exist", originalName)
	}

	if errs := config.Validate(); len(errs) > 0 {
		return fmt.Errorf("invalid %s config: %v", s.kind, errs)
	}

	if err := fileutil.WriteTOMLFile(path, config); err != nil {
		return fmt.Errorf("edit %s config %q: %w", s.kind, originalName, err)
	}

	return nil
}

func (s *configStore[C]) LoadConfig(name string) (C, error) {
	config, err := fileutil.ReadTOMLFile[C](s.configPath(name))
	if err != nil {
		var zero C
		return zero, fmt.Errorf("load %s config %q: %w", s.kind, name, err)
	}

	return config, nil
}

// GetAllConfigFiles returns all valid configuration files in the store
// directory, skipping files that fail to load or validate.
func (s *configStore[C]) GetAllConfigFiles() ([]ConfigFile[C], error) {
	if err := fileutil.EnsureDir(s.dir); err != nil {
		return nil, err
	}

	files, err := fileutil.ListFiles(s.dir, func(name string, _ os.FileInfo) bool {
		return strings.EqualFold(filepath.Ext(name), ".toml")
	})
	if err != nil {
		return nil, err
	}

	configs := make([]ConfigFile[C], 0, len(files))

	for _, file := range files {
		cfg, err := s.LoadConfig(file.Name)
		if err != nil || len(cfg.Validate()) != 0 {
			continue
		}

		configs = append(configs, ConfigFile[C]{
			Name:      normalizeConfigName(file.Name),
			Path:      file.Path,
			CreatedAt: file.Info.ModTime(),
			Config:    cfg,
		})
	}

	return configs, nil
}

func (s *configStore[C]) ValidateAllConfigs() ([]ConfigValidationResult, error) {
	if err := fileutil.EnsureDir(s.dir); err != nil {
		return nil, err
	}

	files, err := fileutil.ListFiles(s.dir, func(name string, _ os.FileInfo) bool {
		return strings.EqualFold(filepath.Ext(name), ".toml")
	})
	if err != nil {
		return nil, err
	}

	results := make([]ConfigValidationResult, 0, len(files))

	for _, file := range files {
		cfg, err := s.LoadConfig(file.Name)
		if err != nil {
			results = append(results, ConfigValidationResult{
				File:   file,
				Errors: map[string]error{},
			})
			continue
		}

		validationErrors := cfg.Validate()
		if len(validationErrors) == 0 {
			continue
		}

		results = append(results, ConfigValidationResult{
			File:   file,
			Errors: validationErrors,
		})
	}

	return results, nil
}

func (s *configStore[C]) RenameConfig(oldName, newName string) error {
	oldPath := s.configPath(oldName)
	newPath := s.configPath(newName)

	if _, err := os.Stat(oldPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("config %q does not exist", oldName)
		}

		return fmt.Errorf("check current config: %w", err)
	}

	if _, err := os.Stat(newPath); err == nil {
		return fmt.Errorf("config %q already exists", newName)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("check destination config: %w", err)
	}

	return fileutil.RenameFile(oldPath, newPath)
}
