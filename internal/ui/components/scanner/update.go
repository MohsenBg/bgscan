package scanner

import (
	"errors"

	"github.com/MohsenBg/bgscan/internal/ui/components/basic/confirm"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/notice"
	"github.com/MohsenBg/bgscan/internal/ui/shared/dialog"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyPressMsg:
		if cmd := m.warnNonLatin(msg); cmd != nil {
			return m, cmd
		}

		if cmd, isKeyUsed := m.handleKey(msg); isKeyUsed {
			return m, cmd
		}
		return m, m.updateCurrentStage(msg)

	case scanCloseResultMsg:
		if msg.err != nil {
			return m, m.handleScanCloseError(msg.err)
		}
		return m, func() tea.Msg { return ui.ResetComponentStacksMsg{} }

	}

	switch msg := msg.(type) {
	// window size change
	case tea.WindowSizeMsg:
		m.updateLayout()
		return m, nil

	case pulseMsg:
		m.pulse += m.pulseDir
		if m.pulse >= pulseSteps-1 {
			m.pulse = pulseSteps - 1
			m.pulseDir = -1
		} else if m.pulse <= 0 {
			m.pulse = 0
			m.pulseDir = 1
		}
		return m, pulseCmd()

	case scanErrorMsg:
		return m, m.handleCriticalError(msg.err)

	case stageErrorMsg:
		if st := m.getStage(msg.id); st != nil && !st.finished() {
			st.status = StatusError
		}
		return m, m.handleCriticalError(msg.err)

	case stageProgressMsg:
		stage := m.getStage(msg.id)
		if stage == nil {
			return m, m.handleCriticalError(errors.New("stage not found"))
		}
		stage.mergeBatch()
		stage.setProgress(msg.p)
		return m, stage.updateProgressBar()

	case stageEndMsg:
		stage := m.getStage(msg.id)
		if stage == nil {
			return m, m.handleCriticalError(errors.New("stage not found"))
		}
		if stage.status != StatusError {
			stage.status = StatusEnded
		}
		stage.mergeBatch()
		return m, stage.updateProgressBar()
	}

	return m, m.updateAllStage(msg)
}

// handleKey processes global keybindings.
func (m *Model) handleKey(msg tea.KeyMsg) (tea.Cmd, bool) {

	k := m.session.Env().Keys.Scan
	if k.Cancel.Matches(msg) {
		return confirm.ConfirmCmd(m.session.Env(),
			"Do you want to exit the scan?",
			m.closeScan,
			false,
			dialog.WithTitle("Exit Scan"),
		), true
	}

	if k.TogglePause.Matches(msg) {
		m.togglePause()
		return nil, true
	}

	if k.ToggleLogs.Matches(msg) {
		return m.openLogViewer(), true
	}

	return nil, false
}

// updateComponents forwards the message to the active stage and the tab bar.
func (m *Model) updateCurrentStage(msg tea.Msg) tea.Cmd {
	var tabCmd tea.Cmd
	m.tabs, tabCmd = m.tabs.Update(msg)
	return tea.Batch(m.current().update(msg), tabCmd)
}

// updateAllStage forwards the message to every stage and the tab bar.
func (m *Model) updateAllStage(msg tea.Msg) tea.Cmd {
	var tabCmd tea.Cmd
	m.tabs, tabCmd = m.tabs.Update(msg)
	cmds := []tea.Cmd{tabCmd}
	for _, s := range m.stages {
		cmds = append(cmds, s.update(msg))
	}
	return tea.Batch(cmds...)
}

func (m *Model) handleCriticalError(err error) tea.Cmd {
	return notice.NewNoticeCmd(
		m.session.Env(),
		"Scanner Error",
		err.Error(),
		notice.NOTICE_ERROR,
		dialog.WithOnClose(m.closeScan),
	)
}

func (m *Model) handleScanCloseError(err error) tea.Cmd {
	return notice.NewNoticeCmd(
		m.session.Env(),
		"Scanner Error",
		err.Error(),
		notice.NOTICE_ERROR,
		dialog.WithOnClose(func() tea.Msg { return ui.ResetComponentStacksMsg{} }),
	)
}

func (m *Model) warnNonLatin(key tea.KeyPressMsg) tea.Cmd {
	if !notice.IsNonLatin(key.Text) {
		return nil
	}

	return notice.NewKeyboardLayoutNoticeCmd(m.session.Env(), func() tea.Msg {
		return nil
	})
}
