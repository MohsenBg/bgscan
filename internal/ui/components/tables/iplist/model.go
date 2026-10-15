package iplist

import (
	"context"

	tea "charm.land/bubbletea/v2"

	"github.com/MohsenBg/bgscan/internal/core/iplist"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/crud"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/input"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/input/textinput"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/notice"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/picker"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
	"github.com/MohsenBg/bgscan/internal/ui/shared/validation"
)

// Model wraps the generic CRUD table for managing IP list files.
type Model struct {
	id        ui.ComponentID
	name      string
	session   *ui.Session
	crudTable *crud.Model[iplist.IPFileInfo]
}

func New(session *ui.Session, title string, onSelect func(*iplist.IPFileInfo) tea.Cmd) *Model {
	m := &Model{
		id:      ui.NewComponentID(),
		name:    "IP Files",
		session: session,
	}

	m.crudTable = crud.New(session.Env(), newSource(session, title), newHooks(session, onSelect, m.openFilePicker), 90)

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

func (m *Model) Mode() ui.KeyMode {
	return m.crudTable.Mode()
}

// handleFileSelect opens the name dialog for a file picked from disk.
func (m *Model) handleFileSelect(path string) tea.Cmd {
	if path == "" {
		m.session.Env().Log.Info("[%s]: File selection cancelled", m.name)
		return nil
	}

	inp := textinput.New(m.session.Env(),
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
				func() tea.Msg { return crud.RefreshMsg{} },
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
			m.session.Env().Log.Error("Failed to resolve destination path: %v", err)
			return notice.NewNoticeCmd(m.session.Env(), "Copy Failed", err.Error(), notice.NOTICE_ERROR)()
		}

		coreLog := m.session.CoreLog()
		if err := iplist.ImportIPList(&coreLog, context.Background(), srcPath, dstPath, iplist.DefaultImportOption()); err != nil {
			m.session.Env().Log.Error("Failed to copy IP file: %v", err)
			return notice.NewNoticeCmd(m.session.Env(), "Copy Failed", err.Error(), notice.NOTICE_ERROR)()
		}

		m.session.Env().Log.Info("Successfully saved IP file: %s", filename)
		return nil
	}
}

func (m *Model) openFilePicker() tea.Cmd {
	return picker.OpenFilePickerCmd(
		m.session.Env(),
		"Select IP File (.txt)",
		"",
		[]string{".txt"},
		m.handleFileSelect,
	)

}
