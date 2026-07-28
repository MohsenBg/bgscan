package startup

import (
	"errors"

	"bgscan/internal/core/config"
	"bgscan/internal/core/config/validate"
)

// checkConfigHealth initializes the configuration system, normalizes any
// bad values loaded from TOML, saves corrections back to disk, and reports
// what changed to the user.
func checkConfigHealth(cfg *config.ScannerConfig, store *config.Store) error {
	info("[INFO] Initializing configuration system...")

	if cfg == nil {
		err := errors.New("config is nil")
		errMsg("config is nil", err)
		return err
	}

	success("[SUCCESS] Config files loaded")

	info("[INFO] Validating configuration values...")

	warns := validate.NormalizeAll(cfg)

	if !warns.HasWarnings() {
		success("[CONFIG] All values are valid ✅")
		return nil
	}

	// Print every correction made.
	warn("[CONFIG] Some values were invalid and have been reset to defaults:")
	printSectionWarnings("General", warns.General)
	printSectionWarnings("Writer", warns.Writer)
	printSectionWarnings("ICMP", warns.ICMP)
	printSectionWarnings("TCP", warns.TCP)
	printSectionWarnings("HTTP", warns.HTTP)
	printSectionWarnings("Xray", warns.Xray)
	printSectionWarnings("DNS", warns.DNS)

	// Save corrected values back so TOML reflects reality.
	info("[INFO] Saving corrected values back to disk...")
	if err := saveNormalized(warns, cfg, store); err != nil {
		errMsg("Failed to save corrected config", err)
		return err
	}

	warn("[CONFIG] Completed with corrections ⚠")
	return nil
}

// saveNormalized writes back only the sections that had corrections.
func saveNormalized(warns validate.AllWarnings, cfg *config.ScannerConfig, store *config.Store) error {
	if len(warns.General) > 0 {
		if err := store.SaveGeneral(cfg.General); err != nil {
			return err
		}
	}
	if len(warns.Writer) > 0 {
		if err := store.SaveWriter(cfg.Writer); err != nil {
			return err
		}
	}
	if len(warns.ICMP) > 0 {
		if err := store.SaveICMP(cfg.ICMP); err != nil {
			return err
		}
	}
	if len(warns.TCP) > 0 {
		if err := store.SaveTCP(cfg.TCP); err != nil {
			return err
		}
	}
	if len(warns.HTTP) > 0 {
		if err := store.SaveHTTP(cfg.HTTP); err != nil {
			return err
		}
	}
	if len(warns.Xray) > 0 {
		if err := store.SaveXray(cfg.Xray); err != nil {
			return err
		}
	}
	if len(warns.DNS) > 0 {
		if err := store.SaveDNS(cfg.DNS); err != nil {
			return err
		}
	}
	return nil
}

// printSectionWarnings logs all warnings for one config section.
func printSectionWarnings(section string, warns []validate.Warning) {
	for _, w := range warns {
		warnf("  [%s] %s", section, w.String())
	}
}
