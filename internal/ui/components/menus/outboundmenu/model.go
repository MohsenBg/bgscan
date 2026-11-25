// Package outboundmenu is the menu that asks how to import an outbound
// config, from a sharing link or from a JSON file. It reports the choice
// to its caller with a MsgSelectImportMethod message.
package outboundmenu

import (
	tea "charm.land/bubbletea/v2"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/menu"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
)

// ImportMethod identifies the source format for an outbound import.
type ImportMethod string

const (
	// MethodLink identifies an import from a sharing link.
	MethodLink ImportMethod = "link"
	// MethodJSON identifies an import from a raw JSON file.
	MethodJSON ImportMethod = "json"
)

// MsgSelectImportMethod reports the import method the user selected.
// The Method field holds the selected method.
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

// ID returns the unique identifier of the component.
func (m *Model) ID() ui.ComponentID { return m.id }

// Name returns the display name of the component.
func (m *Model) Name() string { return title }

// OnClose runs when the component closes and returns no command.
func (m *Model) OnClose() tea.Cmd { return nil }

// Mode returns the key mode the component uses while it is active.
func (m *Model) Mode() ui.KeyMode { return ui.KeyModeNormal }

// Init starts the inner menu and returns its first command.
func (m *Model) Init() tea.Cmd { return m.menu.Init() }

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
