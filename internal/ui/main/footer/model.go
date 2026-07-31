// Package footer displays app metadata, status, and runtime stats.
package footer

import (
	"bgscan/internal/core/config"
	"bgscan/internal/ui/shared/env"
	"bgscan/internal/ui/shared/layout"
	"bgscan/internal/ui/shared/ui"
	"runtime"
	"time"

	tea "charm.land/bubbletea/v2"
)

// Model is the footer component responsible for displaying runtime
// information and current application status.
type Model struct {
	layout *layout.Layout

	id         ui.ComponentID
	name       string
	appVersion string
	status     string

	goroutines  int
	memoryBytes uint64
	sys         uint64
}

// RuntimeStats contains Go runtime metrics collected for display.
type RuntimeStats struct {
	Goroutines  int
	MemoryBytes uint64
	Sys         uint64
}

// timesTickMsg triggers periodic runtime stat refreshes.
type timesTickMsg time.Time

// New creates the footer model.
func New(l *layout.Layout) *Model {
	return &Model{
		id:         ui.NewComponentID(),
		name:       "footer",
		layout:     l,
		appVersion: config.AppVersion,
		status:     "Main Menu",
	}
}

// ID returns the component identifier.
func (m *Model) ID() ui.ComponentID {
	return m.id
}

// Name returns the component name.
func (m *Model) Name() string {
	return m.name
}

// Mode returns the footer interaction mode.
func (m *Model) Mode() env.Mode {
	return env.NormalMode
}

// Init starts the periodic runtime stats ticker.
func (m *Model) Init() tea.Cmd {
	return tickCmd()
}

// OnClose is a no-op for the footer.
func (m *Model) OnClose() tea.Cmd {
	return nil
}

// tickCmd schedules a runtime stats refresh every second.
func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return timesTickMsg(t)
	})
}

// getRuntimeStats samples goroutine count and memory usage from the Go runtime.
func getRuntimeStats() RuntimeStats {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	return RuntimeStats{
		Goroutines:  runtime.NumGoroutine(),
		MemoryBytes: mem.Alloc,
		Sys:         mem.Sys,
	}
}
