package startup

import (
	"bgscan/internal/ui/shared/ui"

	"charm.land/bubbles/v2/progress"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
)

func (m *model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case categoryStartMsg:
		m.startCategory(msg.categoryID)
		m.syncViewport()
		return m, listenForLogs(m.logCh)

	case categoryEndMsg:
		m.endCategory(msg.categoryID, msg.status)
		m.syncViewport()
		progressCmd := m.progressBar.SetPercent(m.progressTarget())
		return m, tea.Batch(listenForLogs(m.logCh), progressCmd)

	case logMsg:
		m.appendLine(msg.categoryID, msg.status, msg.line)
		m.syncViewport()

		if msg.critical {
			m.fatal = &fatalError{
				category: m.categoryLabel(msg.categoryID),
				message:  msg.line,
			}
			return m, nil
		}

		return m, listenForLogs(m.logCh)

	case configLoadedMsg:
		if m.state != nil {
			m.state.Config = msg.cfg
			m.state.Store = msg.store
		}
		return m, nil

	case allChecksDoneMsg:
		m.checked = true
		m.syncViewport()
		progressCmd := m.progressBar.SetPercent(m.progressTarget())
		return m, progressCmd

	case spinner.TickMsg:
		if m.checked || m.fatal != nil {
			return m, nil
		}
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		m.syncViewport()
		return m, cmd

	case progress.FrameMsg:
		var cmd tea.Cmd
		m.progressBar, cmd = m.progressBar.Update(msg)
		return m, cmd

	case tea.WindowSizeMsg:
		if m.state != nil && m.state.Layout != nil {
			m.applyLayout(m.state.Layout)
		}
		m.syncViewport()
		return m, nil

	case tea.KeyMsg:
		if m.fatal != nil {
			return m, tea.Quit
		}
		if m.showHelp {
			m.showHelp = false
			return m, nil
		}

		switch msg.String() {
		case "?":
			m.showHelp = true
			return m, nil
		case "q":
			return m, tea.Quit
		case "up", "k":
			m.viewport.ScrollUp(1)
			return m, nil
		case "down", "j":
			m.viewport.ScrollDown(1)
			return m, nil
		case "enter":
			if m.checked {
				return m, func() tea.Msg { return StartupChecksDoneMsg{} }
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}
