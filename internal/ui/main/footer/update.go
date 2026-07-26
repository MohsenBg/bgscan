package footer

import (
	"bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

// Update handles incoming Bubble Tea messages and updates
// the footer component state accordingly.
//
// Responsibilities:
//   - Refresh runtime statistics on periodic tick messages.
//   - Update application version when requested.
//   - Update current application status.
//
// Returns the updated component and an optional command
// to be executed by Bubble Tea.
func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	switch msg := msg.(type) {

	// Periodic runtime update
	case timesTickMsg:
		stats := getRuntimeStats()
		m.goroutines = stats.Goroutines
		m.memoryBytes = stats.MemoryBytes
		m.sys = stats.Sys

		// Schedule the next tick to keep stats updated
		return m, tickCmd()

	// Update application version displayed in the footer
	case UpdateAppVersion:
		m.appVersion = msg.AppVersion
		return m, nil

	// Update current application status
	case UpdateStatus:
		m.status = msg.Status
		return m, nil
	}

	return m, nil
}
