package scanner

import (
	"context"
	"fmt"
	"time"

	"bgscan/internal/core/config"
	"bgscan/internal/core/dns"
	"bgscan/internal/core/result"
	"bgscan/internal/core/scanner/engine"
	"bgscan/internal/core/scanner/portmgr"
	"bgscan/internal/core/scanner/probe"
)

//
// ────────────────────────────────────────────────────────────────
// Scan Modes
// ────────────────────────────────────────────────────────────────
//

type ScanMode string

const (
	ICMP_SCAN       ScanMode = "ICMP"
	TCP_SCAN        ScanMode = "TCP"
	HTTP_SCAN       ScanMode = "HTTP"
	XRAY_SCAN       ScanMode = "Xray"
	RESOLVE_SCAN    ScanMode = "Resolve"
	DNSTT_SCAN      ScanMode = "DNSTT"
	SLIPSTREAM_SCAN ScanMode = "Slipstream"
)

//
// ────────────────────────────────────────────────────────────────
// StageConfig
// ────────────────────────────────────────────────────────────────
//

type StageConfig struct {
	Mode    ScanMode
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

func (s *Scanner) Run() {
	if s.closed {
		panic("scanner already closed")
	}

	if s.ctx.Err() != nil {
		panic("scanner context is canceled")
	}

	if len(s.stages) == 0 {
		panic("no stages added to Scanner")
	}

	defer s.pause.Stop()
	defer s.pm.Close()

	if len(s.stages) == 1 {
		stg := s.stages[0]
		s.runSingle(stg, stg.Hooks)
		return
	}

	s.runChain(s.stages)
}

func (s *Scanner) runSingle(stage StageConfig, hooks engine.ScanHooks) {
	engine.RunScan(
		s.ctx,
		s.input,
		config.GetGeneral().MaxIPsToTest,
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
	engine.RunScanWithChain(s.ctx, s.input, config.GetGeneral().MaxIPsToTest, &engine.ChainConfig{
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
	file, err := result.BuildResultFilePath(result.ICMPResultDir, cfg.PrefixOutput)
	if err != nil {
		return StageConfig{}, err
	}
	writer, err := result.NewWriter(file, writerConfig(), ctx)
	if err != nil {
		return StageConfig{}, err
	}
	prb, err := probe.NewICMPProbe(cfg.Timeout.Duration(), cfg.Tries)
	if err != nil {
		return StageConfig{}, err
	}

	return StageConfig{
		Mode:    ICMP_SCAN,
		Workers: cfg.Workers,
		Probe:   prb,
		Writer:  writer,
		Rate:    calcRate(cfg.Workers, 25*time.Millisecond),
	}, nil
}

func (s *Scanner) BuildTCPStage(ctx context.Context) (StageConfig, error) {
	cfg := config.GetTCP()
	file, err := result.BuildResultFilePath(result.TCPResultDir, cfg.PrefixOutput)
	if err != nil {
		return StageConfig{}, err
	}
	writer, err := result.NewWriter(file, writerConfig(), ctx)
	if err != nil {
		return StageConfig{}, err
	}
	prb := probe.NewTCPProbe(fmt.Sprint(cfg.Port), cfg.Timeout.Duration(), cfg.Tries)

	return StageConfig{
		Mode:    TCP_SCAN,
		Workers: cfg.Workers,
		Probe:   prb,
		Writer:  writer,
		Rate:    calcRate(cfg.Workers, 50*time.Millisecond),
	}, nil
}

func (s *Scanner) BuildHTTPStage(ctx context.Context) (StageConfig, error) {
	cfg := config.GetHTTP()
	file, err := result.BuildResultFilePath(result.HTTPResultDir, cfg.PrefixOutput)
	if err != nil {
		return StageConfig{}, err
	}
	writer, err := result.NewWriter(file, writerConfig(), ctx)
	if err != nil {
		return StageConfig{}, err
	}
	reqCfg, err := probe.NewHTTPRequestFromConfig(*cfg)
	if err != nil {
		return StageConfig{}, err
	}
	prb := probe.NewHTTPProbe(*reqCfg)

	return StageConfig{
		Mode:    HTTP_SCAN,
		Workers: cfg.Workers,
		Probe:   prb,
		Writer:  writer,
		Rate:    calcRate(cfg.Workers, 80*time.Millisecond),
	}, nil
}

func (s *Scanner) BuildXrayStage(ctx context.Context, template string) (StageConfig, error) {
	cfg := config.GetXray()
	file, err := result.BuildResultFilePath(result.XRAYResultDir, cfg.PrefixOutput)
	if err != nil {
		return StageConfig{}, err
	}
	writer, err := result.NewWriter(file, writerConfig(), ctx)
	if err != nil {
		return StageConfig{}, err
	}
	prb, err := probe.NewXrayProbe(cfg, template, s.pm)
	if err != nil {
		return StageConfig{}, err
	}

	return StageConfig{
		Mode:    XRAY_SCAN,
		Workers: cfg.Workers,
		Probe:   prb,
		Writer:  writer,
		Rate:    calcRate(cfg.Workers, 200*time.Millisecond),
	}, nil
}

func (s *Scanner) BuildResolveStage(ctx context.Context) (StageConfig, error) {
	cfg := config.GetDNS().Resolver
	file, err := result.BuildResultFilePath(result.ResolveResultDir, cfg.PrefixOutput)
	if err != nil {
		return StageConfig{}, err
	}
	writer, err := result.NewWriter(file, writerConfig(), ctx)
	if err != nil {
		return StageConfig{}, err
	}

	rcodes := make([]uint16, 0, len(cfg.AcceptedRCodes))
	for _, r := range cfg.AcceptedRCodes {
		rcodes = append(rcodes, uint16(dns.ParseDNSRcode(r)))
	}

	prb := probe.NewResolverProbe(&probe.DnsRequest{
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
		Mode:    RESOLVE_SCAN,
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
	file, err := result.BuildResultFilePath(result.DNSTTResultDir, cfg.PrefixOutput)
	if err != nil {
		return StageConfig{}, err
	}
	writer, err := result.NewWriter(file, writerConfig(), ctx)
	if err != nil {
		return StageConfig{}, err
	}

	prb, err := probe.NewDNSTTProbe(probe.DNSTTConfig{
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
		Mode:    DNSTT_SCAN,
		Workers: cfg.Workers,
		Probe:   prb,
		Writer:  writer,
		Rate:    calcRate(cfg.Workers, time.Second),
	}, nil
}

func (s *Scanner) BuildSlipStreamStage(ctx context.Context) (StageConfig, error) {
	cfg := config.GetDNS().SlipStream
	port := config.GetDNS().Resolver.Port
	file, err := result.BuildResultFilePath(result.SlipStreamResultDir, cfg.PrefixOutput)
	if err != nil {
		return StageConfig{}, err
	}
	writer, err := result.NewWriter(file, writerConfig(), ctx)
	if err != nil {
		return StageConfig{}, err
	}

	prb, err := probe.NewSlipstreamProbe(cfg.Workers, probe.SlipstreamConfig{
		Domain:   cfg.Domain,
		CertPath: cfg.CertPath,
		DNSPort:  port,
		Timeout:  cfg.Timeout.Duration(),
	}, s.pm)
	if err != nil {
		return StageConfig{}, err
	}

	return StageConfig{
		Mode:    SLIPSTREAM_SCAN,
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
