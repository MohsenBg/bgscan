package scanner

import (
	"context"
	"fmt"

	"bgscan/internal/core/config"
	"bgscan/internal/core/dns"
	"bgscan/internal/core/result"
	"bgscan/internal/core/scanner/engine"
	"bgscan/internal/core/scanner/probe"
	"bgscan/internal/core/scanner/probe/dnsttprobe"
	"bgscan/internal/core/scanner/probe/httpprobe"
	"bgscan/internal/core/scanner/probe/icmpprobe"
	"bgscan/internal/core/scanner/probe/resolveprobe"
	"bgscan/internal/core/scanner/probe/slipstreamprobe"
	"bgscan/internal/core/scanner/probe/tcpprobe"
	"bgscan/internal/core/scanner/probe/vaydnsprobe"
	"bgscan/internal/core/scanner/probe/xrayprobe"
)

// BuildICMPStage creates an ICMP scan stage.
func (s *scanner) BuildICMPStage(
	ctx context.Context,
	hooks ...engine.ScanHooks,
) (StageConfig, error) {
	cfg := s.config.ICMP

	prb, err := icmpprobe.NewICMPProbe(icmpprobe.Options{
		Timeout: cfg.Timeout.Duration(),
		Tries:   cfg.Tries,
	})
	if err != nil {
		return StageConfig{}, fmt.Errorf("create ICMP probe: %w", err)
	}

	writer, err := s.newWriter(
		ctx,
		cfg.OutputPrefix,
		icmpprobe.Schema,
	)
	if err != nil {
		return StageConfig{}, err
	}

	return s.newStage(cfg.Workers, prb, writer, hooks...), nil
}

// BuildTCPStage creates a TCP scan stage.
func (s *scanner) BuildTCPStage(
	ctx context.Context,
	hooks ...engine.ScanHooks,
) (StageConfig, error) {
	cfg := s.config.TCP

	prb := tcpprobe.NewTCPProbe(
		fmt.Sprint(cfg.Port),
		cfg.Timeout.Duration(),
		cfg.Tries,
	)

	writer, err := s.newWriter(
		ctx,
		cfg.OutputPrefix,
		tcpprobe.Schema,
	)
	if err != nil {
		return StageConfig{}, err
	}

	return s.newStage(cfg.Workers, prb, writer, hooks...), nil
}

// BuildHTTPStage creates an HTTP scan stage.
func (s *scanner) BuildHTTPStage(
	ctx context.Context,
	hooks ...engine.ScanHooks,
) (StageConfig, error) {
	cfg := s.config.HTTP

	req, err := httpprobe.NewHTTPRequestFromConfig(cfg)
	if err != nil {
		return StageConfig{}, fmt.Errorf("create HTTP request: %w", err)
	}

	prb, err := s.newHTTPProbe(cfg, *req)
	if err != nil {
		return StageConfig{}, err
	}

	writer, err := s.newWriter(
		ctx,
		cfg.OutputPrefix,
		httpprobe.Schema,
	)
	if err != nil {
		return StageConfig{}, err
	}

	return s.newStage(cfg.Workers, prb, writer, hooks...), nil
}

func (s *scanner) newHTTPProbe(
	cfg config.HTTPConfig,
	req httpprobe.HTTPRequest,
) (probe.Probe, error) {
	if isHTTP3(cfg.Version) {
		prb, err := httpprobe.NewHTTP3Probe(
			req,
			cfg.AcceptedStatusCodes,
		)
		if err != nil {
			return nil, fmt.Errorf("create HTTP/3 probe: %w", err)
		}
		return prb, nil
	}

	return httpprobe.NewHTTPProbe(
		req,
		cfg.AcceptedStatusCodes,
	), nil
}

func isHTTP3(version string) bool {
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
	)
	if err != nil {
		return nil, fmt.Errorf("create Xray probe: %w", err)
	}

	writer, err := s.newWriter(
		ctx,
		cfg.OutputPrefix,
		xrayprobe.Schema,
	)
	if err != nil {
		return nil, err
	}

	stages = append(stages, s.newStage(cfg.Workers, prb, writer, hooks...))

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
		return nil, fmt.Errorf(
			"unsupported Xray pre_scan_type %q (allowed: %v)",
			preScanType,
			allowedPreScanTypes,
		)
	}
}

