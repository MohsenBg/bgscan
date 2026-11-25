// Package footer renders the status bar at the bottom of the workspace.
// It shows the app name and version, the current screen name as the
// status, and live runtime stats. View caches its render and rebuilds it
// only after an update marks the model dirty.
package footer

import (
	"runtime"
	"time"

	"github.com/MohsenBg/bgscan/internal/core/config"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

// Model is the footer component.
type Model struct {
	id          ui.ComponentID
	name        string
	appVersion  string
	status      string
	goroutines  int
	memoryBytes uint64
	sys         uint64
	env         *ui.Env

	// diry marks the cached render as stale, and cache holds that render
	// while diry is false.
	diry  bool
	cache string
}

// RuntimeStats is a snapshot of the Go runtime shown in the footer.
// MemoryBytes is the bytes of allocated heap objects, and Sys is the
// bytes of memory obtained from the OS.
type RuntimeStats struct {
	Goroutines  int
	MemoryBytes uint64
	Sys         uint64
}

type timesTickMsg time.Time

// New creates the footer with the app version from the config package
// and the main menu as the initial status.
func New(env *ui.Env) *Model {
	return &Model{
		id:         ui.NewComponentID(),
		name:       "footer",
		env:        env,
		appVersion: config.AppVersion,
		status:     "Main Menu",
		diry:       true,
	}
}

// ID returns the component ID of the model. Mode returns
// KeyModeNormal, because the footer takes no input. Name returns its
// display name. OnClose returns nil, because the footer has no close
// work of its own.
func (m *Model) ID() ui.ComponentID { return m.id }
func (m *Model) Mode() ui.KeyMode   { return ui.KeyModeNormal }
func (m *Model) Name() string       { return m.name }
func (m *Model) OnClose() tea.Cmd   { return nil }

// Init starts the one second stats ticker.
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
