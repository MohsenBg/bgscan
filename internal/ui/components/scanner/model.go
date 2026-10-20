package scanner

import (
	"time"

	"github.com/MohsenBg/bgscan/internal/core/result"
	"github.com/MohsenBg/bgscan/internal/core/scanner"
	logview "github.com/MohsenBg/bgscan/internal/ui/components/basic/logview"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/tabs"
	"github.com/MohsenBg/bgscan/internal/ui/components/tables/ipviewer"
	"github.com/MohsenBg/bgscan/internal/ui/shared/dialog"
	"github.com/MohsenBg/bgscan/internal/ui/shared/keys"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

const (
	maxWidth      = 120
	minWidth      = 20
	minHeight     = 10
	pulseSteps    = 8
	pulseStrength = 0.65

	tabHeight      = 2 // 2 + 1 padding
	progressHeight = 2 // 2 + 1 padding
	panelsHeight   = 10
	paddingX       = 2
	paddingY       = -2
)

// Model is the scanner screen
type Model struct {
	id      ui.ComponentID
	name    string
	session *ui.Session
	tabs    ui.Component

	scanner scanner.Scanner
	stages  []*scanStage

	currentTab int

	pulse    int
	pulseDir int

	closing bool
	width   int
}

// New builds the scanner model and wires stage hooks for progress/result updates.
func New(session *ui.Session, scanner scanner.Scanner) *Model {
	stages := scanner.GetStages()
	n := len(stages)

	m := &Model{
		id:      ui.NewComponentID(),
		name:    "Scanner",
		session: session,
		scanner: scanner,
		stages:  make([]*scanStage, n),
	}

	tabsList := make([]tabs.Tab[int], n)
	for i, stage := range stages {
		s := newScanStage(session.Env(), session.Program, i, stage)
		tabsList[i] = tabs.NewTab(stage.Probe.Schema().Name, i)
		m.stages[i] = s
	}

	m.tabs = tabs.New(session.Env(), tabsList, func(idx int, _ tabs.Tab[int]) tea.Cmd {
		m.currentTab = idx
		return nil
	}, tabs.WithSlideAccent[int]())

	m.updateLayout()
	return m
}

func (m *Model) ID() ui.ComponentID { return m.id }
func (m *Model) Name() string       { return m.name }
func (m *Model) Mode() ui.KeyMode   { return ui.KeyModeManage }
func (m *Model) OnClose() tea.Cmd   { return nil }

// Init registers scanner hooks and starts the scan run plus periodic UI ticks.
func (m *Model) Init() tea.Cmd {
	for i, s := range m.stages {
		_ = m.scanner.UpdateStageHooks(i, s.hooks())
	}

	runCmd := func() tea.Msg {
		if err := m.scanner.Run(); err != nil {
			return scanErrorMsg{err: err}
		}
		return nil
	}
	return tea.Batch(tea.Cmd(runCmd), pulseCmd(), m.tabs.Init())
}

func (m *Model) SetLayoutSize(w, h int) {

}

// current returns the active stage for the current tab.
func (m *Model) current() *scanStage {
	return m.stages[m.currentTab]
}

// getStage returns the stage for idx, or nil when out of range.
func (m *Model) getStage(idx int) *scanStage {
	if idx < 0 || idx >= len(m.stages) {
		return nil
	}
	return m.stages[idx]
}

func (m *Model) updateLayout() {
	env := m.session.Env()
	body := env.Layout.Body()

	width := min(maxWidth, body.Width)
	width = max(minWidth, width-paddingX*2)

	tableH := body.Height - tabHeight - progressHeight - panelsHeight - paddingY
	tableH = max(minHeight, tableH)

	for _, s := range m.stages {
		s.resize(width, tableH)
	}

	if t, ok := m.tabs.(*tabs.Model[int]); ok {
		t.SetMaxWidth(width)
	}
	m.width = width
}

func (m *Model) togglePause() {
	if m.scanner.IsPaused() {
		m.scanner.Resume()
	} else {
		m.scanner.Pause()
	}
}

func (m *Model) closeScan() tea.Msg {
	m.closing = true
	return scanCloseResultMsg{err: m.scanner.Close()}
}

func (m *Model) openLogViewer() tea.Cmd {
	return func() tea.Msg {
		env := m.session.Env()
		coreLog := m.session.CoreLog()
		v := logview.New(env, &coreLog, "core logs")
		v.SetContainerWidth(min(70, env.Layout.Body().Width))
		v.SetShowBorder(false)
		return dialog.OpenDialog(v)
	}
}

func createIPViewer(env *ui.Env, schema result.ResultSchema) ui.Component {
	k := env.Keys
	viewer := ipviewer.New(env, "", nil, schema,
		ipviewer.WithMaxRow(env.Settings.MaxRow),
		ipviewer.WithFiltering(false),
		ipviewer.WithShortHints(
			keys.Short(k.Nav.NextTab, k.Scan.TogglePause, k.Scan.ToggleLogs)...,
		),
		ipviewer.WithLongHints(
			keys.Long(k.Nav.NextTab, k.Scan.TogglePause, k.Scan.ToggleLogs)...,
		),
	)
	return viewer
}

func pulseCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg {
		return pulseMsg{}
	})
}
