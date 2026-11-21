package scanner

import (
	"context"
	"fmt"
	"strings"

	"github.com/MohsenBg/bgscan/internal/core/config"
	"github.com/MohsenBg/bgscan/internal/core/dns"
	"github.com/MohsenBg/bgscan/internal/core/result"
	"github.com/MohsenBg/bgscan/internal/core/scanner/engine"
	"github.com/MohsenBg/bgscan/internal/core/scanner/probe"
	"github.com/MohsenBg/bgscan/internal/core/scanner/probe/dnsttprobe"
	"github.com/MohsenBg/bgscan/internal/core/scanner/probe/httpprobe"
	"github.com/MohsenBg/bgscan/internal/core/scanner/probe/icmpprobe"
	"github.com/MohsenBg/bgscan/internal/core/scanner/probe/masterdnsprobe"
	"github.com/MohsenBg/bgscan/internal/core/scanner/probe/resolveprobe"
	"github.com/MohsenBg/bgscan/internal/core/scanner/probe/slipstreamprobe"
	"github.com/MohsenBg/bgscan/internal/core/scanner/probe/stormdnsprobe"
	"github.com/MohsenBg/bgscan/internal/core/scanner/probe/tcpprobe"
	"github.com/MohsenBg/bgscan/internal/core/scanner/probe/thefeedprobe"
	"github.com/MohsenBg/bgscan/internal/core/scanner/probe/vaydnsprobe"
	"github.com/MohsenBg/bgscan/internal/core/scanner/probe/xrayprobe"
)

// BuildICMPStage creates an ICMP scan stage.
func (s *scanner) BuildICMPStage(ctx context.Context, hooks ...engine.ScanHooks) (StageConfig, error) {
	cfg := s.config.ICMP

	prb, err := icmpprobe.NewICMPProbe(icmpprobe.Options{
		Timeout: cfg.Timeout.Duration(),
		Tries:   cfg.Tries,
	})
	if err != nil {
		return StageConfig{}, fmt.Errorf("create ICMP probe: %w", err)
	}

	return s.buildStage(ctx, cfg.Workers, prb, cfg.OutputPrefix, icmpprobe.Schema, hooks...)
}

// BuildTCPStage creates a TCP scan stage.
func (s *scanner) BuildTCPStage(ctx context.Context, hooks ...engine.ScanHooks) (StageConfig, error) {
	cfg := s.config.TCP

	prb := tcpprobe.NewTCPProbe(
		fmt.Sprint(cfg.Port),
		cfg.Timeout.Duration(),
		cfg.Tries,
		s.log.With("probe", "tcp"),
	)

	return s.buildStage(ctx, cfg.Workers, prb, cfg.OutputPrefix, tcpprobe.Schema, hooks...)
}

// BuildHTTPStage creates an HTTP scan stage.
func (s *scanner) BuildHTTPStage(ctx context.Context, hooks ...engine.ScanHooks) (StageConfig, error) {
	cfg := s.config.HTTP

	set, err := httpprobe.NewSettings(cfg)
	if err != nil {
		return StageConfig{}, fmt.Errorf("create HTTP settings: %w", err)
	}

	mode, err := httpprobe.NewMode(cfg, set)
	if err != nil {
		return StageConfig{}, fmt.Errorf("create HTTP mode: %w", err)
	}

	prb, err := s.newHTTPProbe(cfg, set, mode)
	if err != nil {
		return StageConfig{}, err
	}

	return s.buildStage(ctx, cfg.Workers, prb, cfg.OutputPrefix, httpprobe.Schema, hooks...)
}

func (s *scanner) newHTTPProbe(
	cfg config.HTTPConfig,
	settings httpprobe.Settings,
	mode httpprobe.Mode,
) (probe.Probe, error) {
	if isHTTP3(cfg.Version) {
		prb, err := httpprobe.NewHTTP3Probe(
			settings,
			mode,
			cfg.AcceptedStatusCodes,
			s.log.With("probe", "http3"),
		)
		if err != nil {
			return nil, fmt.Errorf("create HTTP/3 probe: %w", err)
		}
		return prb, nil
	}

	return httpprobe.NewHTTPProbe(
		settings,
		mode,
		cfg.AcceptedStatusCodes,
		s.log.With("probe", "http"),
	), nil
}

