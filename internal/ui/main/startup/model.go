package startup

import (
	"sync/atomic"

	"github.com/MohsenBg/bgscan/internal/ui/shared/env"
	"github.com/MohsenBg/bgscan/internal/ui/shared/layout"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
	"github.com/MohsenBg/bgscan/internal/ui/theme"

	"charm.land/bubbles/v2/progress"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
)

type categoryStatus int

const (
	catRunning categoryStatus = iota
	catWait
	catOK
	catWarn
	catError
)

type category struct {
	id      string
	label   string
	status  categoryStatus
	started bool
	lines   []string
}

const (
	sidebarWidth     = 15
	titleHeight      = 2
	progressHeight   = 1
	helpHeight       = 2
	progressBarWidth = 24
)

type model struct {
	id         ui.ComponentID
	name       string
	categories []category
	state      *ui.AppState

	viewport    viewport.Model
	spinner     spinner.Model
	progressBar progress.Model
	width       int
	height      int

	showHelp bool

	logCh   chan tea.Msg
	checked bool
	fatal   *fatalError
}

type fatalError struct {
	category string
	message  string
}

func (m *model) ID() ui.ComponentID { return m.id }
func (m *model) Mode() env.Mode     { return env.NormalMode }
func (m *model) Name() string       { return m.name }
func (m *model) OnClose() tea.Cmd   { return nil }

func New(state *ui.AppState) ui.Component {
	pb := progress.New(
		progress.WithoutPercentage(),
		progress.WithColors(theme.Current().ProgressEnd, theme.Current().ProgressStart),
	)
	pb.Full = '▬'
	pb.Empty = '─'
	pb.EmptyColor = theme.Current().BorderActive

	vp := viewport.New()
	if state.Layout != nil {
		vp.SetWidth(state.Layout.Terminal.Width)
		vp.SetHeight(state.Layout.Terminal.Height)
		pb.SetWidth(state.Layout.Terminal.Width)
	}

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = spinnerStyle()

	m := &model{
		id:          ui.NewComponentID(),
		name:        "startup",
		state:       state,
		viewport:    vp,
		spinner:     sp,
		progressBar: pb,
		logCh:       make(chan tea.Msg, 64),
		categories: []category{
			{id: "logger", label: "Logger"},
			{id: "config", label: "Config"},
			{id: "xray", label: "Xray"},
			{id: "dnstt", label: "DNSTT"},
			{id: "slipstream", label: "Slipstream"},
			{id: "vaydns", label: "Vaydns"},
			{id: "app", label: "App"},
		},
	}

	if state != nil && state.Layout != nil {
		m.applyLayout(state.Layout)
	}

	return m
}

func (m *model) applyLayout(l *layout.Layout) {
	m.width = l.Terminal.Width
	m.height = l.Terminal.Height
	m.progressBar.SetWidth(l.Terminal.Width)

	vpWidth := max(l.Terminal.Width-sidebarWidth-1, 10)
	vpHeight := max(l.Terminal.Height-titleHeight-progressHeight-helpHeight, 3)

	m.viewport.SetWidth(vpWidth)
	m.viewport.SetHeight(vpHeight)
}

// runCheck sends categoryStartMsg, runs fn, then sends categoryEndMsg with
// the reporter's final status — unless the run was aborted (critical error),
// in which case the fatal logMsg already sent is enough and no end message
// is emitted. Returns false when the pipeline should stop.
func (m *model) runCheck(id string, abort *atomic.Bool, fn func(r *reporter)) bool {
	m.logCh <- categoryStartMsg{categoryID: id}

	r := newReporter(id, m.logCh, abort)
	fn(r)

	if abort.Load() {
		return false
	}

	m.logCh <- categoryEndMsg{categoryID: id, status: r.status}
	return true
}

func (m *model) Init() tea.Cmd {
	go func() {
		defer close(m.logCh)

		var abort atomic.Bool

		if !m.runCheck("logger", &abort, checkLoggerHealth) {
			return
		}

		if m.state != nil {
			m.logCh <- categoryStartMsg{categoryID: "config"}

			r := newReporter("config", m.logCh, &abort)
			cfg, store := checkConfigHealth(r)

			if abort.Load() {
				return
			}

			// Sent directly through the program instead of a mutex-guarded field.
			m.state.Program.Send(configLoadedMsg{cfg: cfg, store: store})
			m.logCh <- categoryEndMsg{categoryID: "config", status: r.status}
		}

		if !m.runCheck("xray", &abort, checkXrayHealth) {
			return
		}
		if !m.runCheck("dnstt", &abort, checkDNSTTHealth) {
			return
		}
		if !m.runCheck("slipstream", &abort, checkSlipstreamHealth) {
			return
		}
		if !m.runCheck("vaydns", &abort, checkVayDNSHealth) {
			return
		}
		m.runCheck("app", &abort, checkAppHealth)
	}()

	m.applyLayout(m.state.Layout)
	return tea.Batch(listenForLogs(m.logCh), m.spinner.Tick)
}

func listenForLogs(ch <-chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return allChecksDoneMsg{}
		}
		return msg
	}
}

func (m *model) startCategory(categoryID string) {
	for i := range m.categories {
		if m.categories[i].id == categoryID {
			m.categories[i].started = true
			m.categories[i].status = catRunning
			return
		}
	}
}

func (m *model) endCategory(categoryID string, status categoryStatus) {
	for i := range m.categories {
		if m.categories[i].id == categoryID {
			m.categories[i].status = status
			return
		}
	}
}

func (m *model) appendLine(categoryID string, status categoryStatus, line string) {
	for i := range m.categories {
		cat := &m.categories[i]
		if cat.id != categoryID {
			continue
		}

		var prefix string
		switch status {
		case catRunning:
			prefix = statusPrefixStyle(catRunning).Render("[INFO]")
		case catWait:
			prefix = statusPrefixStyle(catWait).Render("[WAIT]")
		case catOK:
			prefix = statusPrefixStyle(catOK).Render("[SUCCESS]")
		case catWarn:
			prefix = statusPrefixStyle(catWarn).Render("[WARN]")
		case catError:
			prefix = statusPrefixStyle(catError).Render("[ERROR]")
		}
		cat.lines = append(cat.lines, prefix+" "+line)
		return
	}
}

func (m *model) categoryLabel(categoryID string) string {
	for _, cat := range m.categories {
		if cat.id == categoryID {
			return cat.label
		}
	}
	return categoryID
}

func (m *model) overallStatus() categoryStatus {
	worst := catOK
	anyRunning := false
	for _, cat := range m.categories {
		if cat.status == catRunning {
			anyRunning = true
		}
		if cat.status > worst {
			worst = cat.status
		}
	}
	if anyRunning && worst == catOK {
		return catRunning
	}
	return worst
}

func (m *model) statusIndicator(cat category) string {
	if cat.status == catRunning {
		if !cat.started {
			return pendingDotStyle().Render("○")
		}
		return m.spinner.View()
	}
	return statusPrefixStyle(cat.status).Render("●")
}

func (m *model) completedCount() int {
	n := 0
	for _, cat := range m.categories {
		if cat.status != catRunning {
			n++
		}
	}
	return n
}

func (m *model) progressTarget() float64 {
	total := len(m.categories)
	if total == 0 {
		return 1
	}
	return float64(m.completedCount()) / float64(total)
}
