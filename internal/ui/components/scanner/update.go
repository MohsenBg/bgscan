package scanner

import (
	"fmt"

	"bgscan/internal/logger"
	"bgscan/internal/ui/components/basic/confirm"
	logview "bgscan/internal/ui/components/basic/logview"
	"bgscan/internal/ui/components/basic/notice"
	"bgscan/internal/ui/components/basic/progress"
	"bgscan/internal/ui/shared/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	// Regular periodic update
	case tickMsg:
		cmds = append(cmds, m.updateTick(), m.tick())
		return m, tea.Batch(cmds...)

	// Instant refresh
	case immediateTickMsg:
		cmds = append(cmds, m.updateTick(), m.forceResize())
		return m, tea.Batch(cmds...)

	// Pause toggle via UI
	case TogglePauseMsg:
		m.togglePause()
		return m, nil

	// Global keybindings
	case tea.KeyMsg:
		cmds = append(cmds, m.handleKey(msg))
	}

	// Component updates (tab UI, progress, IP viewers)
	cmds = append(cmds, m.updateComponents(msg))
	return m, tea.Batch(cmds...)
}

//
// ────────────────────────────────────────────────────────────
//   Key Handling
// ────────────────────────────────────────────────────────────
//

func (m *Model) handleKey(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {

	case "q", "b":
		return confirm.ConfirmCmd(
			m.layout,
			"Do you want to exit the scan?",
			func() tea.Msg {
				m.scn.Close()
				return ui.ResetComponentStacksMsg{}
			},
			false,
		)

	case "p":
		m.togglePause()
		return nil

	case "l":
		return m.openLogViewer()
	}

	return nil
}

//
// ────────────────────────────────────────────────────────────
//   Component Update Routing
// ────────────────────────────────────────────────────────────
//

func (m *Model) updateComponents(msg tea.Msg) tea.Cmd {
	idx := m.currentTab
	var tCmd, pCmd, tabCmd tea.Cmd

	m.ipViewers[idx], tCmd = m.ipViewers[idx].Update(msg)
	m.progress[idx], pCmd = m.progress[idx].Update(msg)
	m.tabs, tabCmd = m.tabs.Update(msg)

	return tea.Batch(tCmd, pCmd, tabCmd)
}

//
// ────────────────────────────────────────────────────────────
//   Pause Toggle
// ────────────────────────────────────────────────────────────
//

func (m *Model) togglePause() {
	if m.scn.IsPaused() {
		m.scn.Resume()
	} else {
		m.scn.Pause()
	}
}

//
// ────────────────────────────────────────────────────────────
//   Log Viewer Overlay
// ────────────────────────────────────────────────────────────
//

func (m *Model) openLogViewer() tea.Cmd {
	return func() tea.Msg {
		v := logview.New(m.layout, logger.Core(), "core logs")
		v.SetContainerWidth(min(80, m.layout.Body.Width))
		v.SetShowBorder(false)
		return ui.AddNewOverlay(v, ui.Center, ui.Center, 0, 0)
	}
}

//
// ────────────────────────────────────────────────────────────
//   Notices
// ────────────────────────────────────────────────────────────
//

func (m *Model) errorCmd(title, msg string) tea.Cmd {
	return notice.NewNoticeCmd(m.layout, title, msg, notice.NOTICE_ERROR)
}

//
// ────────────────────────────────────────────────────────────
//   Tick Update Handler
// ────────────────────────────────────────────────────────────
//

func (m *Model) updateTick() tea.Cmd {
	var cmds []tea.Cmd

	m.mergeBatch()

	idx := m.currentTab

	switch m.currentStatus() {

	case StatusScanning:
		pct := m.currentProgress()
		cmds = append(cmds, progress.UpdateProgressMsg{
			ID:       m.progress[idx].ID(),
			Progress: pct,
		}.Cmd())

	case StatusEnded:
		cmds = append(cmds, progress.UpdateProgressMsg{
			ID:       m.progress[idx].ID(),
			Progress: 1,
		}.Cmd())

	case StatusError:
		if err := m.currentError(); err != nil {
			cmds = append(cmds, m.errorCmd(
				"Error while scanning",
				fmt.Sprintf("%v", err),
			))
		}

	case StatusPreProcess, StatusWaiting:
		// No UI update needed
	}

	return tea.Batch(cmds...)
}
