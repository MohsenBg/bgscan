package header

import (
	"github.com/MohsenBg/bgscan/internal/logger"
	"github.com/MohsenBg/bgscan/internal/ui/shared/env"
	"github.com/MohsenBg/bgscan/internal/ui/shared/layout"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
	"github.com/MohsenBg/bgscan/internal/ui/theme"

	tea "charm.land/bubbletea/v2"
)

type Model struct {
	id     ui.ComponentID
	name   string
	layout *layout.Layout
	theme  *theme.Theme
	log    logger.Set
}

func New(deps ui.Deps) *Model {
	return &Model{layout: deps.Layout, theme: deps.Theme, log: deps.Log, name: "Header", id: ui.NewComponentID()}
}

func (m *Model) ID() ui.ComponentID { return m.id }
func (m *Model) Mode() env.Mode     { return env.NormalMode }
func (m *Model) Name() string       { return m.name }
func (m *Model) Init() tea.Cmd      { return nil }
func (m *Model) OnClose() tea.Cmd   { return nil }

func (m *Model) Theme() *theme.Theme { return m.theme }

func (m *Model) SetTheme(th *theme.Theme) { m.theme = th }