// BuildResolveStage creates a DNS resolver scan stage.
func (s *scanner) BuildResolveStage(
	ctx context.Context,
	hooks ...engine.ScanHooks,
) (StageConfig, error) {
	cfg := s.config.DNS.Resolver
	return s.buildResolverStage(ctx, cfg, hooks...)
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

	if tunCfg.CheckDNSResolver {
		resolverCfg := s.config.DNS.Resolver
		resolverCfg.Port = dnsttCfg.ResolverPort
		resolverCfg.Transport = string(dnsttCfg.ResolverType)

		stage, err := s.buildResolverStage(
			ctx,
			resolverCfg,
		)
		if err != nil {
			return nil, err
		}

		stages = append(stages, stage)
	}

	prb, err := dnsttprobe.NewDNSTTProbe(
		dnsttCfg,
		tunCfg.Timeout.Duration(),
	)
	if err != nil {
		return nil, fmt.Errorf("create DNSTT probe: %w", err)
	}

	writer, err := s.newWriter(
		ctx,
		tunCfg.OutputPrefix,
		dnsttprobe.Schema,
	)
	if err != nil {
		return nil, err
	}

	stages = append(
		stages,
		s.newStage(tunCfg.Workers, prb, writer, hooks...),
	)

	return stages, nil
}

func (s *scanner) buildResolverStage(
	ctx context.Context,
	cfg config.ResolverConfig,
	hooks ...engine.ScanHooks,
) (StageConfig, error) {
	rcodes := make([]uint16, 0, len(cfg.AcceptedRCodes))
	for _, code := range cfg.AcceptedRCodes {
		rcodes = append(
			rcodes,
			uint16(dns.ParseDNSRcode(code)),
		)
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

	writer, err := s.newWriter(
		ctx,
		cfg.OutputPrefix,
		resolveprobe.Schema,
	)
	if err != nil {
		return StageConfig{}, err
	}

	return s.newStage(cfg.Workers, prb, writer, hooks...), nil
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

	if tunCfg.CheckDNSResolver {
		resolverCfg := s.config.DNS.Resolver
		resolverCfg.Port = slipstreamCfg.ResolverPort
		resolverCfg.Transport = string(dns.ResolverTypeUDP)

		stage, err := s.buildResolverStage(
			ctx,
			resolverCfg,
		)
		if err != nil {
			return nil, err
		}

		stages = append(stages, stage)
	}

	prb, err := slipstreamprobe.NewSlipstreamProbe(
		slipstreamCfg,
		tunCfg.Timeout.Duration(),
		s.pm,
		slipstreamprobe.WithSlipstreamService(s.slipstreamService),
	)
	if err != nil {
		return nil, fmt.Errorf("create Slipstream probe: %w", err)
	}

	writer, err := s.newWriter(
		ctx,
		tunCfg.OutputPrefix,
		slipstreamprobe.Schema,
	)
	if err != nil {
		return nil, err
	}

	stages = append(
		stages,
		s.newStage(tunCfg.Workers, prb, writer, hooks...),
	)

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

	if tunCfg.CheckDNSResolver {
		resolverCfg := s.config.DNS.Resolver
		resolverCfg.Port = vaydnsCfg.ResolverPort
		resolverCfg.Transport = string(vaydnsCfg.ResolverType)

		stage, err := s.buildResolverStage(
			ctx,
			resolverCfg,
		)
		if err != nil {
			return nil, err
		}

		stages = append(stages, stage)
	}

	prb, err := vaydnsprobe.NewVayDNSProbe(
		vaydnsCfg,
		tunCfg.Timeout.Duration(),
		vaydnsprobe.WithVayDNSService(s.vaydnsService),
	)
	if err != nil {
		return nil, fmt.Errorf("create VayDNS probe: %w", err)
	}

	writer, err := s.newWriter(
		ctx,
		tunCfg.OutputPrefix,
		vaydnsprobe.Schema,
	)
	if err != nil {
		return nil, err
	}

	stages = append(
		stages,
		s.newStage(tunCfg.Workers, prb, writer, hooks...),
	)

	return stages, nil
}

func (s *scanner) newWriter(ctx context.Context, prefix string, schema result.ResultSchema) (result.Writer, error) {
	writer, err := s.writerFactory(ctx, result.WriterOptions{
		ResultPrefix: prefix,
		Schema:       schema,
		Config:       s.config.Writer,
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
