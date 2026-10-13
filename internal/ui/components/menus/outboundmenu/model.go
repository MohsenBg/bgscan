// Package outboundmenu selects how outbound configs are imported.
package outboundmenu

import (
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/menu"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

// ImportMethod defines how the user wants to add outbound configurations.
type ImportMethod string

const (
	MethodLink ImportMethod = "link"
	MethodJSON ImportMethod = "json"
)

// MsgSelectImportMethod is fired when a selection is made, carrying the chosen method.
type MsgSelectImportMethod struct {
	Method ImportMethod
}

// Model is the outbound import method menu.
type Model struct {
	id      ui.ComponentID
	name    string
	menu    ui.Component
	session *ui.Session
}

// New creates the outbound import menu.
func New(session *ui.Session) *Model {
	m := &Model{
		id:      ui.NewComponentID(),
		name:    "Outbound Menu",
		session: session,
	}
	m.menu = newMenu(session)
	return m
}

func (m *Model) ID() ui.ComponentID {
	return m.id
}

func (m *Model) Name() string {
	return m.name
}

func (m *Model) OnClose() tea.Cmd {
	return nil
}

func (m *Model) Mode() ui.KeyMode {
	return ui.KeyModeNormal
}

func (m *Model) Init() tea.Cmd {
	return m.menu.Init()
}

// Title is the dialog-frame title for the outbound import menu.
const Title = "Addition Method"

func newMenu(session *ui.Session) *menu.Model {
	items := []menu.MenuItem{
		menu.NewMenuItem(
			"↗",
			"Via Link",
			"l",
			func() tea.Msg {
				return MsgSelectImportMethod{Method: MethodLink}
			},
		),
		menu.NewMenuItem(
			"{}",
			"Select JSON",
			"j",
			func() tea.Msg {
				return MsgSelectImportMethod{Method: MethodJSON}
			},
		),
	}

	return menu.New(
		items, Title, session.Env(),
		menu.WithHeight(12),
		menu.WithWidth(40),
	)
}
