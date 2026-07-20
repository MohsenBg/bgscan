package scanner

import (
	"context"
	"errors"
	"fmt"
	"time"

	"bgscan/internal/core/config"
	"bgscan/internal/core/dns"
	"bgscan/internal/core/result"
	"bgscan/internal/core/scanner/engine"
	"bgscan/internal/core/scanner/portmgr"
	"bgscan/internal/core/scanner/probe"
	"bgscan/internal/core/scanner/probe/dnsttprobe"
	"bgscan/internal/core/scanner/probe/httpprobe"
	"bgscan/internal/core/scanner/probe/icmpprobe"
	"bgscan/internal/core/scanner/probe/resolveprobe"
	"bgscan/internal/core/scanner/probe/slipstreamprobe"
	"bgscan/internal/core/scanner/probe/tcpprobe"
	"bgscan/internal/core/scanner/probe/xrayprobe"
)

const baseResultDir = "result"

//
// ────────────────────────────────────────────────────────────────
// StageConfig
// ────────────────────────────────────────────────────────────────
//

type StageConfig struct {
	Workers int
	Probe   probe.Probe
	Writer  *result.Writer
	Rate    int
	Hooks   engine.ScanHooks
}

func (s *StageConfig) AddHooks(h engine.ScanHooks) *StageConfig {
	s.Hooks = h
	return s
}

//
// ────────────────────────────────────────────────────────────────
// Scanner
// ────────────────────────────────────────────────────────────────
//

type Scanner struct {
	ctx    context.Context
	cancel context.CancelFunc
	closed bool

	pause  *engine.PauseController
	input  string
	pm     *portmgr.PortManager
	stages []StageConfig
}

func NewScanner(ctx context.Context, input string) *Scanner {
	scanCtx, cancel := context.WithCancel(ctx)

	var poolSize uint16 = 3000
	pm, _ := portmgr.NewPortManager(portmgr.RandomBasePort(poolSize), poolSize)

	return &Scanner{
		ctx:    scanCtx,
		cancel: cancel,
		pause:  engine.NewPauseController(),
		pm:     pm,
		input:  input,
		stages: make([]StageConfig, 0),
	}
}

func (s *Scanner) GetStages() []StageConfig {
	return s.stages
}

func (s *Scanner) AddStage(stage StageConfig) {
	s.stages = append(s.stages, stage)
}

//
// ────────────────────────────────────────────────────────────────
// Run / Lifecycle
// ────────────────────────────────────────────────────────────────
//

func (s *Scanner) Run() error {
	if s.closed {
		return errors.New("scanner already closed")
	}

	if s.ctx.Err() != nil {
		return errors.New("scanner context is canceled")
	}

	if len(s.stages) == 0 {
		return errors.New("no stages added to Scanner")
	}

	defer s.pause.Stop()
	defer s.pm.Close()

	if len(s.stages) == 1 {
		stg := s.stages[0]
		s.runSingle(stg, stg.Hooks)
		return nil
	}

	s.runChain(s.stages)
	return nil
}

func (s *Scanner) runSingle(stage StageConfig, hooks engine.ScanHooks) {
	maxIP := max(config.GetGeneral().MaxIPsToTest, 0)

	engine.RunScan(
		s.ctx,
		s.input,
		uint64(maxIP),
		engine.ScanConfig{
			Workers: stage.Workers,
			Probe:   stage.Probe,
			Writer:  stage.Writer,
			Rate:    stage.Rate,
			Hooks:   hooks,
		},
		config.GetGeneral().Shuffled,
		s.pause,
	)
}

func (s *Scanner) runChain(stages []StageConfig) {
	maxIP := max(config.GetGeneral().MaxIPsToTest, 0)

	engineStages := make([]engine.ScanConfig, len(stages))
	for i, stage := range stages {
		engineStages[i] = engine.ScanConfig{
			Workers: stage.Workers,
			Probe:   stage.Probe,
			Writer:  stage.Writer,
			Rate:    stage.Rate,
			Hooks:   stage.Hooks,
		}
	}

	engine.RunScanWithChain(s.ctx, s.input, uint64(maxIP), &engine.ChainConfig{
		Mode:      engine.ParsePipelineMode(config.GetGeneral().PipelineMode),
		Stages:    engineStages,
		Pause:     s.pause,
		Shuffled:  config.GetGeneral().Shuffled,
		MaxBuffer: config.GetGeneral().MaxIPsPerStage,
	})
}

