package confirm

import (
	"github.com/MohsenBg/bgscan/internal/logger"
	"github.com/MohsenBg/bgscan/internal/ui/shared/layout"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
	"github.com/MohsenBg/bgscan/internal/ui/theme"

	tea "charm.land/bubbletea/v2"
)

// Model represents a confirmation dialog component.
type Model struct {
	id   ui.ComponentID
	name string

	layout *layout.Layout
	theme  *theme.Theme
	log    logger.Set

	message string
	confirm bool

	confirmFunc tea.Cmd
}

func New(
	deps ui.Deps,
	message string,
	onConfirm func() tea.Msg,
	defaultYes bool,
) *Model {
	return &Model{
		id:          ui.NewComponentID(),
		name:        "confirm",
		layout:      deps.Layout,
		theme:       deps.Theme,
		log:         deps.Log,
		message:     message,
		confirm:     defaultYes,
		confirmFunc: onConfirm,
	}
}

func (m *Model) Theme() *theme.Theme { return m.theme }

func (m *Model) SetTheme(th *theme.Theme) { m.theme = th }

func (m *Model) ID() ui.ComponentID {
	return m.id
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Name() string {
	return m.name
}

func (m *Model) OnClose() tea.Cmd {
	return nil
}
