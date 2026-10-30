package footer

import (
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
	"github.com/MohsenBg/bgscan/internal/ui/theme"

	tea "charm.land/bubbletea/v2"
)

func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg, theme.ThemeChanged:
		m.diry = true
		return m, nil

	case timesTickMsg:
		stats := getRuntimeStats()
		m.goroutines = stats.Goroutines
		m.memoryBytes = stats.MemoryBytes
		m.sys = stats.Sys
		m.diry = true
		return m, tickCmd()

	case UpdateAppVersion:
		m.appVersion = msg.AppVersion
		m.diry = true
		return m, nil

	case UpdateStatus:
		m.status = msg.Status
		m.diry = true
		return m, nil
	}
	return m, nil
}
