package startup

import (
	"fmt"

	"github.com/MohsenBg/bgscan/internal/core"
	"github.com/MohsenBg/bgscan/internal/core/config"
	"github.com/MohsenBg/bgscan/internal/core/config/validate"
	"github.com/MohsenBg/bgscan/internal/core/dns"
	"github.com/MohsenBg/bgscan/internal/core/ffi/slipstream"
	"github.com/MohsenBg/bgscan/internal/core/xray"
	"github.com/MohsenBg/bgscan/internal/logger"
)

func checkConfigHealth(r *reporter) (*config.ScannerConfig, *config.Store) {
	store := config.NewStore()
	cfg, err := store.Load()
	if err != nil {
		r.critical("Configuration failed to load", err)
	}

	r.info("Initializing configuration system...")
	r.success("Config files loaded")

	r.info("Validating configuration values...")
	warns := validate.NormalizeAll(&cfg)
	if !warns.HasWarnings() {
		r.success("All values are valid")
	} else {
		r.warn("Some values were invalid and have been reset to defaults:")
		printSectionWarnings(r, "General", warns.General)
		printSectionWarnings(r, "Writer", warns.Writer)
		printSectionWarnings(r, "ICMP", warns.ICMP)
		printSectionWarnings(r, "TCP", warns.TCP)
		printSectionWarnings(r, "HTTP", warns.HTTP)
		printSectionWarnings(r, "Xray", warns.Xray)
		printSectionWarnings(r, "DNS", warns.DNS)
		r.warn("Completed with corrections ⚠")
	}

	return &cfg, &store
}

func printSectionWarnings(r *reporter, section string, warns []validate.Warning) {
	for _, w := range warns {
		r.warnf("  [%s] %s", section, w.String())
	}
}

func printConfigValidationResults(r *reporter, label string, results []dns.ConfigValidationResult) {
	if len(results) == 0 {
		r.successf("All %s config files are valid", label)
		return
	}

	r.warnf("%d %s config file(s) have problems:", len(results), label)
	for _, res := range results {
		if len(res.Errors) == 0 {
			r.warnf("  [%s] failed to load (no further detail available)", res.File.Name)
			continue
		}
		for field, fieldErr := range res.Errors {
			r.warnf("  [%s] %s: %v", res.File.Name, field, fieldErr)
		}
	}
}

func checkDNSTTHealth(r *reporter) {
	r.info("Checking DNSTT client...")

	r.info("Validating DNSTT config files...")
	results, err := dns.NewDNSTTService().ValidateAllConfigs()
	if err != nil {
		r.errMsg("Failed to validate DNSTT configs", err)
		return
	}
	printConfigValidationResults(r, "DNSTT", results)

	r.success("Health check completed successfully")
}

func checkVayDNSHealth(r *reporter) {
	r.info("Checking VayDNS client...")

	r.info("Validating VayDNS config files...")
	results, err := dns.NewVayDNSService().ValidateAllConfigs()
	if err != nil {
		r.errMsg("Failed to validate VayDNS configs", err)
		return
	}
	printConfigValidationResults(r, "VayDNS", results)

	r.success("Health check completed successfully")
}

func checkMasterDNSHealth(r *reporter) {
	r.info("Checking MasterDNS client...")

	r.info("Validating MasterDNS config files...")
	results, err := dns.NewMasterDNSService().ValidateAllConfigs()
	if err != nil {
		r.errMsg("Failed to validate MasterDNS configs", err)
		return
	}
	printConfigValidationResults(r, "MasterDNS", results)

	r.success("Health check completed successfully")
}

func checkStormDNSHealth(r *reporter) {
	r.info("Checking StormDNS client...")

	r.info("Validating StormDNS config files...")
	results, err := dns.NewStormDNSService().ValidateAllConfigs()
	if err != nil {
		r.errMsg("Failed to validate StormDNS configs", err)
		return
	}
	printConfigValidationResults(r, "StormDNS", results)

	r.success("Health check completed successfully")
}

func checkTheFeedHealth(r *reporter) {
	r.info("Checking TheFeed client...")

	r.info("Validating TheFeed config files...")
	results, err := dns.NewTheFeedService().ValidateAllConfigs()
	if err != nil {
		r.errMsg("Failed to validate TheFeed configs", err)
		return
	}
	printConfigValidationResults(r, "TheFeed", results)

	r.success("Health check completed successfully")
}

func checkSlipstreamHealth(r *reporter) {
	r.info("Finding Slipstream library...")
	path, err := slipstream.FindLibSlipstream()
	if err != nil {
		r.binaryMissing("Slipstream", "libslipstream-client")
		r.errMsg("Library lookup error", err)
		return
	}
	r.successf("Slipstream found at: %s", path)

	r.info("Verifying Slipstream library...")
	version, err := slipstream.VerifyLibSlipstream()
	if err != nil {
		r.errMsg("Slipstream library validation failed", err)
		return
	}
	r.successf("Slipstream library verified (version %s)", version)

	r.info("Validating Slipstream config files...")
	srv, err := dns.NewSlipstreamService()
	if err != nil {
		r.errMsg("Failed to create Slipstream service", err)
		return
	}
	results, err := srv.ValidateAllConfigs()
	if err != nil {
		r.errMsg("Failed to validate Slipstream configs", err)
		return
	}
	printConfigValidationResults(r, "Slipstream", results)

	r.success("Health check completed successfully")
}

func checkLoggerHealth(r *reporter) {
	r.info("Initializing loggers...")

	if err := logger.InitCore(); err != nil {
		r.errMsg("Core logger initialization failed", err)
		return
	}
	r.success("Core logger initialized")

	if err := logger.InitUI(); err != nil {
		r.errMsg("UI logger initialization failed", err)
		return
	}
	r.success("UI logger initialized")

	if err := logger.InitDebug(); err != nil {
		r.errMsg("Debug logger initialization failed", err)
		return
	}
	r.success("Debug logger initialized")

	r.info("Registering probe schemas...")
	if err := core.Init(); err != nil {
		r.errMsg("Probe schema registration failed", err)
		return
	}
	r.success("Probe schemas registered")

	r.success("Health check completed successfully")
}

func checkXrayHealth(r *reporter) {
	r.info("Checking embedded Xray core...")
	svc := xray.NewXrayService()

	r.info("Checking Xray core version...")
	version := svc.Version()
	if version == "" {
		r.errMsg("Failed to retrieve Xray version", fmt.Errorf("empty version string"))
		return
	}
	r.successf("Xray core version: %s", version)

	r.info("Searching for configuration templates...")
	outbounds, err := xray.ListOutboundTemplates()
	if err != nil {
		r.errMsg("Failed to retrieve outbounds", err)
		return
	}
	r.infof("Found %d outbound templates.", len(outbounds))

	for _, outbound := range outbounds {
		r.infof("Validating outbound: %s", outbound.Name)

		if err := xray.ValidateOutbound(outbound.Name); err != nil {
			r.errMsg("Outbound validation failed: "+outbound.Name, err)
			continue
		}

		r.successf("%s OK", outbound.Name)
	}

	r.success("Health check completed successfully")
}

func checkAppHealth(r *reporter) {
	r.wait("Press Enter to continue...")
}
