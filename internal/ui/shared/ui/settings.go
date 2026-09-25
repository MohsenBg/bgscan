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

const DefaultMaxRow = 10000

const DefaultLogLevel = "info"

type UISettings struct {
	ThemeName       string       `toml:"theme" comment:"Color theme name. Empty = auto-detect."`
	Theme           *theme.Theme `toml:"-"`
	SolidBackground bool         `toml:"solid_background" comment:"Paint the theme background color behind the app root view."`
	MaxRow          uint32       `toml:"max_row" comment:"Maximum number of rows kept for display in result tables."`
	LogLevel        string       `toml:"log_level" comment:"Minimum severity written to log files: debug, info, warn or error."`
}

func DefaultUISettings() UISettings {
	theme, _ := theme.Get("")
	return UISettings{
		ThemeName:       "",
		Theme:           theme,
		SolidBackground: false,
		MaxRow:          DefaultMaxRow,
		LogLevel:        DefaultLogLevel,
	}
}

// LoadSettings reads the settings file, creating it with defaults when
// missing. A corrupt file falls back to defaults without overwriting the
// user's file; invalid fields (zero max rows, unknown log level) are
// normalized.
func LoadSettings(log *logger.Logger) UISettings {
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

// SaveSettings persists s to disk (normalized), so the file never
// carries values that would brick the UI. Best effort: a failed write
// must never break the caller.
func SaveSettings(s UISettings) {
	path := uiSettingsPath()
	if path == "" {
		return
	}

	_ = fileutil.WriteTOMLFile(path, normalize(s))
}

// normalize repairs hand-edited values that would otherwise brick the UI.
func normalize(cfg UISettings) UISettings {
	if cfg.MaxRow == 0 {
		cfg.MaxRow = DefaultMaxRow
	}
	if _, err := logger.ParseLevel(cfg.LogLevel); err != nil {
		cfg.LogLevel = DefaultLogLevel
	}
	// Resolve the active palette from ThemeName. Unknown names fall back
	// to auto-detect and reset ThemeName so the file never carries a
	// theme that would brick the UI.
	th, err := theme.Get(cfg.ThemeName)
	if err != nil {
		cfg.ThemeName = ""
		th, _ = theme.Get("")
	}
	cfg.Theme = th
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
