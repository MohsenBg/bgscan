package startup

import (
	"errors"
	"fmt"

	"github.com/MohsenBg/bgscan/internal/core"
	"github.com/MohsenBg/bgscan/internal/core/config"
	"github.com/MohsenBg/bgscan/internal/core/config/validate"
	"github.com/MohsenBg/bgscan/internal/core/dns"
	"github.com/MohsenBg/bgscan/internal/core/ffi/slipstream"
	"github.com/MohsenBg/bgscan/internal/core/xray"
	"github.com/MohsenBg/bgscan/internal/logger"
)

func checkConfigHealth(r *reporter, logs logger.Set) (*config.ScannerConfig, *config.Store) {
	store := config.NewStore(config.WithLogger(logs.Core))

	cfg, err := store.Load()
	if err != nil {
		r.critical("Configuration failed to load", err)
		return nil, nil
	}

	r.info("Initializing configuration system...")
	r.success("Config files loaded")
	r.info("Validating configuration values...")

	warns := validate.NormalizeAll(&cfg)
	if !warns.HasWarnings() {
		r.success("All values are valid")
		return &cfg, &store
	}

	r.warn("Some values were invalid and have been reset to defaults:")
	printSectionWarnings(r, "General", warns.General)
	printSectionWarnings(r, "Writer", warns.Writer)
	printSectionWarnings(r, "ICMP", warns.ICMP)
	printSectionWarnings(r, "TCP", warns.TCP)
	printSectionWarnings(r, "HTTP", warns.HTTP)
	printSectionWarnings(r, "Xray", warns.Xray)
	printSectionWarnings(r, "DNS", warns.DNS)
	r.warn("Completed with corrections ⚠")

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

func checkDNSClientHealth(
	r *reporter,
	label string,
	validateConfigs func() ([]dns.ConfigValidationResult, error),
) {
	r.infof("Checking %s client...", label)
	r.infof("Validating %s config files...", label)

	results, err := validateConfigs()
	if err != nil {
		r.errMsg(fmt.Sprintf("Failed to validate %s configs", label), err)
		return
	}

	printConfigValidationResults(r, label, results)
	r.success("Health check completed successfully")
}

func checkDNSTTHealth(r *reporter) {
	checkDNSClientHealth(r, "DNSTT", func() ([]dns.ConfigValidationResult, error) {
		return dns.NewDNSTTService().ValidateAllConfigs()
	})
}

func checkVayDNSHealth(r *reporter) {
	checkDNSClientHealth(r, "VayDNS", func() ([]dns.ConfigValidationResult, error) {
		return dns.NewVayDNSService().ValidateAllConfigs()
	})
}

func checkMasterDNSHealth(r *reporter) {
	checkDNSClientHealth(r, "MasterDNS", func() ([]dns.ConfigValidationResult, error) {
		return dns.NewMasterDNSService().ValidateAllConfigs()
	})
}

func checkStormDNSHealth(r *reporter) {
	checkDNSClientHealth(r, "StormDNS", func() ([]dns.ConfigValidationResult, error) {
		return dns.NewStormDNSService().ValidateAllConfigs()
	})
}

func checkTheFeedHealth(r *reporter) {
	checkDNSClientHealth(r, "TheFeed", func() ([]dns.ConfigValidationResult, error) {
		return dns.NewTheFeedService().ValidateAllConfigs()
	})
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

func checkLoggerHealth(r *reporter, logs logger.Set) {
	r.info("Verifying loggers...")

	if logs.Core == nil || logs.UI == nil || logs.Debug == nil {
		r.errMsg(
			"Loggers not initialized",
			errors.New("logger.Set is empty, main must call logger.NewSet before app.New"),
		)
		return
	}

	r.success("Loggers ready (core/ui/debug)")
	r.info("Registering probe schemas...")

	if err := core.Init(); err != nil {
		r.errMsg("Probe schema registration failed", err)
		return
	}

	r.success("Probe schemas registered")
	r.success("Health check completed successfully")
}

func checkXrayHealth(r *reporter, log *logger.Logger) {
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

	outbounds, err := xray.ListOutboundTemplates(log)
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