func (s *Scanner) Pause()                        { s.pause.Pause() }
func (s *Scanner) Resume()                       { s.pause.Resume() }
func (s *Scanner) IsPaused() bool                { return s.pause.IsPaused() }
func (s *Scanner) PausedDuration() time.Duration { return s.pause.PausedDuration() }

func (s *Scanner) Close() {
	if s.closed {
		return
	}
	s.closed = true

	if s.cancel != nil {
		s.cancel()
	}
}

//
// ────────────────────────────────────────────────────────────────
// Stage Builders
// ────────────────────────────────────────────────────────────────
//

func (s *Scanner) BuildICMPStage(ctx context.Context) (StageConfig, error) {
	cfg := config.GetICMP()

	file, err := result.BuildResultFilePath(baseResultDir, icmpprobe.Schema, cfg.PrefixOutput)
	if err != nil {
		return StageConfig{}, err
	}

	writer, err := result.NewWriter(file, icmpprobe.Schema, writerConfig(), ctx)
	if err != nil {
		return StageConfig{}, err
	}

	prb, err := icmpprobe.NewICMPProbe(cfg.Timeout.Duration(), cfg.Tries)
	if err != nil {
		return StageConfig{}, err
	}

	return StageConfig{
		Workers: cfg.Workers,
		Probe:   prb,
		Writer:  writer,
		Rate:    calcRate(cfg.Workers, 25*time.Millisecond),
	}, nil
}

func (s *Scanner) BuildTCPStage(ctx context.Context) (StageConfig, error) {
	cfg := config.GetTCP()

	file, err := result.BuildResultFilePath(baseResultDir, tcpprobe.Schema, cfg.PrefixOutput)
	if err != nil {
		return StageConfig{}, err
	}

	writer, err := result.NewWriter(file, tcpprobe.Schema, writerConfig(), ctx)
	if err != nil {
		return StageConfig{}, err
	}

	prb := tcpprobe.NewTCPProbe(fmt.Sprint(cfg.Port), cfg.Timeout.Duration(), cfg.Tries)

	return StageConfig{
		Workers: cfg.Workers,
		Probe:   prb,
		Writer:  writer,
		Rate:    calcRate(cfg.Workers, 50*time.Millisecond),
	}, nil
}

func (s *Scanner) BuildHTTPStage(ctx context.Context) (StageConfig, error) {
	cfg := config.GetHTTP()

	file, err := result.BuildResultFilePath(baseResultDir, httpprobe.Schema, cfg.PrefixOutput)
	if err != nil {
		return StageConfig{}, err
	}

	writer, err := result.NewWriter(file, httpprobe.Schema, writerConfig(), ctx)
	if err != nil {
		return StageConfig{}, err
	}

	if isHTTP3(cfg.Version) {
		reqCfg, err := httpprobe.NewHTTPRequestFromConfig(*cfg)
		if err != nil {
			return StageConfig{}, err
		}
		prb, err := httpprobe.NewHTTP3Probe(*reqCfg, cfg.AcceptedStatusCodes)
		if err != nil {
			return StageConfig{}, err
		}
		return StageConfig{
			Workers: cfg.Workers,
			Probe:   prb,
			Writer:  writer,
			Rate:    calcRate(cfg.Workers, 80*time.Millisecond),
		}, nil
	}

	reqCfg, err := httpprobe.NewHTTPRequestFromConfig(*cfg)
	if err != nil {
		return StageConfig{}, err
	}
	prb := httpprobe.NewHTTPProbe(*reqCfg, cfg.AcceptedStatusCodes)

	return StageConfig{
		Workers: cfg.Workers,
		Probe:   prb,
		Writer:  writer,
		Rate:    calcRate(cfg.Workers, 80*time.Millisecond),
	}, nil
}

func isHTTP3(version string) bool {
	return version == "h3" || version == "http3"
}

func (s *Scanner) BuildXrayStage(ctx context.Context, template string) (StageConfig, error) {
	cfg := config.GetXray()

	file, err := result.BuildResultFilePath(baseResultDir, xrayprobe.Schema, cfg.PrefixOutput)
	if err != nil {
		return StageConfig{}, err
	}

	writer, err := result.NewWriter(file, xrayprobe.Schema, writerConfig(), ctx)
	if err != nil {
		return StageConfig{}, err
	}

	prb, err := xrayprobe.NewXrayProbe(cfg, template, s.pm)
	if err != nil {
		return StageConfig{}, err
	}

	return StageConfig{
		Workers: cfg.Workers,
		Probe:   prb,
		Writer:  writer,
		Rate:    calcRate(cfg.Workers, 200*time.Millisecond),
	}, nil
}

