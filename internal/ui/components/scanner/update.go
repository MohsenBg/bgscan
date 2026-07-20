package scanner

import (
	"bgscan/internal/logger"
	"bgscan/internal/ui/components/basic/confirm"
	logview "bgscan/internal/ui/components/basic/logview"
	"bgscan/internal/ui/components/basic/notice"
	"bgscan/internal/ui/components/basic/progress"
	"bgscan/internal/ui/shared/dialog"
	"bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
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

	// Scanner.Run() returned an error before the scan started
	case scanErrorMsg:
		m.onError(msg.err)
		return m, nil

	// Scanner.Close() finished
	case scanClosedMsg:
		if msg.err != nil {
			cmds = append(cmds, m.errorCmd("Failed to close scanner", msg.err.Error()))
		}
		cmds = append(cmds, func() tea.Msg { return ui.ResetComponentStacksMsg{} })
		return m, tea.Batch(cmds...)

	// Global keybindings
	case tea.KeyMsg:
		cmds = append(cmds, m.handleKey(msg))
	}

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
				return tea.BatchMsg{m.asyncClose()}
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
		return dialog.OpenDialog(v)
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
//   Close (async)
// ────────────────────────────────────────────────────────────
//

// asyncClose returns a tea.Cmd that runs Scanner.Close() on a
// goroutine and delivers scanClosedMsg when it finishes.
func (m *Model) asyncClose() tea.Cmd {
	return func() tea.Msg {
		ch := make(chan error, 1)
		go func() { ch <- m.scn.Close() }()
		return scanClosedMsg{err: <-ch}
	}
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
		m.mu.Lock()
		err := m.scanError
		shown := m.errorShown
		if err != nil && !shown {
			m.errorShown = true
		}
		m.mu.Unlock()

		if err != nil && !shown {
			cmds = append(cmds, m.errorCmd(
				"Error while scanning",
				err.Error(),
			))
		}

	case StatusPreProcess, StatusWaiting:
		// No UI update needed
	}

	return tea.Batch(cmds...)
}
