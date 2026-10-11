// Package startup runs preflight checks before entering the workspace.
package startup

import (
	"sync/atomic"

	"charm.land/bubbles/v2/progress"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/MohsenBg/bgscan/internal/logger"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/help"
	"github.com/MohsenBg/bgscan/internal/ui/shared/env"
	"github.com/MohsenBg/bgscan/internal/ui/shared/keys"
	"github.com/MohsenBg/bgscan/internal/ui/shared/layout"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
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
	sidebarWidth     = 18
	titleHeight      = 2
	progressHeight   = 1
	helpHeight       = 2
	progressBarWidth = 24
)

type model struct {
	id          ui.ComponentID
	name        string
	categories  []category
	state       *ui.AppState
	viewport    viewport.Model
	spinner     spinner.Model
	progressBar progress.Model
	width       int
	height      int
	helpSet     help.HelpSet
	help        ui.Component
	logCh       chan tea.Msg
	checked     bool
	fatal       *fatalError
}

type fatalError struct {
	category string
	message  string
}

func (m *model) ID() ui.ComponentID { return m.id }
func (m *model) Mode() env.Mode     { return env.NormalMode }
func (m *model) Name() string       { return m.name }
func (m *model) OnClose() tea.Cmd   { return nil }

// New creates the startup preflight component.
func New(state *ui.AppState) ui.Component {
	th := state.Theme()

	pb := progress.New(
		progress.WithoutPercentage(),
		progress.WithColors(th.ProgressEnd, th.ProgressStart),
	)
	pb.Full = '▬'
	pb.Empty = '─'
	pb.EmptyColor = th.BorderActive

	vp := viewport.New()
	if state.Layout != nil {
		vp.SetWidth(state.Layout.Terminal.Width)
		vp.SetHeight(state.Layout.Terminal.Height)
		pb.SetWidth(state.Layout.Terminal.Width)
	}

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(th.Primary)

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
			{id: "masterdns", label: "MasterDNS"},
			{id: "stormdns", label: "StormDNS"},
			{id: "thefeed", label: "TheFeed"},
			{id: "app", label: "App"},
		},
	}

	deps := state.Deps()
	m.helpSet = startupHelpSet(deps.Keys)
	m.help = help.New(deps, "Startup", m.helpSet)
	m.applyLayout(state.Layout)

	return m
}

func startupHelpSet(k keys.Keys) help.HelpSet {
	scroll := keys.Combine("scroll", k.Nav.Up, k.Nav.Down)

	short := keys.Hints{scroll}
	short = append(short, keys.Short(
		k.Notice.Continue,
		k.Global.Quit,
		k.Global.Help,
	)...)

	long := keys.Hints{scroll}
	long = append(long, keys.Long(
		k.Notice.Continue,
		k.Global.Quit,
		k.Global.Help,
	)...)

	return help.HelpSet{Short: short, Long: long}
}

func rebuildProgressBar(state *ui.AppState, old progress.Model) progress.Model {
	th := state.Theme()

	pb := progress.New(
		progress.WithoutPercentage(),
		progress.WithColors(th.ProgressEnd, th.ProgressStart),
	)
	pb.Full = old.Full
	pb.Empty = old.Empty
	pb.EmptyColor = th.BorderActive
	pb.SetWidth(old.Width())

	return pb
}

func (m *model) applyLayout(l *layout.Layout) {
	if m == nil || l == nil {
		return
	}

	m.width = l.Terminal.Width
	m.height = l.Terminal.Height
	m.progressBar.SetWidth(l.Terminal.Width)

	vpWidth := max(l.Terminal.Width-sidebarWidth-1, 10)
	vpHeight := max(l.Terminal.Height-titleHeight-progressHeight-helpHeight, 3)

	m.viewport.SetWidth(vpWidth)
	m.viewport.SetHeight(vpHeight)

	if h, ok := m.help.(*help.Model); ok {
		h.SetShortWidth(m.width)
	}
}

// Init starts background checks and the spinner/log listeners.
func (m *model) Init() tea.Cmd {
	go m.runChecks()

	if m.state != nil {
		m.applyLayout(m.state.Layout)
	}

	return tea.Batch(
		listenForLogs(m.logCh),
		m.spinner.Tick,
	)
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

func (m *model) runChecks() {
	defer close(m.logCh)

	var abort atomic.Bool

	if !m.runCheck("logger", &abort, func(r *reporter) {
		checkLoggerHealth(r, m.loggerSet())
	}) {
		return
	}

	if !m.runConfigCheck(&abort) {
		return
	}

	if !m.runCheck("xray", &abort, func(r *reporter) {
		checkXrayHealth(r, m.coreLogger())
	}) {
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

	if !m.runCheck("masterdns", &abort, checkMasterDNSHealth) {
		return
	}

	if !m.runCheck("stormdns", &abort, checkStormDNSHealth) {
		return
	}

	if !m.runCheck("thefeed", &abort, checkTheFeedHealth) {
		return
	}

	m.runCheck("app", &abort, checkAppHealth)
}

func (m *model) loggerSet() logger.Set {
	if m.state == nil {
		return logger.Set{}
	}

	return m.state.Log
}

func (m *model) coreLogger() *logger.Logger {
	if m.state == nil {
		return nil
	}

	return m.state.Log.Core
}

// runCheck sends start/end messages around fn and stops on fatal abort.
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

func (m *model) runConfigCheck(abort *atomic.Bool) bool {
	if m.state == nil {
		return true
	}

	m.logCh <- categoryStartMsg{categoryID: "config"}

	r := newReporter("config", m.logCh, abort)
	cfg, store := checkConfigHealth(r, m.state.Log)

	if abort.Load() {
		return false
	}

	if m.state.Program != nil {
		m.state.Program.Send(configLoadedMsg{cfg: cfg, store: store})
	}

	m.logCh <- categoryEndMsg{categoryID: "config", status: r.status}

	return true
}

func (m *model) startCategory(categoryID string) {
	for i := range m.categories {
		if m.categories[i].id != categoryID {
			continue
		}

		m.categories[i].started = true
		m.categories[i].status = catRunning

		return
	}
}

func (m *model) endCategory(categoryID string, status categoryStatus) {
	for i := range m.categories {
		if m.categories[i].id != categoryID {
			continue
		}

		m.categories[i].status = status

		return
	}
}

func (m *model) appendLine(categoryID string, status categoryStatus, line string) {
	for i := range m.categories {
		cat := &m.categories[i]
		if cat.id != categoryID {
			continue
		}

		prefix := m.statusPrefixStyle(status).Render(statusLabel(status))
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

// overallStatus returns the worst terminal status across categories.
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
			return m.pendingDotStyle().Render("○")
		}

		return m.spinner.View()
	}

	return m.statusPrefixStyle(cat.status).Render("●")
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

func statusLabel(status categoryStatus) string {
	switch status {
	case catRunning:
		return "[INFO]"
	case catWait:
		return "[WAIT]"
	case catOK:
		return "[SUCCESS]"
	case catWarn:
		return "[WARN]"
	case catError:
		return "[ERROR]"
	default:
		return "[INFO]"
	}
}
