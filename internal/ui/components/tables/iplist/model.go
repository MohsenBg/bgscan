package iplist

import (
	"context"

	tea "charm.land/bubbletea/v2"

	"github.com/MohsenBg/bgscan/internal/core/iplist"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/crud"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/input"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/input/textinput"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/notice"
	"github.com/MohsenBg/bgscan/internal/ui/shared/env"
	"github.com/MohsenBg/bgscan/internal/ui/shared/layout"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
	"github.com/MohsenBg/bgscan/internal/ui/shared/validation"
)

// Model wraps the generic CRUD table for managing IP list files.
type Model struct {
	id        ui.ComponentID
	name      string
	state     *ui.AppState
	layout    *layout.Layout
	crudTable *crud.Model[iplist.IPFileInfo]
}

func New(state *ui.AppState, title string, onSelect func(*iplist.IPFileInfo) tea.Cmd) *Model {
	m := &Model{
		id:     ui.NewComponentID(),
		name:   "IP Files",
		state:  state,
		layout: state.Layout,
	}

	m.crudTable = crud.New(state.Deps(), newSource(state, title), newHooks(state, onSelect), 90)

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

// handleFileSelect opens the name dialog for a file picked from disk.
func (m *Model) handleFileSelect(path string) tea.Cmd {
	if path == "" {
		m.state.Log.UI.Info("[%s]: File selection cancelled", m.name)
		return nil
	}

	inp := textinput.New(m.state.Deps(),
		"What do you want to call this IP file?",
		textinput.WithPlaceholder("filename"),
		textinput.WithValue(""),
		textinput.WithValidation(validation.ValidateFilename),
		textinput.WithFocus(),
	)
	inp.OnSubmit(
		func(filename string) tea.Cmd {
			return tea.Sequence(
				m.saveIPFileCmd(path, filename),
				func() tea.Msg { return crud.MsgRefresh{} },
			)
		},
	)

	return input.OpenInputDialog(inp)
}

// saveIPFileCmd copies the picked file into the IP list directory off the
// UI thread so large imports don't block rendering.
func (m *Model) saveIPFileCmd(srcPath, filename string) tea.Cmd {
	return func() tea.Msg {
		dstPath, err := iplist.GetIPFilePath(filename)
		if err != nil {
			m.state.Log.UI.Error("Failed to resolve destination path: %v", err)
			return notice.NewNoticeCmd(m.state.Deps(), "Copy Failed", err.Error(), notice.NOTICE_ERROR)()
		}

		if err := iplist.ImportIPList(m.state.Log.Core, context.Background(), srcPath, dstPath, iplist.DefaultImportOption()); err != nil {
			m.state.Log.UI.Error("Failed to copy IP file: %v", err)
			return notice.NewNoticeCmd(m.state.Deps(), "Copy Failed", err.Error(), notice.NOTICE_ERROR)()
		}

		m.state.Log.UI.Info("Successfully saved IP file: %s", filename)
		return nil
	}
}
