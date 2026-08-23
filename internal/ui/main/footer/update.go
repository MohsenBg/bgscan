package footer

import (
	"bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case timesTickMsg:
		stats := getRuntimeStats()
		m.goroutines = stats.Goroutines
		m.memoryBytes = stats.MemoryBytes
		m.sys = stats.Sys
		return m, tickCmd()

	case UpdateAppVersion:
		m.appVersion = msg.AppVersion
		return m, nil

	case UpdateStatus:
		m.status = msg.Status
		return m, nil
	}
	return m, nil
}
