package footer

import (
	"github.com/MohsenBg/bgscan/internal/core/config"
	"github.com/MohsenBg/bgscan/internal/logger"
	"github.com/MohsenBg/bgscan/internal/ui/shared/env"
	"github.com/MohsenBg/bgscan/internal/ui/shared/layout"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
	"github.com/MohsenBg/bgscan/internal/ui/theme"
	"runtime"
	"time"

	tea "charm.land/bubbletea/v2"
)

type Model struct {
	layout      *layout.Layout
	theme       *theme.Theme
	log         logger.Set
	id          ui.ComponentID
	name        string
	appVersion  string
	status      string
	goroutines  int
	memoryBytes uint64
	sys         uint64
}

type RuntimeStats struct {
	Goroutines  int
	MemoryBytes uint64
	Sys         uint64
}

type timesTickMsg time.Time

func New(deps ui.Deps) *Model {
	return &Model{
		id:         ui.NewComponentID(),
		name:       "footer",
		layout:     deps.Layout,
		theme:      deps.Theme,
		log:        deps.Log,
		appVersion: config.AppVersion,
		status:     "Main Menu",
	}
}

func (m *Model) ID() ui.ComponentID { return m.id }
func (m *Model) Mode() env.Mode     { return env.NormalMode }
func (m *Model) Name() string       { return m.name }
func (m *Model) OnClose() tea.Cmd   { return nil }

func (m *Model) Init() tea.Cmd { return tickCmd() }

func (m *Model) Theme() *theme.Theme { return m.theme }

func (m *Model) SetTheme(th *theme.Theme) { m.theme = th }

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg { return timesTickMsg(t) })
}

func getRuntimeStats() RuntimeStats {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	return RuntimeStats{
		Goroutines:  runtime.NumGoroutine(),
		MemoryBytes: mem.Alloc,
		Sys:         mem.Sys,
	}
}
