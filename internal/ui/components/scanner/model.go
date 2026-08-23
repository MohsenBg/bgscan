package scanner

import (
	"sort"
	"sync"
	"time"

	"bgscan/internal/core/result"
	"bgscan/internal/core/scanner"
	"bgscan/internal/core/scanner/engine"
	"bgscan/internal/logger"
	logview "bgscan/internal/ui/components/basic/logview"
	"bgscan/internal/ui/components/basic/notice"
	"bgscan/internal/ui/components/basic/progress"
	"bgscan/internal/ui/components/basic/table"
	"bgscan/internal/ui/components/basic/tabs"
	"bgscan/internal/ui/components/tables/ipviewer"
	"bgscan/internal/ui/shared/dialog"
	"bgscan/internal/ui/shared/env"
	"bgscan/internal/ui/shared/layout"
	"bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// StageStatus tracks the lifecycle of a single scan stage.
type StageStatus int

const (
	StatusWaiting StageStatus = iota
	StatusPreProcess
	StatusScanning
	StatusEnded
	StatusError
)

// Model is the scanner screen. It owns the stage lifecycle, progress UI,
// tabbed stage navigation, and aggregated results.
type Model struct {
	id    ui.ComponentID
	name  string
	state *ui.AppState
	tabs  ui.Component

	scn        scanner.Scanner
	stages     []scanner.StageConfig
	stageCount int
	maxIPs     int

	progress   []ui.Component
	ipViewers  []ui.Component
	currentTab int

	results [][]result.Result
	batch   [][]result.Result

	mu           sync.Mutex
	status       []StageStatus
	progressInfo []engine.Progress
	scanError    error
	errorShown   bool

	closing bool
}

// New builds the scanner model and wires stage hooks for progress/result updates.
func New(state *ui.AppState, maxIPs int, scn scanner.Scanner) *Model {
	stages := scn.GetStages()
	n := len(stages)

	m := &Model{
		id:           ui.NewComponentID(),
		name:         "Scanner",
		state:        state,
		scn:          scn,
		stages:       stages,
		stageCount:   n,
		maxIPs:       maxIPs,
		progress:     make([]ui.Component, n),
		ipViewers:    make([]ui.Component, n),
		results:      make([][]result.Result, n),
		batch:        make([][]result.Result, n),
		status:       make([]StageStatus, n),
		progressInfo: make([]engine.Progress, n),
	}

	tabsList := make([]tabs.Tab[int], n)
	for i, stage := range stages {
		m.ipViewers[i] = createIPViewer(m.state.Layout, stage.Probe.Schema())
		m.progress[i] = progress.New(m.state.Layout)
		m.results[i] = make([]result.Result, 0, maxIPs)
		m.batch[i] = make([]result.Result, 0, 128)
		m.status[i] = StatusWaiting
		tabsList[i] = tabs.NewTab(stage.Probe.Schema().Name, i)
	}

	m.tabs = tabs.New(m.state.Layout, tabsList, func(idx int, _ tabs.Tab[int]) tea.Cmd {
		m.currentTab = idx
		return m.immediateTick()
	})

	m.recalcTableSizes()
	return m
}

func (m *Model) ID() ui.ComponentID { return m.id }
func (m *Model) Name() string       { return m.name }
func (m *Model) Mode() env.Mode     { return env.ScanMode }
func (m *Model) OnClose() tea.Cmd   { return nil }

// Init registers scanner hooks and starts the scan run plus periodic UI ticks.
func (m *Model) Init() tea.Cmd {
	for i := range m.stages {
		_ = m.scn.UpdateStageHooks(i, engine.ScanHooks{
			OnError:    m.onError,
			OnProgress: m.onProgress(i),
			OnSuccess:  m.onSuccess(i),
			OnScanEnd:  m.onScanEnd(i),
		})
	}
	m.status[0] = StatusPreProcess

	runCmd := func() tea.Msg {
		if err := m.scn.Run(); err != nil {
			return scanErrorMsg{err: err}
		}
		return nil
	}
	return tea.Batch(tea.Cmd(runCmd), m.tick())
}

func (m *Model) SetWidth(w int) {
	for _, v := range m.ipViewers {
		if viewer, ok := v.(*ipviewer.Model); ok {
			viewer.SetWidth(w)
		}
	}
}

func (m *Model) SetHeight(h int) {
	for _, v := range m.ipViewers {
		if viewer, ok := v.(*ipviewer.Model); ok {
			viewer.SetHeight(h)
		}
	}
}

func (m *Model) SetTableSize(w, h int) {
	for _, v := range m.ipViewers {
		if viewer, ok := v.(*ipviewer.Model); ok {
			viewer.SetWidth(w)
			viewer.SetHeight(h)
		}
	}
}

func (m *Model) recalcTableSizes() {
	if m.closing {
		return
	}
	availableWidth := min(90, m.state.Layout.BodyContentWidth())
	availableHeight := m.state.Layout.BodyContentHeight()

	tabsHeight := lipgloss.Height(m.tabs.View())
	progressHeight := lipgloss.Height(m.renderProgress(m.currentTab))
	padding := 2

	tableWidth := availableWidth - padding
	tableHeight := availableHeight - tabsHeight - progressHeight - padding

	if tableWidth < 20 {
		tableWidth = 20
	}
	if tableHeight < 5 {
		tableHeight = 15
	}

	m.SetTableSize(tableWidth, tableHeight)

	if t, ok := m.tabs.(*tabs.Model[int]); ok {
		t.SetMaxWidth(tableWidth)
	}
}

func (m *Model) onSuccess(i int) func(result.Result) {
	return func(ip result.Result) {
		m.mu.Lock()
		m.batch[i] = append(m.batch[i], ip)
		m.mu.Unlock()
	}
}

func (m *Model) onProgress(i int) func(engine.Progress) {
	return func(p engine.Progress) {
		m.mu.Lock()
		defer m.mu.Unlock()
		if m.status[i] <= StatusPreProcess {
			m.status[i] = StatusScanning
		}
		m.progressInfo[i] = p
	}
}

func (m *Model) onError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := range m.status {
		if m.status[i] != StatusEnded {
			m.status[i] = StatusError
		}
	}
	m.scanError = err
	m.errorShown = false
}

