package logs

import (
	"bgscan/internal/logger"
	"bgscan/internal/ui/components/basic/logview"
	"bgscan/internal/ui/components/basic/menu"
	"bgscan/internal/ui/shared/env"
	"bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

// Model represents the main logs menu component.
type Model struct {
	id    ui.ComponentID
	name  string
	menu  ui.Component
	state *ui.AppState
}

// ID returns the component's unique identifier.
func (m *Model) ID() ui.ComponentID {
	return m.id
}

// Name returns the component's display name.
func (m *Model) Name() string {
	return m.name
}

// OnClose is called when the component is closed.
func (m *Model) OnClose() tea.Cmd {
	return nil
}

// New creates and returns a new logs menu model.
func New(state *ui.AppState) *Model {
	return &Model{
		id:    ui.NewComponentID(),
		name:  "Logs Menu",
		state: state,
		menu:  newLogsMenu(state),
	}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

// newLogsMenu creates the menu with log category options.
func newLogsMenu(state *ui.AppState) *menu.Model {
	items := []menu.MenuItem{
		menu.NewMenuItem(
			"▶", "Core Logs", "c",
			func() tea.Msg {
				return ui.OpenComponentMsg{
					Component: logview.New(state, logger.Core(), "Core Logs"),
				}
			},
		),
		menu.NewMenuItem(
			"⚙", "UI Logs", "u",
			func() tea.Msg {
				return ui.OpenComponentMsg{
					Component: logview.New(state, logger.UI(), "UI Logs"),
				}
			},
		),
		menu.NewMenuItem(
			"::", "Debug Logs", "d",
			func() tea.Msg {
				return ui.OpenComponentMsg{
					Component: logview.New(state, logger.Debug(), "Debug Logs"),
				}
			},
		),
	}
	return menu.New(items, "Logs Menu", state.Layout)
}

func (m *Model) Mode() env.Mode {
	return env.NormalMode
}
