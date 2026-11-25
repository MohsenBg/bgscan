// Package targetlist provides the target list screen for the TUI.
// It wraps the generic CRUD table with a source and hooks backed by the
// target package, so users can list, import, rename, delete, and select
// target list files that hold IP addresses or domains.
package targetlist

import (
	"context"

	tea "charm.land/bubbletea/v2"

	"github.com/MohsenBg/bgscan/internal/core/target"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/crud"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/input"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/input/textinput"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/notice"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/picker"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
	"github.com/MohsenBg/bgscan/internal/ui/shared/validation"
)

// Model wraps the generic CRUD table for managing target list files.
type Model struct {
	id        ui.ComponentID
	name      string
	session   *ui.Session
	crudTable *crud.Model[target.FileInfo]
}

// New creates the target list model with its CRUD table.
// A nil onSelect disables the select action.
func New(session *ui.Session, title string, onSelect func(*target.FileInfo) tea.Cmd) *Model {
	m := &Model{
		id:      ui.NewComponentID(),
		name:    "Target Lists",
		session: session,
	}
	m.crudTable = crud.New(session.Env(), newSource(session, title), newHooks(session, onSelect, m.openFilePicker), 90)
	return m
}

// Init starts the wrapped CRUD table.
func (m *Model) Init() tea.Cmd {
	return m.crudTable.Init()
}

// ID returns the component ID of the model.
func (m *Model) ID() ui.ComponentID {
	return m.id
}

// Name returns the display name of the component.
func (m *Model) Name() string {
	return m.name
}

// OnClose runs the close work of the wrapped CRUD table.
func (m *Model) OnClose() tea.Cmd {
	return m.crudTable.OnClose()
}

// Mode returns the key mode of the wrapped CRUD table.
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
		"What do you want to call this target list?",
		textinput.WithPlaceholder("filename"),
		textinput.WithValue(""),
		textinput.WithValidation(validation.ValidateFilename),
		textinput.WithFocus(),
	)
	inp.OnSubmit(
		func(filename string) tea.Cmd {
			return tea.Sequence(
				m.saveTargetFileCmd(path, filename),
				func() tea.Msg { return crud.RefreshMsg{} },
			)
		},
	)

	return input.OpenInputDialog(inp)
}

// saveTargetFileCmd imports the picked file off the UI thread so large lists
// don't block rendering.
func (m *Model) saveTargetFileCmd(srcPath, filename string) tea.Cmd {
	return func() tea.Msg {
		result, err := target.Import(context.Background(), srcPath, filename, target.DefaultImportOption())
		if err != nil {
			m.session.Env().Log.Error("Failed to import target file: %v", err)
			return notice.NewNoticeCmd(m.session.Env(), "Import Failed", err.Error(), notice.NOTICE_ERROR)()
		}

		m.session.Env().Log.Info(
			"Imported target list %q: %d IPs, %d domains",
			filename, result.IPs, result.Domains,
		)
		return nil
	}
}

func (m *Model) openFilePicker() tea.Cmd {
	return picker.OpenFilePickerCmd(
		m.session.Env(),
		"Select Target File (.txt)",
		"",
		[]string{".txt"},
		m.handleFileSelect,
	)
}
