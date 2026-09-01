// Package outboundmenu selects how outbound configs are imported.
package outboundmenu

import (
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/menu"
	"github.com/MohsenBg/bgscan/internal/ui/shared/env"
	"github.com/MohsenBg/bgscan/internal/ui/shared/layout"
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
	id     ui.ComponentID
	name   string
	menu   ui.Component
	Layout *layout.Layout
}

// New creates the outbound import menu.
func New(layout *layout.Layout) *Model {
	m := &Model{
		id:     ui.NewComponentID(),
		name:   "Outbound Menu",
		Layout: layout,
	}
	m.menu = newMenu(layout)
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

func (m *Model) Mode() env.Mode {
	return env.NormalMode
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func newMenu(layout *layout.Layout) *menu.Model {
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
		items, "Addition Method", layout,
		menu.WithHeight(12),
		menu.WithWidth(40),
	)
}
