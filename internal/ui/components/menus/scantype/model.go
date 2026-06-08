package scantype

import (
	"bgscan/internal/core/config"
	"bgscan/internal/core/scanner"
	"bgscan/internal/core/xray"
	"bgscan/internal/ui/components/basic/menu"
	"bgscan/internal/ui/components/basic/notice"
	scannerUi "bgscan/internal/ui/components/scanner"
	"bgscan/internal/ui/components/tables/outbounds"
	"bgscan/internal/ui/shared/env"
	"bgscan/internal/ui/shared/layout"
	"bgscan/internal/ui/shared/ui"
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	id           ui.ComponentID
	name         string
	layout       *layout.Layout
	input        string
	xrayTemplate string
	menu         ui.Component
	closeScanner bool
	scanner      *scanner.Scanner
}

func (m *Model) Init() tea.Cmd      { return nil }
func (m *Model) ID() ui.ComponentID { return m.id }
func (m *Model) Name() string       { return m.name }
func (m *Model) OnClose() tea.Cmd {
	if m.closeScanner && m.scanner != nil {
		m.scanner.Close()
	}
	return nil
}

func New(layout *layout.Layout, input string) *Model {
	m := &Model{
		id:           ui.NewComponentID(),
		name:         "Scan Menu",
		layout:       layout,
		input:        input,
		closeScanner: true,
	}

	items := []menu.MenuItem{
		menu.NewMenuItem("▦", "ICMP Scan", "i", m.openScanner(scanner.ICMP_SCAN)),
		menu.NewMenuItem("≡", "TCP Scan", "t", m.openScanner(scanner.TCP_SCAN)),
		menu.NewMenuItem("▦", "HTTP Scan", "h", m.openScanner(scanner.HTTP_SCAN)),
		menu.NewMenuItem("#", "DNS Scan", "d", m.openScanner(scanner.RESOLVE_SCAN)),
		menu.NewMenuItem("▦", "Xray Scan", "x", m.openXrayTemplates()),
	}

	m.menu = menu.New(items, "Select Scan Type", layout)
	return m
}

func (m *Model) Mode() env.Mode {
	return env.NormalMode
}

func (m *Model) openXrayTemplates() tea.Cmd {
	return ui.OpenComponentCmd(
		outbounds.New(m.layout, "select outbound", func(xof *xray.XrayOutboundsFile) tea.Cmd {
			m.xrayTemplate = xof.Name
			return m.openScanner(scanner.XRAY_SCAN)
		}))
}

// openScanner closes overlay and opens the selected scanner UI.
func (m *Model) openScanner(mode scanner.ScanMode) tea.Cmd {
	scn, err := m.createScanner(mode, m.input)
	if err != nil {
		return m.errorCmd("Error Creating Scanner", err.Error())
	}

	m.closeScanner = false
	return ui.OpenComponentCmd(
		scannerUi.New(m.layout, 10_000, scn),
	)
}

// createScanner builds a fully configured scanner with all appropriate stages.
func (m *Model) createScanner(mode scanner.ScanMode, input string) (*scanner.Scanner, error) {
	ctx := context.Background()
	scn := scanner.NewScanner(ctx, input)

	addShuffle := func(enabled bool) {
		if enabled {
			scn.AddPreprocessor(&scanner.ShufflePreprocessor{})
		}
	}

	var (
		stage scanner.StageConfig
		err   error
	)

	switch mode {

	case scanner.TCP_SCAN:
		stage, err = scn.BuildTCPStage(ctx)
		addShuffle(config.GetTCP().ShuffleIPs)

	case scanner.ICMP_SCAN:
		stage, err = scn.BuildICMPStage(ctx)
		addShuffle(config.GetICMP().ShuffleIPs)

	case scanner.HTTP_SCAN:
		stage, err = scn.BuildHTTPStage(ctx)
		addShuffle(config.GetHTTP().ShuffleIPs)

	case scanner.XRAY_SCAN:
		return m.buildXrayScanner(ctx, scn)

	case scanner.RESOLVE_SCAN:
		return m.buildResolveScanner(ctx, scn)

	default:
		stage, err = scn.BuildTCPStage(ctx)
		addShuffle(config.GetTCP().ShuffleIPs)
	}

	if err != nil {
		return nil, err
	}

	scn.AddStage(stage)
	return scn, nil
}

func (m *Model) buildResolveScanner(ctx context.Context, scn *scanner.Scanner) (*scanner.Scanner, error) {

	stage, err := scn.BuildResolveStage(ctx)
	if err != nil {
		return nil, err
	}

	scn.AddStage(stage)

	if config.GetDNS().Resolver.ShuffleIPs {
		scn.AddPreprocessor(&scanner.ShufflePreprocessor{})
	}

	if config.GetDNS().DNSTT.Enabled {
		ttStage, err := scn.BuildDNSTTStage(ctx)
		if err != nil {
			return nil, err
		}
		scn.AddStage(ttStage)
	}

	if config.GetDNS().SlipStream.Enabled {
		ssStage, err := scn.BuildSlipStreamStage(ctx)
		if err != nil {
			return nil, err
		}
		scn.AddStage(ssStage)
	}

	return scn, nil
}

func (m *Model) buildXrayScanner(ctx context.Context, scn *scanner.Scanner) (*scanner.Scanner, error) {
	xrayCfg := config.GetXray()

	// Add shuffle if enabled
	if xrayCfg.ShuffleIPs {
		scn.AddPreprocessor(&scanner.ShufflePreprocessor{})
	}

	// Add pre-scan stage based on configuration
	switch xrayCfg.PreScanType {
	case "tcp":
		tcpStage, err := scn.BuildTCPStage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to build TCP pre-scan: %w", err)
		}
		scn.AddStage(tcpStage)

	case "icmp":
		icmpStage, err := scn.BuildICMPStage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to build ICMP pre-scan: %w", err)
		}
		scn.AddStage(icmpStage)

	case "http":
		httpStage, err := scn.BuildHTTPStage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to build HTTP pre-scan: %w", err)
		}
		scn.AddStage(httpStage)

	case "none", "":
		// No pre-scan, skip
	}

	// Add main Xray stage
	xrayStage, err := scn.BuildXrayStage(ctx, m.xrayTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to build Xray stage: %w", err)
	}
	scn.AddStage(xrayStage)

	return scn, nil
}

// errorCmd returns a styled error notice.
func (m *Model) errorCmd(title, message string) tea.Cmd {
	return notice.NewNoticeCmd(m.layout, title, message, notice.NOTICE_ERROR)
}
