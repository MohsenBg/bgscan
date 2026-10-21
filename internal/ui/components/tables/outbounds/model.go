package outbounds

import (
	tea "charm.land/bubbletea/v2"
	"github.com/MohsenBg/bgscan/internal/core/xray"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/crud"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/input"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/input/textarea"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/input/textinput"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/notice"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
	"github.com/MohsenBg/bgscan/internal/ui/shared/validation"
)

// Model manages the outbound template list: the CRUD table plus the
// multi-step dialogs for importing outbounds from a file or link.
type Model struct {
	id           ui.ComponentID
	session      *ui.Session
	outboundMenu ui.Component
	crudTable    *crud.Model[xray.OutboundFile]
}

// New creates the outbounds management component.
func New(session *ui.Session, onSelect func(*xray.OutboundFile) tea.Cmd) *Model {
	m := &Model{
		id:      ui.NewComponentID(),
		session: session,
	}
	m.crudTable = crud.New(
		session.Env(),
		newSource(session),
		newHooks(session, onSelect, m.showAdditionMethodCmd),
		100,
	)
	return m
}

func (m *Model) Init() tea.Cmd      { return m.crudTable.Init() }
func (m *Model) ID() ui.ComponentID { return m.id }
func (m *Model) Name() string       { return "outbounds" }
func (m *Model) OnClose() tea.Cmd   { return m.crudTable.OnClose() }
func (m *Model) Mode() ui.KeyMode   { return m.crudTable.Mode() }

// showAdditionMethodCmd returns a command that triggers the add menu in Update.
// This avoids mutating component state from inside a tea.Cmd.
func (m *Model) showAdditionMethodCmd() tea.Cmd {
	return func() tea.Msg { return openAddMenuMsg{} }
}

// closeMenuCmd safely clears the outbound menu reference and returns a command
// to close its dialog. Must be called from Update.
func (m *Model) closeMenuCmd() tea.Cmd {
	if m.outboundMenu == nil {
		return nil
	}
	id := m.outboundMenu.ID()
	m.outboundMenu = nil
	return func() tea.Msg { return ui.CloseComponentMsg{ID: id} }
}

// handleFileSelect receives the file picker selection path and asks for a destination name.
func (m *Model) handleFileSelect(path string) tea.Cmd {
	if path == "" {
		return nil // cancelled
	}

	inp := textinput.New(m.session.Env(),
		"What do you want to call this outbound?",
		textinput.WithPlaceholder("outbound name"),
		textinput.WithValidation(validation.ValidateFilename),
		textinput.WithFocus(),
		textinput.WithOnSubmit(func(filename string) tea.Cmd {
			return tea.Batch(
				m.saveOutboundFromFileCmd(path, filename),
				func() tea.Msg { return crud.RefreshMsg{} },
			)
		}),
	)
	return input.OpenInputDialog(inp)
}

// handleLinkImport starts the sharing link parsing sequence and asks for a destination name.
func (m *Model) handleLinkImport() tea.Cmd {
	linkInput := textarea.New(m.session.Env(),
		"Enter your outbound link:",
		textarea.WithHeight(15),
		textarea.WithValidation(validation.ValidateXrayLink),
		textarea.WithFocus(),
	)
	linkInput.OnSubmit(func(link string) tea.Cmd {
		if _, err := xray.ParseLink(link); err != nil {
			return notice.NewNoticeCmd(m.session.Env(), "Parsing Error", err.Error(), notice.NOTICE_ERROR)
		}
		return m.openFilenameDialog(link)
	})
	return input.OpenInputDialog(linkInput)
}

func (m *Model) openFilenameDialog(link string) tea.Cmd {
	nameInput := textinput.New(m.session.Env(),
		"What do you want to call this link template?",
		textinput.WithPlaceholder("link profile name"),
		textinput.WithValidation(validation.ValidateFilename),
		textinput.WithFocus(),
	)
	nameInput.OnSubmit(func(filename string) tea.Cmd {
		return tea.Batch(
			m.saveOutboundFromLinkCmd(link, filename),
			func() tea.Msg { return crud.RefreshMsg{} },
		)
	})
	return input.OpenInputDialog(nameInput)
}

func (m *Model) saveOutboundFromFileCmd(srcPath, filename string) tea.Cmd {
	return func() tea.Msg {
		coreLog := m.session.CoreLog()
		meta, err := xray.SaveOutboundFromFile(&coreLog, srcPath, filename)
		if err != nil {
			m.session.Env().Log.Error("Failed to save outbound from file: %v", err)
			cmd := notice.NewNoticeCmd(m.session.Env(), "Save Failed", err.Error(), notice.NOTICE_ERROR)
			if cmd != nil {
				return cmd()
			}
			return nil
		}
		m.session.Env().Log.Info("Saved outbound file template: %s at path: %s", meta.Name, meta.Path)
		return nil
	}
}

func (m *Model) saveOutboundFromLinkCmd(link, filename string) tea.Cmd {
	return func() tea.Msg {
		coreLog := m.session.CoreLog()
		meta, err := xray.SaveOutboundFromLink(&coreLog, link, filename)
		if err != nil {
			m.session.Env().Log.Error("Failed to save outbound from link: %v", err)
			cmd := notice.NewNoticeCmd(m.session.Env(), "Save Failed", err.Error(), notice.NOTICE_ERROR)
			if cmd != nil {
				return cmd()
			}
			return nil
		}
		m.session.Env().Log.Info("Saved outbound link template: %s at path: %s", meta.Name, meta.Path)
		return nil
	}
}
