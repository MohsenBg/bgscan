package startup

import (
	"bgscan/internal/core/config"
	"bgscan/internal/core/config/validate"
)

// checkConfigHealth initializes the configuration system, normalizes any
// bad values loaded from TOML, saves corrections back to disk, and reports
// what changed to the user.
func checkConfigHealth() {
	info("[INFO] Initializing configuration system...")

	if err := config.Init(); err != nil {
		errMsg("Failed to initialize configs", err)
		return
	}

	success("[SUCCESS] Config files loaded")

	info("[INFO] Validating configuration values...")

	warns := validate.NormalizeAll()

	if !warns.HasWarnings() {
		success("[CONFIG] All values are valid ✅")
		return
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
	if err := saveNormalized(warns); err != nil {
		errMsg("Failed to save corrected config", err)
		return
	}

	warn("[CONFIG] Completed with corrections ⚠")
}

// saveNormalized writes back only the sections that had corrections.
func saveNormalized(warns validate.AllWarnings) error {
	if len(warns.General) > 0 {
		if err := config.SaveGeneralConfig(config.GetGeneral()); err != nil {
			return err
		}
	}
	if len(warns.Writer) > 0 {
		if err := config.SaveWriterConfig(config.GetWriter()); err != nil {
			return err
		}
	}
	if len(warns.ICMP) > 0 {
		if err := config.SaveICMPConfig(config.GetICMP()); err != nil {
			return err
		}
	}
	if len(warns.TCP) > 0 {
		if err := config.SaveTCPConfig(config.GetTCP()); err != nil {
			return err
		}
	}
	if len(warns.HTTP) > 0 {
		if err := config.SaveHTTPConfig(config.GetHTTP()); err != nil {
			return err
		}
	}
	if len(warns.Xray) > 0 {
		if err := config.SaveXrayConfig(config.GetXray()); err != nil {
			return err
		}
	}
	if len(warns.DNS) > 0 {
		if err := config.SaveDNSConfig(config.GetDNS()); err != nil {
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
