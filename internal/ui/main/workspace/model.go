package workspace

import (
	"bgscan/internal/ui/main/body"
	"bgscan/internal/ui/main/footer"
	"bgscan/internal/ui/main/header"
	"bgscan/internal/ui/shared/dialog"
	"bgscan/internal/ui/shared/env"
	"bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

type dialogPosition struct {
	XPos    dialog.DialogPosition
	YPos    dialog.DialogPosition
	XOffset int
	YOffset int
}

type model struct {
	id               ui.ComponentID
	name             string
	state            *ui.AppState
	dialog           []ui.Component
	dialogPlacements map[ui.ComponentID]*dialogPosition
	header           ui.Component
	body             ui.Component
	footer           ui.Component
}

func (m *model) ID() ui.ComponentID { return m.id }
func (m *model) Mode() env.Mode     { return env.NormalMode }
func (m *model) Name() string       { return m.name }
func (m *model) OnClose() tea.Cmd {
	return tea.Batch(m.header.OnClose(), m.body.OnClose(), m.footer.OnClose())
}

func New(state *ui.AppState) ui.Component {
	return &model{
		id:               ui.NewComponentID(),
		name:             "workspace",
		state:            state,
		dialog:           make([]ui.Component, 0, 5),
		dialogPlacements: make(map[ui.ComponentID]*dialogPosition),
		header:           header.New(state.Layout),
		body:             body.New(state),
		footer:           footer.New(state.Layout),
	}
}

func (m *model) Init() tea.Cmd {
	return tea.Batch(m.header.Init(), m.body.Init(), m.footer.Init())
}

func (m *model) getDialogPlacement(id ui.ComponentID) *dialogPosition {
	if p, ok := m.dialogPlacements[id]; ok {
		return p
	}
	p := &dialogPosition{XPos: dialog.Center, YPos: dialog.Center}
	m.dialogPlacements[id] = p
	return p
}