func isHTTP3(version string) bool {
	version = strings.ToLower(version)
	return version == "h3" || version == "http3"
}

// BuildXrayStage creates the stages required for an Xray scan against
// template. If the Xray config specifies a pre-scan type (tcp, icmp, or
// http), a stage for that pre-scan is prepended so it runs ahead of the
// Xray probe in the chain. A pre-scan type of "none" (or empty) skips this
// and returns only the Xray stage.
func (s *scanner) BuildXrayStage(
	ctx context.Context,
	template string,
	hooks ...engine.ScanHooks,
) ([]StageConfig, error) {
	cfg := s.config.Xray

	stages := make([]StageConfig, 0, 2)

	preScan, err := s.buildXrayPreScanStage(ctx, cfg.PreScanType)
	if err != nil {
		return nil, err
	}
	if preScan != nil {
		stages = append(stages, *preScan)
	}

	prb, err := xrayprobe.NewXrayProbe(
		&cfg,
		template,
		s.pm,
		xrayprobe.WithLogger(s.log.With("probe", "xray")),
	)
	if err != nil {
		return nil, fmt.Errorf("create Xray probe: %w", err)
	}

	stage, err := s.buildStage(ctx, cfg.Workers, prb, cfg.OutputPrefix, xrayprobe.Schema, hooks...)
	if err != nil {
		return nil, err
	}
	stages = append(stages, stage)

	return stages, nil
}

// buildXrayPreScanStage builds the optional connectivity pre-scan stage
// that runs ahead of an Xray stage, based on preScanType. It returns a nil
// stage (no error) when preScanType is "none" or empty.
func (s *scanner) buildXrayPreScanStage(
	ctx context.Context,
	preScanType string,
) (*StageConfig, error) {
	switch preScanType {
	case "", "none":
		return nil, nil

	case "tcp":
		stage, err := s.BuildTCPStage(ctx)
		if err != nil {
			return nil, fmt.Errorf("create Xray TCP pre-scan: %w", err)
		}
		return &stage, nil

	case "icmp":
		stage, err := s.BuildICMPStage(ctx)
		if err != nil {
			return nil, fmt.Errorf("create Xray ICMP pre-scan: %w", err)
		}
		return &stage, nil

	case "http":
		stage, err := s.BuildHTTPStage(ctx)
		if err != nil {
			return nil, fmt.Errorf("create Xray HTTP pre-scan: %w", err)
		}
		return &stage, nil

	default:
		return nil, fmt.Errorf("unsupported Xray pre_scan_type %q", preScanType)
	}
}

// BuildResolveStage creates a DNS resolver scan stage.
func (s *scanner) BuildResolveStage(ctx context.Context, hooks ...engine.ScanHooks) (StageConfig, error) {
	cfg := s.config.DNS.Resolver
	return s.buildResolverStage(ctx, cfg, hooks...)
}

func (s *scanner) buildResolverStage(
	ctx context.Context,
	cfg config.ResolverConfig,
	hooks ...engine.ScanHooks,
) (StageConfig, error) {
	rcodes := make([]uint16, 0, len(cfg.AcceptedRCodes))
	for _, code := range cfg.AcceptedRCodes {
		rcodes = append(rcodes, uint16(dns.ParseDNSRcode(code)))
	}

	prb := resolveprobe.NewResolverProbe(&resolveprobe.DNSRequest{
		Domain:          cfg.Domain,
		Port:            cfg.Port,
		RandomSubdomain: cfg.RandomSubdomain,
		DpiCheck:        cfg.DPI.Enabled,
		DpiTimeout:      cfg.DPI.Timeout.Duration(),
		DpiTries:        cfg.DPI.Tries,
		Edns0Size:       cfg.EDNSBufSize,
		CheckTypes:      cfg.CheckTypes,
		AcceptedRcodes:  rcodes,
		Timeout:         cfg.Timeout.Duration(),
		Transport:       dns.ParseResolverType(cfg.Transport),
		Tries:           cfg.Tries,
	})

	return s.buildStage(ctx, cfg.Workers, prb, cfg.OutputPrefix, resolveprobe.Schema, hooks...)
}

