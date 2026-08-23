package scanner

import (
	"bgscan/internal/logger"
	"bgscan/internal/ui/components/basic/confirm"
	"bgscan/internal/ui/components/basic/progress"
	"bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	var cmds []tea.Cmd

	if m.closing {
		// Block all input while closing, but still forward resize to keep layout.
		switch msg := msg.(type) {
		case tea.WindowSizeMsg:
			m.recalcTableSizes()
			return m, nil

		case scanClosedMsg:
			m.closing = false
			if msg.err != nil {
				logger.UIError("Failed to close scanner: %v", msg.err)
				cmds = append(cmds, m.errorCmd("Failed to close scanner", msg.err.Error()))
			}
			logger.UIInfo("Scan closed")
			cmds = append(cmds, func() tea.Msg { return ui.ResetComponentStacksMsg{} })
			return m, tea.Batch(cmds...)

		default:
			return m, nil
		}
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.recalcTableSizes()
		return m, nil

	case tickMsg:
		cmds = append(cmds, m.updateTick(), m.tick())
		return m, tea.Batch(cmds...)

	case immediateTickMsg:
		cmds = append(cmds, m.updateTick(), m.forceResize())
		return m, tea.Batch(cmds...)

	case TogglePauseMsg:
		m.togglePause()
		return m, nil

	case scanErrorMsg:
		logger.UIError("Scan error: %v", msg.err)
		m.onError(msg.err)
		return m, nil

	case tea.KeyMsg:
		cmds = append(cmds, m.handleKey(msg))
	}

	cmds = append(cmds, m.updateComponents(msg))
	return m, tea.Batch(cmds...)
}

// handleKey processes global keybindings.
func (m *Model) handleKey(msg tea.KeyMsg) tea.Cmd {
	if m.closing {
		return nil
	}
	switch msg.String() {
	case "q", "b":
		return confirm.ConfirmCmd(
			m.state.Layout,
			"Do you want to exit the scan?",
			func() tea.Msg {
				m.closing = true // block further input
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

// updateComponents forwards the message to the active tab's components.
func (m *Model) updateComponents(msg tea.Msg) tea.Cmd {
	idx := m.currentTab
	var tCmd, pCmd, tabCmd tea.Cmd

	m.ipViewers[idx], tCmd = m.ipViewers[idx].Update(msg)
	m.progress[idx], pCmd = m.progress[idx].Update(msg)
	m.tabs, tabCmd = m.tabs.Update(msg)

	return tea.Batch(tCmd, pCmd, tabCmd)
}

// updateTick merges new results and updates the progress bar.
func (m *Model) updateTick() tea.Cmd {
	if m.closing {
		return nil
	}
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
			cmds = append(cmds, m.errorCmd("Error while scanning", err.Error()))
		}
	case StatusPreProcess, StatusWaiting:
	}
	return tea.Batch(cmds...)
}
