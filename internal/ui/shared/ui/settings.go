package ui

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/MohsenBg/bgscan/internal/core/fileutil"
	"github.com/MohsenBg/bgscan/internal/logger"
	"github.com/MohsenBg/bgscan/internal/ui/theme"
)

const uiSettingsFile = "ui_settings.toml"

const DefaultMaxRow = 10_000

const DefaultLogLevel = "info"

type UISettings struct {
	ThemeName       string `toml:"theme" comment:"Color theme name. Empty = auto-detect."`
	SolidBackground bool   `toml:"solid_background" comment:"Paint the theme background color behind the app root view."`
	Animated        bool   `toml:"animated" comment:"Master switch for all TUI animations (current and future)."`
	MaxRow          uint32 `toml:"max_row" comment:"Maximum number of rows kept for display in result tables."`
	LogLevel        string `toml:"log_level" comment:"Minimum severity written to log files: debug, info, warn or error."`
}

func DefaultUISettings() UISettings {
	return UISettings{
		ThemeName:       "",
		SolidBackground: false,
		Animated:        true,
		MaxRow:          DefaultMaxRow,
		LogLevel:        DefaultLogLevel,
	}
}

// LoadSettings reads the settings file and creates it with defaults when missing.
// Invalid settings are normalized before being returned.
func LoadSettings(log logger.Logger) UISettings {
	path := uiSettingsPath()
	if path == "" {
		log.Warn("UI settings path is unavailable")
		return DefaultUISettings()
	}

	cfg, err := fileutil.ReadTOMLFile[UISettings](path)
	if err != nil {
		cfg = DefaultUISettings()

		if errors.Is(err, os.ErrNotExist) {
			if err := fileutil.WriteTOMLFile(path, cfg); err != nil {
				log.Warn("failed to create default UI settings: %v", err)
			}
		} else {
			log.Warn("failed to read UI settings, using defaults: %v", err)
		}

		return cfg
	}

	return normalize(cfg)
}

// SaveSettings writes normalized settings to disk. Write errors are ignored.
func SaveSettings(s UISettings) error {
	path := uiSettingsPath()
	if path == "" {
		return errors.New("settings path not found")
	}

	return fileutil.WriteTOMLFile(path, normalize(s))
}

// normalize repairs hand-edited values that would otherwise brick the UI.
func normalize(cfg UISettings) UISettings {
	if cfg.MaxRow == 0 {
		cfg.MaxRow = DefaultMaxRow
	}
	if _, err := logger.ParseLevel(cfg.LogLevel); err != nil {
		cfg.LogLevel = DefaultLogLevel
	}
	if _, err := theme.Get(cfg.ThemeName); err != nil {
		cfg.ThemeName = ""
	}
	return cfg
}

// uiSettingsPath returns the absolute path to the UI settings file, or ""
// if the base path cannot be determined.
func uiSettingsPath() string {
	base, err := fileutil.BasePath()
	if err != nil {
		return ""
	}

	return filepath.Join(base, "settings", uiSettingsFile)
}
