// Package workspace composes the root application screen.
package workspace

import (
	"github.com/MohsenBg/bgscan/internal/ui/main/body"
	"github.com/MohsenBg/bgscan/internal/ui/main/footer"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

// model is the root layout component.
type model struct {
	id      ui.ComponentID
	name    string
	session *ui.Session

	body   ui.Component
	footer ui.Component
}

// New builds the workspace root component.
func New(session *ui.Session) ui.Component {
	env := session.Env()
	return &model{
		id:      ui.NewComponentID(),
		name:    "workspace",
		session: session,
		body:    body.New(session),
		footer:  footer.New(env),
	}
}

func (m *model) ID() ui.ComponentID { return m.id }
func (m *model) Name() string       { return m.name }

// Mode delegates to the body, whose topmost open component owns the
// effective key mode.
func (m *model) Mode() ui.KeyMode { return m.body.Mode() }

func (m *model) Init() tea.Cmd {
	return tea.Batch(m.body.Init(), m.footer.Init())
}

// OnClose shuts down the workspace's children.
func (m *model) OnClose() tea.Cmd {
	return tea.Batch(m.body.OnClose(), m.footer.OnClose())
}
