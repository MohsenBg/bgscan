package outbounds

import (
	"bgscan/internal/core/xray"
	"bgscan/internal/logger"
	"bgscan/internal/ui/components/basic/crud"
	"bgscan/internal/ui/components/basic/input"
	"bgscan/internal/ui/components/basic/notice"
	"bgscan/internal/ui/shared/env"
	"bgscan/internal/ui/shared/layout"
	"bgscan/internal/ui/shared/ui"
	"bgscan/internal/ui/shared/validation"

	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	id        ui.ComponentID
	name      string
	layout    *layout.Layout
	crudTable *crud.Model[xray.XrayOutboundsFile]
}

// New creates a new outbound template list component.
func New(l *layout.Layout, title string, onSelect func(*xray.XrayOutboundsFile) tea.Cmd) *Model {
	m := &Model{
		id:     ui.NewComponentID(),
		name:   "outbounds",
		layout: l,
	}

	canAdd := true
	m.crudTable = crud.New("outbound", l, newProvider(l, onSelect), canAdd)

	return m
}

func (m *Model) Init() tea.Cmd      { return m.crudTable.Init() }
func (m *Model) ID() ui.ComponentID { return m.id }
func (m *Model) Name() string       { return m.name }
func (m *Model) OnClose() tea.Cmd   { return m.crudTable.OnClose() }
func (m *Model) Mode() env.Mode     { return m.crudTable.Mode() }

// Update the addition trigger to fire the prompt steps elegantly
func (m *Model) handleFileSelect(path string) tea.Cmd {
	if path == "" {
		logger.UIInfo("[%s]: File selection cancelled", m.name)
		return nil
	}

	return input.ShowInputCmd(
		m.layout,
		"What do you want to call this outbound?",
		"outbound name",
		"",
		validation.ValidateFilename,
		nil,
		func(filename string) tea.Cmd {
			return tea.Sequence(
				m.saveOutboundCmd(path, filename),
				func() tea.Msg { return crud.MsgRefresh{} },
			)
		},
	)
}

func (m *Model) saveOutboundCmd(srcPath, filename string) tea.Cmd {
	meta, err := xray.SaveOutbound(srcPath, filename)
	if err != nil {
		logger.UIError("Failed to save outbound: %v", err)
		return notice.NewNoticeCmd(m.layout, "Save Failed", err.Error(), notice.NOTICE_ERROR)
	}
	logger.UIInfo("save outbound:%s path:%s", meta.Name, meta.Path)
	return nil
}
