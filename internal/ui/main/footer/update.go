package footer

import (
	"bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

// Update processes footer-specific messages and refreshes displayed state.
//
// Handled messages:
//   - timesTickMsg: refresh runtime stats and schedule the next tick
//   - UpdateAppVersion: change the displayed app version
//   - UpdateStatus: change the status text
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
