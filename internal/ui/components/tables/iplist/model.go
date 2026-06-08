package iplist

import (
	tea "github.com/charmbracelet/bubbletea"

	"bgscan/internal/core/iplist"
	"bgscan/internal/logger"
	"bgscan/internal/ui/components/basic/crud"
	"bgscan/internal/ui/components/basic/input"
	"bgscan/internal/ui/components/basic/notice"
	"bgscan/internal/ui/shared/env"
	"bgscan/internal/ui/shared/layout"
	"bgscan/internal/ui/shared/ui"
	"bgscan/internal/ui/shared/validation"
)

// Model wraps a generic CRUD administration table layer tailored specifically
// to monitoring, adding, and selecting source IP target lists.
type Model struct {
	id        ui.ComponentID
	name      string
	layout    *layout.Layout
	crudTable *crud.Model[iplist.IPFileInfo]
}

// New instantiates a reactive UI node orchestrating underlying IP asset tracking configurations.
func New(l *layout.Layout, title string, onSelect func(*iplist.IPFileInfo) tea.Cmd) *Model {
	m := &Model{
		id:     ui.NewComponentID(),
		name:   "iplist",
		layout: l,
	}

	canAdd := true
	m.crudTable = crud.New("ip file", l, newProvider(l, onSelect), canAdd)

	return m
}

func (m *Model) Init() tea.Cmd {
	return m.crudTable.Init()
}

func (m *Model) ID() ui.ComponentID {
	return m.id
}

func (m *Model) Name() string {
	return m.name
}

func (m *Model) OnClose() tea.Cmd {
	return m.crudTable.OnClose()
}

func (m *Model) Mode() env.Mode {
	return m.crudTable.Mode()
}

// handleFileSelect manages context transitions when mounting external system target sources.
func (m *Model) handleFileSelect(path string) tea.Cmd {
	if path == "" {
		logger.UIInfo("[%s]: File selection cancelled", m.name)
		return nil
	}

	return input.ShowInputCmd(
		m.layout,
		"What do you want to call this IP file?",
		"filename",
		"",
		validation.ValidateFilename,
		nil,
		func(filename string) tea.Cmd {
			return tea.Sequence(
				m.saveIPFileCmd(path, filename),
				func() tea.Msg { return crud.MsgRefresh{} },
			)
		},
	)
}

// saveIPFileCmd contractually isolates filesystem disk operations away from the main loop thread.
func (m *Model) saveIPFileCmd(srcPath, filename string) tea.Cmd {
	return func() tea.Msg {
		dstPath, err := iplist.GetIPFilePath(filename)
		if err != nil {
			logger.UIError("Failed to resolve destination path: %v", err)
			return notice.NewNoticeCmd(m.layout, "Copy Failed", err.Error(), notice.NOTICE_ERROR)()
		}

		if err := iplist.CopyIPFile(srcPath, dstPath); err != nil {
			logger.UIError("Failed to copy IP file: %v", err)
			return notice.NewNoticeCmd(m.layout, "Copy Failed", err.Error(), notice.NOTICE_ERROR)()
		}

		logger.UIInfo("Successfully saved IP file: %s", filename)
		return nil
	}
}