func (s *Scanner) BuildResolveStage(ctx context.Context) (StageConfig, error) {
	cfg := config.GetDNS().Resolver

	file, err := result.BuildResultFilePath(baseResultDir, resolveprobe.Schema, cfg.PrefixOutput)
	if err != nil {
		return StageConfig{}, err
	}

	writer, err := result.NewWriter(file, resolveprobe.Schema, writerConfig(), ctx)
	if err != nil {
		return StageConfig{}, err
	}

	rcodes := make([]uint16, 0, len(cfg.AcceptedRCodes))
	for _, r := range cfg.AcceptedRCodes {
		rcodes = append(rcodes, uint16(dns.ParseDNSRcode(r)))
	}

	prb := resolveprobe.NewResolverProbe(&resolveprobe.DNSRequest{
		Domain:          cfg.Domain,
		Port:            cfg.Port,
		RandomSubdomain: cfg.RandomSubdomain,
		DpiCheck:        cfg.CheckDPI,
		DpiTimeout:      cfg.DPITimeout.Duration(),
		DpiTries:        cfg.DPITries,
		Edns0Size:       cfg.EDNSBufSize,
		CheckTypes:      cfg.CheckTypes,
		AcceptedRcodes:  rcodes,
		Timeout:         cfg.Timeout.Duration(),
		Transport:       dns.ParseTransport(cfg.Protocol),
		Tries:           cfg.Tries,
	})

	return StageConfig{
		Workers: cfg.Workers,
		Probe:   prb,
		Writer:  writer,
		Rate:    calcRate(cfg.Workers, 500*time.Millisecond),
	}, nil
}

func (s *Scanner) BuildDNSTTStage(ctx context.Context) (StageConfig, error) {
	cfg := config.GetDNS().DNSTT
	transport := config.GetDNS().Resolver.Protocol
	port := config.GetDNS().Resolver.Port

	file, err := result.BuildResultFilePath(baseResultDir, dnsttprobe.Schema, cfg.PrefixOutput)
	if err != nil {
		return StageConfig{}, err
	}

	writer, err := result.NewWriter(file, dnsttprobe.Schema, writerConfig(), ctx)
	if err != nil {
		return StageConfig{}, err
	}

	prb, err := dnsttprobe.NewDNSTTProbe(dnsttprobe.DNSTTConfig{
		Domain:    cfg.Domain,
		PubKey:    cfg.PublicKey,
		Transport: dns.ParseTransport(transport),
		DNSPort:   port,
		Timeout:   cfg.Timeout.Duration(),
	}, s.pm)
	if err != nil {
		return StageConfig{}, err
	}

	return StageConfig{
		Workers: cfg.Workers,
		Probe:   prb,
		Writer:  writer,
		Rate:    calcRate(cfg.Workers, time.Second),
	}, nil
}

func (s *Scanner) BuildSlipStreamStage(ctx context.Context) (StageConfig, error) {
	cfg := config.GetDNS().SlipStream
	port := config.GetDNS().Resolver.Port

	file, err := result.BuildResultFilePath(baseResultDir, slipstreamprobe.Schema, cfg.PrefixOutput)
	if err != nil {
		return StageConfig{}, err
	}

	writer, err := result.NewWriter(file, slipstreamprobe.Schema, writerConfig(), ctx)
	if err != nil {
		return StageConfig{}, err
	}

	prb, err := slipstreamprobe.NewSlipstreamProbe(cfg.Workers, slipstreamprobe.SlipstreamConfig{
		Domain:   cfg.Domain,
		CertPath: cfg.CertPath,
		DNSPort:  port,
		Timeout:  cfg.Timeout.Duration(),
	}, s.pm)
	if err != nil {
		return StageConfig{}, err
	}

	return StageConfig{
		Workers: cfg.Workers,
		Probe:   prb,
		Writer:  writer,
		Rate:    calcRate(cfg.Workers, time.Second),
	}, nil
}

//
// ────────────────────────────────────────────────────────────────
// Helpers
// ────────────────────────────────────────────────────────────────
//

func calcRate(workers int, minProbeTime time.Duration) int {
	return int(time.Second/minProbeTime) * workers
}

func writerConfig() result.Config {
	cfg := config.GetWriter()
	return result.Config{
		MergeFlushInterval: cfg.MergeFlushInterval.Duration(),
		ChanSize:           cfg.ChanSize,
		BatchSize:          cfg.BatchSize,
	}
}