// BuildDNSTTStage creates the stages required for a DNSTT scan.
func (s *scanner) BuildDNSTTStage(
	ctx context.Context,
	configName string,
	hooks ...engine.ScanHooks,
) ([]StageConfig, error) {

	tunCfg := s.config.DNS.DNSTunneling

	dnsttCfg, err := s.dnsttService.LoadConfig(configName)
	if err != nil {
		return nil, fmt.Errorf("load DNSTT config: %w", err)
	}

	stages := make([]StageConfig, 0, 2)

	preStage, err := s.resolverPreStage(ctx, func(rc *config.ResolverConfig) {
		rc.Port = dnsttCfg.ResolverPort
		rc.Transport = string(dnsttCfg.ResolverType)
		rc.Domain = dnsttCfg.Domain
	})
	if err != nil {
		return nil, err
	}
	if preStage != nil {
		stages = append(stages, *preStage)
	}

	prb, err := dnsttprobe.NewDNSTTProbe(
		dnsttCfg,
		tunCfg.Timeout.Duration(),
		dnsttprobe.WithTries(tunCfg.Tries),
		dnsttprobe.WithLogger(s.log.With("probe", "dnstt")),
	)
	if err != nil {
		return nil, fmt.Errorf("create DNSTT probe: %w", err)
	}

	stage, err := s.buildStage(ctx, tunCfg.Workers, prb, tunCfg.OutputPrefix, dnsttprobe.Schema, hooks...)
	if err != nil {
		return nil, err
	}
	stages = append(stages, stage)

	return stages, nil
}

// BuildSlipStreamStage creates the stages required for a Slipstream scan.
func (s *scanner) BuildSlipStreamStage(
	ctx context.Context,
	configName string,
	hooks ...engine.ScanHooks,
) ([]StageConfig, error) {
	tunCfg := s.config.DNS.DNSTunneling

	slipstreamCfg, err := s.slipstreamService.LoadConfig(configName)
	if err != nil {
		return nil, fmt.Errorf("load Slipstream config: %w", err)
	}

	stages := make([]StageConfig, 0, 2)

	preStage, err := s.resolverPreStage(ctx, func(rc *config.ResolverConfig) {
		rc.Port = slipstreamCfg.ResolverPort
		rc.Transport = string(dns.ResolverTypeUDP)
		rc.Domain = slipstreamCfg.Domain
	})
	if err != nil {
		return nil, err
	}
	if preStage != nil {
		stages = append(stages, *preStage)
	}

	prb, err := slipstreamprobe.NewSlipstreamProbe(
		slipstreamCfg,
		tunCfg.Timeout.Duration(),
		s.pm,
		slipstreamprobe.WithSlipstreamService(s.slipstreamService),
		slipstreamprobe.WithTries(tunCfg.Tries),
		slipstreamprobe.WithLogger(s.log.With("probe", "slipstream")),
	)
	if err != nil {
		return nil, fmt.Errorf("create Slipstream probe: %w", err)
	}

	stage, err := s.buildStage(ctx, tunCfg.Workers, prb, tunCfg.OutputPrefix, slipstreamprobe.Schema, hooks...)
	if err != nil {
		return nil, err
	}
	stages = append(stages, stage)

	return stages, nil
}

