package startup

import (
	"charm.land/bubbles/v2/progress"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"

	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
)

func (m *model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	if _, ok := msg.(ui.ThemeChangedMsg); ok {
		ui.SyncState(msg, m.state)
		m.progressBar = rebuildProgressBar(m.state, m.progressBar)
		m.spinner.Style = m.spinnerStyle()

		var cmd tea.Cmd
		if m.help != nil {
			m.help, cmd = m.help.Update(msg)
		}

		m.syncViewport()
		return m, cmd
	}

	switch t := msg.(type) {
	case categoryStartMsg:
		m.startCategory(t.categoryID)
		m.syncViewport()
		return m, listenForLogs(m.logCh)

	case categoryEndMsg:
		m.endCategory(t.categoryID, t.status)
		m.syncViewport()
		return m, tea.Batch(
			listenForLogs(m.logCh),
			m.progressBar.SetPercent(m.progressTarget()),
		)

	case logMsg:
		m.appendLine(t.categoryID, t.status, t.line)
		m.syncViewport()

		if t.critical {
			m.fatal = &fatalError{
				category: m.categoryLabel(t.categoryID),
				message:  t.line,
			}
			return m, nil
		}

		return m, listenForLogs(m.logCh)

	case configLoadedMsg:
		if m.state != nil {
			m.state.Config = t.cfg
			m.state.Store = t.store
		}

		return m, nil

	case allChecksDoneMsg:
		m.checked = true
		m.syncViewport()
		return m, m.progressBar.SetPercent(m.progressTarget())

	case spinner.TickMsg:
		if m.checked || m.fatal != nil {
			return m, nil
		}

		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(t)
		m.syncViewport()

		return m, cmd

	case progress.FrameMsg:
		var cmd tea.Cmd
		m.progressBar, cmd = m.progressBar.Update(t)

		return m, cmd

	case tea.WindowSizeMsg:
		if m.state != nil && m.state.Layout != nil {
			m.applyLayout(m.state.Layout)
		}

		m.syncViewport()

	case tea.KeyMsg:
		if m.fatal != nil {
			return m, tea.Quit
		}

		if m.state == nil {
			break
		}

		deps := m.state.Deps()
		if deps.Keys == nil {
			break
		}

		k := deps.Keys

		switch {
		case k.Global.Quit.Matches(t):
			return m, tea.Quit

		case k.Nav.Up.Matches(t):
			m.viewport.ScrollUp(1)
			return m, nil

		case k.Nav.Down.Matches(t):
			m.viewport.ScrollDown(1)
			return m, nil

		case k.Notice.Continue.Matches(t):
			if m.checked {
				return m, func() tea.Msg {
					return StartupChecksDoneMsg{}
				}
			}
		}
	}

	var viewportCmd, helpCmd tea.Cmd

	m.viewport, viewportCmd = m.viewport.Update(msg)

	if m.help != nil {
		m.help, helpCmd = m.help.Update(msg)
	}

	return m, tea.Batch(viewportCmd, helpCmd)
}
