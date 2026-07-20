package scanner

import (
	"sort"
	"sync"
	"time"

	"bgscan/internal/core/config"
	"bgscan/internal/core/result"
	"bgscan/internal/core/scanner"
	"bgscan/internal/core/scanner/engine"
	"bgscan/internal/ui/components/basic/progress"
	"bgscan/internal/ui/components/basic/table"
	"bgscan/internal/ui/components/basic/tabs"
	"bgscan/internal/ui/components/tables/ipviewer"
	"bgscan/internal/ui/shared/env"
	"bgscan/internal/ui/shared/layout"
	"bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type StageStatus int

const (
	StatusWaiting StageStatus = iota
	StatusPreProcess
	StatusScanning
	StatusEnded
	StatusError
)

type Model struct {
	// UI
	id     ui.ComponentID
	name   string
	layout *layout.Layout
	tabs   ui.Component

	// Scanner
	scn        *scanner.Scanner
	stages     []scanner.StageConfig
	stageCount int
	maxIPs     int

	// Per-stage UI
	progress   []ui.Component
	ipViewers  []ui.Component
	currentTab int

	// Results
	results [][]result.Result
	batch   [][]result.Result

	// State
	mu           sync.Mutex
	status       []StageStatus
	progressInfo []engine.Progress
	scanError    error
}

func New(layout *layout.Layout, maxIPs int, scn *scanner.Scanner) *Model {
	stages := scn.GetStages()
	n := len(stages)

	m := &Model{
		id:           ui.NewComponentID(),
		name:         "Scanner",
		layout:       layout,
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
		m.ipViewers[i] = createIPViewer(layout, stage.Probe.Schema())
		m.progress[i] = progress.New(layout)
		m.results[i] = make([]result.Result, 0, maxIPs)
		m.batch[i] = make([]result.Result, 0, 128)
		m.status[i] = StatusWaiting
		tabsList[i] = tabs.NewTab(stage.Probe.Schema().Name, i)
	}

	m.tabs = tabs.New(layout, tabsList, func(idx int, _ tabs.Tab[int]) tea.Cmd {
		m.currentTab = idx
		return m.immediateTick()
	})

	paddingY := lipgloss.Height(m.renderProgress(m.currentTab)) + lipgloss.Height(m.tabs.View())
	for _, v := range m.ipViewers {
		if viewer, ok := v.(*ipviewer.Model); ok {
			viewer.Table().SetPaddingY(paddingY)
		}
	}

	return m
}

func (m *Model) ID() ui.ComponentID { return m.id }
func (m *Model) Name() string       { return m.name }
func (m *Model) Mode() env.Mode     { return env.ScanMode }
func (m *Model) OnClose() tea.Cmd   { return nil }

func (m *Model) Init() tea.Cmd {
	for i := range m.stages {
		m.stages[i].AddHooks(engine.ScanHooks{
			OnError:    m.onError,
			OnProgress: m.onProgress(i),
			OnSuccess:  m.onSuccess(i),
			OnScanEnd:  m.onScanEnd(i),
		})
	}

	m.status[0] = StatusPreProcess

	go func() {
		if err := m.scn.Run(); err != nil {
			m.onError(err)
		}
	}()

	return m.tick()
}

func (m *Model) tick() tea.Cmd {
	interval := config.Get().General.StatusInterval.Duration()
	return tea.Tick(interval, func(time.Time) tea.Msg { return tickMsg{} })
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
}

func (m *Model) onScanEnd(i int) func() {
	return func() {
		m.mu.Lock()
		m.status[i] = StatusEnded
		m.mu.Unlock()
	}
}

// mergeBatch merges staged results into the main result set.
// Uses swap-slice pattern to minimize lock duration.
func (m *Model) mergeBatch() {
	for i, stage := range m.stages {
		m.mu.Lock()
		if len(m.batch[i]) == 0 {
			m.mu.Unlock()
			continue
		}

		newBatch := m.batch[i]
		m.batch[i] = m.batch[i][:0]
		m.mu.Unlock()
		for i, batch := range newBatch {
			rec := batch.ToRecord()
			normalizedRs, err := stage.Probe.Schema().Parser(rec)
			if err != nil {
				continue
			}
			newBatch[i] = normalizedRs
		}

		m.results[i] = append(m.results[i], newBatch...)

		sort.SliceStable(m.results[i], func(a, b int) bool {
			scoreA := m.results[i][a].Score()
			scoreB := m.results[i][b].Score()

			if scoreA != scoreB {
				return scoreA > scoreB
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

func (m *Model) currentError() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.scanError
}

func createIPViewer(layout *layout.Layout, schema result.ResultSchema) ui.Component {
	viewer := ipviewer.New(layout, "", nil, schema)

	viewer.Table().SetKeys(
		table.NewKey([]string{env.KeyTab}, "tab", "next tab", nil),
		table.NewKey([]string{"p"}, "pause", "pause/resume scan", nil),
		table.NewKey([]string{"l"}, "log", "view logs", nil),
	)

	return viewer
}

func (m *Model) immediateTick() tea.Cmd {
	return func() tea.Msg { return immediateTickMsg{} }
}

func (m *Model) forceResize() tea.Cmd {
	return func() tea.Msg {
		return tea.WindowSizeMsg{
			Width:  m.layout.Terminal.Width,
			Height: m.layout.Terminal.Height,
		}
	}
}
