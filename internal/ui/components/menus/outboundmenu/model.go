// Package outboundmenu prompts the user to choose how to import an outbound config.
package outboundmenu

import (
	tea "charm.land/bubbletea/v2"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/menu"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
)

// ImportMethod identifies the source format for an outbound import.
type ImportMethod string

const (
	MethodLink ImportMethod = "link"
	MethodJSON ImportMethod = "json"
)

// MsgSelectImportMethod reports the user's chosen import method.
type MsgSelectImportMethod struct{ Method ImportMethod }

// title is the dialog heading for this menu.
const title = "Addition Method"

// Model wraps the basic menu component for outbound import selection.
type Model struct {
	id   ui.ComponentID
	menu ui.Component
}

// New creates the outbound import menu.
func New(session *ui.Session) *Model {
	return &Model{
		id:   ui.NewComponentID(),
		menu: newMenu(session.Env()),
	}
}

func (m *Model) ID() ui.ComponentID { return m.id }
func (m *Model) Name() string       { return title }
func (m *Model) OnClose() tea.Cmd   { return nil }
func (m *Model) Mode() ui.KeyMode   { return ui.KeyModeNormal }
func (m *Model) Init() tea.Cmd      { return m.menu.Init() }

func newMenu(env *ui.Env) *menu.Model {
	items := []menu.MenuItem{
		menu.NewMenuItem(
			"↗",
			"From Link",
			"l",
			func() tea.Msg { return MsgSelectImportMethod{Method: MethodLink} },
		).WithDescription("Sharing URLs (vmess://, vless://, trojan://, etc.)"),

		menu.NewMenuItem(
			"{}",
			"From File",
			"f",
			func() tea.Msg { return MsgSelectImportMethod{Method: MethodJSON} },
		).WithDescription("Raw Xray JSON configuration file"),
	}

	return menu.New(
		items, title, env,
		menu.WithWidth(55),
		menu.WithHelp(false),
	)
}
