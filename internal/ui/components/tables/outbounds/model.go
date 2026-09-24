package outbounds

import (
	"github.com/MohsenBg/bgscan/internal/core/xray"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/crud"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/input"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/input/textarea"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/input/textinput"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/notice"
	"github.com/MohsenBg/bgscan/internal/ui/components/menus/outboundmenu"
	"github.com/MohsenBg/bgscan/internal/ui/shared/dialog"
	"github.com/MohsenBg/bgscan/internal/ui/shared/env"
	"github.com/MohsenBg/bgscan/internal/ui/shared/layout"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
	"github.com/MohsenBg/bgscan/internal/ui/shared/validation"

	tea "charm.land/bubbletea/v2"
)

// Model manages the outbound template list: the CRUD table plus the
// multi-step dialogs for importing outbounds from a file or link.
type Model struct {
	id           ui.ComponentID
	name         string
	state        *ui.AppState
	layout       *layout.Layout
	outboundMenu ui.Component
	crudTable    *crud.Model[xray.XrayOutboundsFile]
}

func New(state *ui.AppState, title string, onSelect func(*xray.XrayOutboundsFile) tea.Cmd) *Model {
	m := &Model{
		id:     ui.NewComponentID(),
		name:   "outbounds",
		state:  state,
		layout: state.Layout,
	}

	m.crudTable = crud.New(state.Deps(), newSource(state), newHooks(state, onSelect), 100)

	return m
}

func (m *Model) Init() tea.Cmd      { return m.crudTable.Init() }
func (m *Model) ID() ui.ComponentID { return m.id }
func (m *Model) Name() string       { return m.name }
func (m *Model) OnClose() tea.Cmd   { return m.crudTable.OnClose() }
func (m *Model) Mode() env.Mode     { return m.crudTable.Mode() }

// ShowAdditionMethod opens the dialog for choosing how to add an outbound
// (file or link) instead of the provider's default add hook.
func (m *Model) ShowAdditionMethod() tea.Cmd {
	return func() tea.Msg {
		m.outboundMenu = outboundmenu.New(m.state.Deps())
		return dialog.OpenDialog(m.outboundMenu)
	}
}

// handleFileSelect receives the file picker selection path and asks for a destination name.
func (m *Model) handleFileSelect(path string) tea.Cmd {
	if path == "" {
		m.state.Log.UI.Info("[%s]: File selection cancelled", m.name)
		return nil
	}

	inp := textinput.New(m.state.Deps(),
		"What do you want to call this outbound?",
		textinput.WithPlaceholder("outbound name"),
		textinput.WithValue(""),
		textinput.WithValidation(validation.ValidateFilename),
		textinput.WithFocus(),
		textinput.WithOnSubmit(
			func(filename string) tea.Cmd {
				return tea.Sequence(
					m.saveOutboundFromFileCmd(path, filename),
					func() tea.Msg { return crud.MsgRefresh{} },
				)
			},
		),
	)

	return input.OpenInputDialog(inp)
}

// handleLinkImport starts the sharing link parsing sequence and asks for a destination name.
func (m *Model) handleLinkImport() tea.Cmd {
	linkInput := textarea.New(m.state.Deps(),
		"Enter your outbound link:",
		textarea.WithHeight(15),
		textarea.WithValidation(validation.ValidateXrayLink),
		textarea.WithFocus(),
		textarea.WithValue(""),
	)

	linkInput.OnSubmit(func(link string) tea.Cmd {
		if _, err := xray.ParseLink(link); err != nil {
			return notice.NewNoticeCmd(m.state.Deps(),
				"Parsing Error",
				err.Error(),
				notice.NOTICE_ERROR,
			)
		}

		return m.openFilenameDialog(link)
	})

	return input.OpenInputDialog(linkInput)
}

func (m *Model) openFilenameDialog(link string) tea.Cmd {
	nameInput := textinput.New(m.state.Deps(),
		"What do you want to call this link template?",
		textinput.WithPlaceholder("link profile name"),
		textinput.WithValue(""),
		textinput.WithValidation(validation.ValidateFilename),
		textinput.WithFocus(),
	)

	nameInput.OnSubmit(func(filename string) tea.Cmd {
		return tea.Sequence(
			m.saveOutboundFromLinkCmd(link, filename),
			func() tea.Msg { return crud.MsgRefresh{} },
		)
	})

	return input.OpenInputDialog(nameInput)
}

func (m *Model) saveOutboundFromFileCmd(srcPath, filename string) tea.Cmd {
	meta, err := xray.SaveOutboundFromFile(m.state.Log.Core, srcPath, filename)
	if err != nil {
		m.state.Log.UI.Error("Failed to save outbound from file: %v", err)
		return notice.NewNoticeCmd(m.state.Deps(), "Save Failed", err.Error(), notice.NOTICE_ERROR)
	}
	m.state.Log.UI.Info("Saved outbound file template: %s at path: %s", meta.Name, meta.Path)
	return nil
}

func (m *Model) saveOutboundFromLinkCmd(link, filename string) tea.Cmd {
	meta, err := xray.SaveOutboundFromLink(m.state.Log.Core, link, filename)
	if err != nil {
		m.state.Log.UI.Error("Failed to save outbound from link: %v", err)
		return notice.NewNoticeCmd(m.state.Deps(), "Save Failed", err.Error(), notice.NOTICE_ERROR)
	}
	m.state.Log.UI.Info("Saved outbound link template: %s at path: %s", meta.Name, meta.Path)
	return nil
}

func (m *Model) closeOutboundMenu() tea.Cmd {
	return func() tea.Msg {
		if m.outboundMenu == nil {
			return nil
		}

		id := m.outboundMenu.ID()
		m.outboundMenu = nil
		return ui.CloseComponentMsg{ID: id}
	}
}
