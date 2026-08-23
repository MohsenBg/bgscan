// Package scantype presents scan type selection and builds the requested scanner.
package scantype

import (
	"context"
	"fmt"

	tea "charm.land/bubbletea/v2"

	"bgscan/internal/core/dns"
	"bgscan/internal/core/scanner"
	"bgscan/internal/core/xray"
	"bgscan/internal/logger"
	"bgscan/internal/ui/components/basic/menu"
	"bgscan/internal/ui/components/basic/notice"
	scannerUi "bgscan/internal/ui/components/scanner"
	"bgscan/internal/ui/components/tables/dnstun"
	"bgscan/internal/ui/components/tables/outbounds"
	"bgscan/internal/ui/shared/env"
	"bgscan/internal/ui/shared/ui"
)

// ScanType selects the scan mode.
type ScanType uint8

const (
	ICMPScan ScanType = iota
	TCPScan
	HTTPScan
	XRAYScan
	DNSResolveScan
	DNSTun
)

func (s ScanType) String() string {
	switch s {
	case ICMPScan:
		return "ICMP"
	case TCPScan:
		return "TCP"
	case HTTPScan:
		return "HTTP"
	case XRAYScan:
		return "Xray"
	case DNSResolveScan:
		return "DNS resolve"
	case DNSTun:
		return "DNS tunneling"
	default:
		return fmt.Sprintf("ScanType(%d)", uint8(s))
	}
}

// Model is the scan type selection screen.
type Model struct {
	id           ui.ComponentID
	name         string
	state        *ui.AppState
	input        string
	xrayTemplate string
	dnsTunCfg    *dns.DNSTunConfigFile
	menu         ui.Component
	closeScanner bool
	scanner      scanner.Scanner
}

// New creates the scan type selector.
func New(state *ui.AppState, input string) *Model {
	m := &Model{
		id:           ui.NewComponentID(),
		name:         "Scan Menu",
		state:        state,
		input:        input,
		closeScanner: true,
	}

	m.menu = menu.New([]menu.MenuItem{
		menu.NewMenuItem("◉", "ICMP Scan", "i", m.open(ICMPScan)),
		menu.NewMenuItem("→", "TCP Scan", "t", m.open(TCPScan)),
		menu.NewMenuItem("◌", "HTTP Scan", "h", m.open(HTTPScan)),
		menu.NewMenuItem("◇", "Xray Scan", "x", m.openXrayTemplates()),
		menu.NewMenuItem("?", "DNS Resolve", "r", m.open(DNSResolveScan)),
		menu.NewMenuItem("≈", "DNS Tunneling", "d", m.openDNSTunConfig()),
	}, "Select Scan Type", state.Layout)
	return m
}

func (m *Model) Init() tea.Cmd      { return nil }
func (m *Model) ID() ui.ComponentID { return m.id }
func (m *Model) Name() string       { return m.name }
func (m *Model) Mode() env.Mode     { return env.NormalMode }

// OnClose closes the active scanner if this menu opened it.
func (m *Model) OnClose() tea.Cmd {
	if m.closeScanner && m.scanner != nil {
		err := m.scanner.Close()
		if err != nil {
			return notice.NewNoticeCmd(m.state.Layout, "Failed to close scanner", err.Error(), notice.NOTICE_ERROR)
		}
	}
	return nil
}

func (m *Model) open(mode ScanType) tea.Cmd {
	return func() tea.Msg {
		logger.UIInfo("Starting %s scan", mode)

		scn, err := m.createScanner(mode, m.input)
		if err != nil {
			logger.UIError("Failed to create %s scanner: %v", mode, err)
			return m.errorCmd("scanner error", err.Error())
		}

		m.closeScanner = false
		return ui.OpenComponentCmd(scannerUi.New(m.state, 10_000, scn))()
	}
}

func (m *Model) openXrayTemplates() tea.Cmd {
	return ui.OpenComponentCmd(
		outbounds.New(m.state.Layout, "select outbound", func(xof *xray.XrayOutboundsFile) tea.Cmd {
			m.xrayTemplate = xof.Name
			return m.open(XRAYScan)
		}),
	)
}

