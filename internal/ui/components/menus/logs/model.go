// Package logs is the menu that lets the user open a log viewer for
// one log source. It offers the core log and the UI log, and each item
// opens a logview component for that source.
package logs

import (
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/logview"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/menu"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

// Model is the logs menu component.
type Model struct {
	id      ui.ComponentID
	name    string
	menu    ui.Component
	session *ui.Session
}

// ID returns the component's unique identifier.
func (m *Model) ID() ui.ComponentID {
	return m.id
}

// Name returns the component's display name.
func (m *Model) Name() string {
	return m.name
}

// OnClose runs when the component closes and returns no command.
func (m *Model) OnClose() tea.Cmd {
	return nil
}

// New creates the logs menu for the given session.
func New(session *ui.Session) *Model {
	return &Model{
		id:      ui.NewComponentID(),
		name:    "Logs Menu",
		session: session,
		menu:    newLogsMenu(session),
	}
}

// Init starts the component and returns no command.
func (m *Model) Init() tea.Cmd {
	return nil
}

// newLogsMenu creates the menu with log category options.
func newLogsMenu(session *ui.Session) *menu.Model {
	env := session.Env()
	items := []menu.MenuItem{
		menu.NewMenuItem(
			"▶", "Core Logs", "c",
			func() tea.Msg {
				coreLog := session.CoreLog()
				return ui.OpenComponentMsg{
					Component: logview.New(env, &coreLog, "Core Logs"),
				}
			},
		).WithDescription("Scanner engine, probes and results activity"),
		menu.NewMenuItem(
			"⚙", "UI Logs", "u",
			func() tea.Msg {
				uiLog := env.Log
				return ui.OpenComponentMsg{
					Component: logview.New(env, &uiLog, "UI Logs"),
				}
			},
		).WithDescription("UI runtime and component events"),
	}
	return menu.New(items, "Logs Menu", env)
}

// Mode returns the key mode the component uses while it is active.
func (m *Model) Mode() ui.KeyMode {
	return ui.KeyModeNormal
}