// BuildVayDNSStage creates the stages required for a VayDNS scan.
func (s *scanner) BuildVayDNSStage(
	ctx context.Context,
	configName string,
	hooks ...engine.ScanHooks,
) ([]StageConfig, error) {
	tunCfg := s.config.DNS.DNSTunneling

	vaydnsCfg, err := s.vaydnsService.LoadConfig(configName)
	if err != nil {
		return nil, fmt.Errorf("load VayDNS config: %w", err)
	}

	stages := make([]StageConfig, 0, 2)

	preStage, err := s.resolverPreStage(ctx, func(rc *config.ResolverConfig) {
		rc.Port = vaydnsCfg.ResolverPort
		rc.Transport = string(vaydnsCfg.ResolverType)
		rc.Domain = vaydnsCfg.Domain
	})
	if err != nil {
		return nil, err
	}
	if preStage != nil {
		stages = append(stages, *preStage)
	}

	prb, err := vaydnsprobe.NewVayDNSProbe(
		vaydnsCfg,
		tunCfg.Timeout.Duration(),
		vaydnsprobe.WithVayDNSService(s.vaydnsService),
		vaydnsprobe.WithTries(tunCfg.Tries),
		vaydnsprobe.WithLogger(s.log.With("probe", "vaydns")),
	)
	if err != nil {
		return nil, fmt.Errorf("create VayDNS probe: %w", err)
	}

	stage, err := s.buildStage(ctx, tunCfg.Workers, prb, tunCfg.OutputPrefix, vaydnsprobe.Schema, hooks...)
	if err != nil {
		return nil, err
	}
	stages = append(stages, stage)

	return stages, nil
}

// BuildMasterDNSStage creates the stages required for a MasterDNS scan.
func (s *scanner) BuildMasterDNSStage(
	ctx context.Context,
	configName string,
	hooks ...engine.ScanHooks,
) ([]StageConfig, error) {
	tunCfg := s.config.DNS.DNSTunneling

	masterCfg, err := s.masterDNSService.LoadConfig(configName)
	if err != nil {
		return nil, fmt.Errorf("load MasterDNS config: %w", err)
	}

	stages := make([]StageConfig, 0, 2)

	preStage, err := s.resolverPreStage(ctx, func(rc *config.ResolverConfig) {
		rc.Port = uint16(masterCfg.ResolverPort)
		rc.Transport = string(dns.ResolverTypeUDP)
		rc.Domain = masterCfg.Domain
	})
	if err != nil {
		return nil, err
	}
	if preStage != nil {
		stages = append(stages, *preStage)
	}

	prb, err := masterdnsprobe.NewMasterDNSProbe(
		masterCfg,
		tunCfg.Timeout.Duration(),
		s.pm,
		masterdnsprobe.WithMasterDNSService(s.masterDNSService),
		masterdnsprobe.WithTries(tunCfg.Tries),
		masterdnsprobe.WithLogger(s.log.With("probe", "masterdns")),
	)
	if err != nil {
		return nil, fmt.Errorf("create MasterDNS probe: %w", err)
	}

	stage, err := s.buildStage(ctx, tunCfg.Workers, prb, tunCfg.OutputPrefix, masterdnsprobe.Schema, hooks...)
	if err != nil {
		return nil, err
	}
	stages = append(stages, stage)

	return stages, nil
}

// BuildStormDNSStage creates the stages required for a StormDNS scan.
func (s *scanner) BuildStormDNSStage(
	ctx context.Context,
	configName string,
	hooks ...engine.ScanHooks,
) ([]StageConfig, error) {
	tunCfg := s.config.DNS.DNSTunneling

	stormCfg, err := s.stormDNSService.LoadConfig(configName)
	if err != nil {
		return nil, fmt.Errorf("load StormDNS config: %w", err)
	}

	stages := make([]StageConfig, 0, 2)

	preStage, err := s.resolverPreStage(ctx, func(rc *config.ResolverConfig) {
		rc.Port = uint16(stormCfg.ResolverPort)
		rc.Transport = string(dns.ResolverTypeUDP)
		rc.Domain = stormCfg.Domain
	})
	if err != nil {
		return nil, err
	}
	if preStage != nil {
		stages = append(stages, *preStage)
	}

	prb, err := stormdnsprobe.NewStormDNSProbe(
		stormCfg,
		tunCfg.Timeout.Duration(),
		s.pm,
		stormdnsprobe.WithStormDNSService(s.stormDNSService),
		stormdnsprobe.WithTries(tunCfg.Tries),
		stormdnsprobe.WithLogger(s.log.With("probe", "stormdns")),
	)
	if err != nil {
		return nil, fmt.Errorf("create StormDNS probe: %w", err)
	}

	stage, err := s.buildStage(ctx, tunCfg.Workers, prb, tunCfg.OutputPrefix, stormdnsprobe.Schema, hooks...)
	if err != nil {
		return nil, err
	}
	stages = append(stages, stage)

	return stages, nil
}