func (m *Model) openDNSTunConfig() tea.Cmd {
	return ui.OpenComponentCmd(
		dnstun.New(m.state.Layout, "select dns tunneling config", func(dof *dns.DNSTunConfigFile) tea.Cmd {
			m.dnsTunCfg = dof
			return m.open(DNSTun)
		}),
	)
}

func (m *Model) createScanner(mode ScanType, input string) (scanner.Scanner, error) {
	ctx := context.Background()

	scn, err := scanner.NewScanner(ctx, input, scanner.WithConfig(*m.state.Config))
	if err != nil {
		return nil, err
	}

	switch mode {
	case XRAYScan:
		return m.buildXrayScanner(ctx, scn)

	case DNSResolveScan:
		return m.buildResolveScanner(ctx, scn)

	case DNSTun:
		return m.buildDNSTunScanner(ctx, scn)

	default:
		stage, err := m.buildStage(ctx, scn, mode)
		if err != nil {
			_ = scn.Close()
			return nil, err
		}

		scn.AddStage(stage)
		return scn, nil
	}
}

func (m *Model) buildStage(ctx context.Context, scn scanner.Scanner, mode ScanType) (scanner.StageConfig, error) {
	switch mode {
	case TCPScan:
		return scn.BuildTCPStage(ctx)
	case ICMPScan:
		return scn.BuildICMPStage(ctx)
	case HTTPScan:
		return scn.BuildHTTPStage(ctx)
	default:
		return scanner.StageConfig{}, fmt.Errorf("unsupported scan type: %d", mode)
	}
}

func (m *Model) buildResolveScanner(ctx context.Context, scn scanner.Scanner) (scanner.Scanner, error) {
	stage, err := scn.BuildResolveStage(ctx)
	if err != nil {
		_ = scn.Close()
		return nil, err
	}

	scn.AddStage(stage)
	return scn, nil
}

func (m *Model) buildXrayScanner(ctx context.Context, scn scanner.Scanner) (scanner.Scanner, error) {
	stages, err := scn.BuildXrayStage(ctx, m.xrayTemplate)
	if err != nil {
		_ = scn.Close()
		return nil, fmt.Errorf("xray stage failed: %w", err)
	}

	for _, stage := range stages {
		scn.AddStage(stage)
	}
	return scn, nil
}

func (m *Model) buildDNSTunScanner(ctx context.Context, scn scanner.Scanner) (scanner.Scanner, error) {
	var stages []scanner.StageConfig

	switch m.dnsTunCfg.Protocol {
	case dns.DNSTunProtocolDNSTT:
		stage, err := scn.BuildDNSTTStage(ctx, m.dnsTunCfg.Name)
		if err != nil {
			return nil, fmt.Errorf("build DNSTT stage: %w", err)
		}

		stages = append(stages, stage...)

	case dns.DNSTunProtocolVayDNS:
		stage, err := scn.BuildVayDNSStage(ctx, m.dnsTunCfg.Name)
		if err != nil {
			return nil, fmt.Errorf("build VayDNS stage: %w", err)
		}

		stages = append(stages, stage...)

	case dns.DNSTunProtocolSlipstream:
		stage, err := scn.BuildSlipStreamStage(ctx, m.dnsTunCfg.Name)
		if err != nil {
			return nil, fmt.Errorf("build Slipstream stage: %w", err)
		}

		stages = append(stages, stage...)

	default:
		return nil, fmt.Errorf(
			"unsupported DNS tunnel protocol: %q",
			m.dnsTunCfg.Protocol,
		)
	}

	for _, stage := range stages {
		scn.AddStage(stage)
	}

	return scn, nil
}

func (m *Model) errorCmd(title, message string) tea.Cmd {
	return notice.NewNoticeCmd(m.state.Layout, title, message, notice.NOTICE_ERROR)
}
