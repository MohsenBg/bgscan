// Package scantype presents scan type selection and builds the requested scanner.
package scantype

import (
	"context"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"bgscan/internal/core/scanner"
	"bgscan/internal/core/xray"
	"bgscan/internal/ui/components/basic/menu"
	"bgscan/internal/ui/components/basic/notice"
	scannerUi "bgscan/internal/ui/components/scanner"
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
	DNSResolveScan
	XRAYScan
)

// Model is the scan type selection screen.
type Model struct {
	id           ui.ComponentID
	name         string
	state        *ui.AppState
	input        string
	xrayTemplate string
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
		menu.NewMenuItem("▦", "ICMP Scan", "i", m.open(ICMPScan)),
		menu.NewMenuItem("≡", "TCP Scan", "t", m.open(TCPScan)),
		menu.NewMenuItem("▦", "HTTP Scan", "h", m.open(HTTPScan)),
		menu.NewMenuItem("#", "DNS Scan", "d", m.open(DNSResolveScan)),
		menu.NewMenuItem("▦", "Xray Scan", "x", m.openXrayTemplates()),
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
		scn, err := m.createScanner(mode, m.input)
		if err != nil {
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

func (m *Model) createScanner(mode ScanType, input string) (scanner.Scanner, error) {
	ctx := context.Background()
	scn, err := scanner.NewScanner(ctx, input)
	if err != nil {
		return nil, err
	}

	if mode == XRAYScan {
		return m.buildXrayScanner(ctx, scn)
	}

	if mode == DNSResolveScan {
		return m.buildResolveScanner(ctx, scn)
	}

	stage, err := m.buildStage(ctx, scn, mode)
	if err != nil {
		return nil, err
	}

	scn.AddStage(stage)
	return scn, nil
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
		return scn.BuildTCPStage(ctx)
	}
}

func (m *Model) buildResolveScanner(ctx context.Context, scn scanner.Scanner) (scanner.Scanner, error) {
	if stage, err := scn.BuildResolveStage(ctx); err != nil {
		return nil, err
	} else {
		scn.AddStage(stage)
	}

	cfg := m.state.Config.DNS
	if cfg.DNSTT.Enabled {
		if stage, err := scn.BuildDNSTTStage(ctx); err == nil {
			scn.AddStage(stage)
		} else {
			return nil, err
		}
	}

	if cfg.SlipStream.Enabled {
		if stage, err := scn.BuildSlipStreamStage(ctx); err == nil {
			scn.AddStage(stage)
		} else {
			return nil, err
		}
	}

	return scn, nil
}

func (m *Model) buildXrayScanner(ctx context.Context, scn scanner.Scanner) (scanner.Scanner, error) {
	cfg := m.state.Config.Xray

	pre := map[string]func() error{
		"tcp": func() error {
			s, err := scn.BuildTCPStage(ctx)
			if err != nil {
				return err
			}
			scn.AddStage(s)
			return nil
		},
		"icmp": func() error {
			s, err := scn.BuildICMPStage(ctx)
			if err != nil {
				return err
			}
			scn.AddStage(s)
			return nil
		},
		"http": func() error {
			s, err := scn.BuildHTTPStage(ctx)
			if err != nil {
				return err
			}
			scn.AddStage(s)
			return nil
		},
	}

	scanType := strings.ToLower(cfg.PreScanType)
	if fn, ok := pre[scanType]; ok {
		if err := fn(); err != nil {
			return nil, fmt.Errorf("pre-scan failed: %w", err)
		}
	}

	xrayStage, err := scn.BuildXrayStage(ctx, m.xrayTemplate)
	if err != nil {
		return nil, fmt.Errorf("xray stage failed: %w", err)
	}

	scn.AddStage(xrayStage)
	return scn, nil
}

func (m *Model) errorCmd(title, message string) tea.Cmd {
	return notice.NewNoticeCmd(m.state.Layout, title, message, notice.NOTICE_ERROR)
}