// BuildTheFeedStage creates the stages required for a TheFeed scan.
func (s *scanner) BuildTheFeedStage(
	ctx context.Context,
	configName string,
	hooks ...engine.ScanHooks,
) ([]StageConfig, error) {
	tunCfg := s.config.DNS.DNSTunneling

	thefeedCfg, err := s.thefeedService.LoadConfig(configName)
	if err != nil {
		return nil, fmt.Errorf("load TheFeed config: %w", err)
	}

	stages := make([]StageConfig, 0, 2)

	preStage, err := s.resolverPreStage(ctx, func(rc *config.ResolverConfig) {
		rc.Port = thefeedCfg.ResolverPort
		rc.Transport = string(thefeedCfg.ResolverType)
		rc.Domain = thefeedCfg.Domain
	})
	if err != nil {
		return nil, err
	}
	if preStage != nil {
		stages = append(stages, *preStage)
	}

	prb, err := thefeedprobe.NewTheFeedProbe(
		thefeedCfg,
		tunCfg.Timeout.Duration(),
		thefeedprobe.WithTheFeedService(s.thefeedService),
		thefeedprobe.WithTries(tunCfg.Tries),
	)
	if err != nil {
		return nil, fmt.Errorf("create TheFeed probe: %w", err)
	}

	stage, err := s.buildStage(ctx, tunCfg.Workers, prb, tunCfg.OutputPrefix, thefeedprobe.Schema, hooks...)
	if err != nil {
		return nil, err
	}
	stages = append(stages, stage)

	return stages, nil
}

// --- Shared helpers ---

// buildStage creates the result writer for a probe and assembles its stage.
func (s *scanner) buildStage(
	ctx context.Context,
	workers int,
	prb probe.Probe,
	prefix string,
	schema result.ResultSchema,
	hooks ...engine.ScanHooks,
) (StageConfig, error) {
	writer, err := s.newWriter(ctx, prefix, schema)
	if err != nil {
		return StageConfig{}, err
	}
	return s.newStage(workers, prb, writer, hooks...), nil
}

// resolverPreStage builds the optional resolver check stage that runs ahead
// of the DNS tunnel stages. adapt (which may be nil) customizes the resolver
// config from the loaded tunnel config when AdaptiveResolver is enabled.
// It returns (nil, nil) when the resolver check is disabled.
func (s *scanner) resolverPreStage(
	ctx context.Context,
	adapt func(*config.ResolverConfig),
) (*StageConfig, error) {
	tunCfg := s.config.DNS.DNSTunneling
	if !tunCfg.CheckDNSResolver {
		return nil, nil
	}

	resolverCfg := s.config.DNS.Resolver
	if tunCfg.AdaptiveResolver && adapt != nil {
		adapt(&resolverCfg)
	}

	stage, err := s.buildResolverStage(ctx, resolverCfg)
	if err != nil {
		return nil, err
	}
	return &stage, nil
}

func (s *scanner) newWriter(ctx context.Context, prefix string, schema result.ResultSchema) (result.Writer, error) {
	writer, err := s.writerFactory(ctx, result.WriterOptions{
		ResultPrefix: prefix,
		Schema:       schema,
		Config:       s.config.Writer,
		Log:          s.log,
	})
	if err != nil {
		return nil, fmt.Errorf("create %s result writer: %w", schema.Name, err)
	}

	return writer, nil
}

func (s *scanner) newStage(
	workers int,
	prb probe.Probe,
	writer result.Writer,
	hooks ...engine.ScanHooks,
) StageConfig {
	stage := StageConfig{
		Workers: workers,
		Probe:   prb,
		Writer:  writer,
	}

	if len(hooks) > 0 {
		stage.AddHooks(hooks[0])
	}

	return stage
}
