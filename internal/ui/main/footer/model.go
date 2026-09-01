package footer

import (
	"github.com/MohsenBg/bgscan/internal/core/config"
	"github.com/MohsenBg/bgscan/internal/ui/shared/env"
	"github.com/MohsenBg/bgscan/internal/ui/shared/layout"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
	"runtime"
	"time"

	tea "charm.land/bubbletea/v2"
)

type Model struct {
	layout      *layout.Layout
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

func New(l *layout.Layout) *Model {
	return &Model{
		id:         ui.NewComponentID(),
		name:       "footer",
		layout:     l,
		appVersion: config.AppVersion,
		status:     "Main Menu",
	}
}

func (m *Model) ID() ui.ComponentID { return m.id }
func (m *Model) Mode() env.Mode     { return env.NormalMode }
func (m *Model) Name() string       { return m.name }
func (m *Model) OnClose() tea.Cmd   { return nil }

func (m *Model) Init() tea.Cmd { return tickCmd() }

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