func (m *Model) onScanEnd(i int) func() {
	return func() {
		m.mu.Lock()
		m.status[i] = StatusEnded
		m.mu.Unlock()
	}
}

func (m *Model) mergeBatch() {
	if m.closing {
		return
	}
	for i, stage := range m.stages {
		m.mu.Lock()
		if len(m.batch[i]) == 0 {
			m.mu.Unlock()
			continue
		}
		batch := m.batch[i]
		m.batch[i] = m.batch[i][:0]
		m.mu.Unlock()

		for idx, r := range batch {
			rec := r.ToRecord()
			norm, err := stage.Probe.Schema().Parser(rec)
			if err == nil {
				batch[idx] = norm
			}
		}

		m.results[i] = append(m.results[i], batch...)
		sort.SliceStable(m.results[i], func(a, b int) bool {
			sa, sb := m.results[i][a].Score(), m.results[i][b].Score()
			if sa != sb {
				return sa > sb
			}
			return m.results[i][a].Key() < m.results[i][b].Key()
		})
		if len(m.results[i]) > m.maxIPs {
			m.results[i] = m.results[i][:m.maxIPs]
		}
		if viewer, ok := m.ipViewers[i].(*ipviewer.Model); ok {
			viewer.SetRows(m.results[i])
		}
	}
}

func (m *Model) currentStatus() StageStatus {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.status[m.currentTab]
}

func (m *Model) currentProgress() float64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.progressInfo[m.currentTab].Percent / 100
}

func (m *Model) tick() tea.Cmd {
	return tea.Tick(m.state.Config.General.StatusInterval.Duration(), func(time.Time) tea.Msg { return tickMsg{} })
}

func (m *Model) immediateTick() tea.Cmd {
	return func() tea.Msg { return immediateTickMsg{} }
}

func (m *Model) forceResize() tea.Cmd {
	return func() tea.Msg {
		return tea.WindowSizeMsg{
			Width:  m.state.Layout.Terminal.Width,
			Height: m.state.Layout.Terminal.Height,
		}
	}
}

func (m *Model) togglePause() {
	if m.scn.IsPaused() {
		m.scn.Resume()
	} else {
		m.scn.Pause()
	}
}

func (m *Model) asyncClose() tea.Cmd {
	return func() tea.Msg {
		ch := make(chan error, 1)
		go func() { ch <- m.scn.Close() }()
		return scanClosedMsg{err: <-ch}
	}
}

func (m *Model) openLogViewer() tea.Cmd {
	return func() tea.Msg {
		v := logview.New(m.state, logger.Core(), "core logs")
		v.SetContainerWidth(min(80, m.state.Layout.Body.Width))
		v.SetShowBorder(false)
		return dialog.OpenDialog(v)
	}
}

func (m *Model) errorCmd(title, message string) tea.Cmd {
	return notice.NewNoticeCmd(m.state.Layout, title, message, notice.NOTICE_ERROR)
}

func createIPViewer(layout *layout.Layout, schema result.ResultSchema) ui.Component {
	viewer := ipviewer.New(layout, "", nil, schema)
	if t := viewer.Table(); t != nil {
		t.SetKeys(
			table.NewKey([]string{env.KeyTab}, "tab", "next tab", nil),
			table.NewKey([]string{"p"}, "pause", "pause/resume scan", nil),
			table.NewKey([]string{"l"}, "log", "view logs", nil),
		)
	}
	return viewer
}
