package crud

import (
	"github.com/MohsenBg/bgscan/internal/logger"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/table"
	"github.com/MohsenBg/bgscan/internal/ui/shared/env"
	"github.com/MohsenBg/bgscan/internal/ui/shared/keys"
	"github.com/MohsenBg/bgscan/internal/ui/shared/layout"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
	"github.com/MohsenBg/bgscan/internal/ui/theme"

	tea "charm.land/bubbletea/v2"
)

type MsgRefresh struct{}

// Model is a generic CRUD controller backed by a sortable table. The
// available commands derive entirely from hooks: zero-value Hooks means a
// read-only view.
type Model[T any] struct {
	id     ui.ComponentID
	layout *layout.Layout
	theme  *theme.Theme
	log    logger.Set
	table  *table.Model
	source Source[T]
	hooks  Hooks[T]

	items    []T
	itemsMap map[string]T
	actions  []actionEntry[T]
}

// New creates a CRUD model; maxWidth bounds the rendered table width.
func New[T any](
	deps ui.Deps,
	source Source[T],
	hooks Hooks[T],
	maxWidth int,
) *Model[T] {
	m := &Model[T]{
		id:       ui.NewComponentID(),
		layout:   deps.Layout,
		theme:    deps.Theme,
		log:      deps.Log,
		source:   source,
		hooks:    hooks,
		itemsMap: make(map[string]T),
		actions:  make([]actionEntry[T], 0, 5),
	}

	// The verbs have to exist before the table is built, because the
	// table's help lines are drawn from them.
	m.bindActions(keys.OrDefault(deps.Keys))

	m.table = table.New(
		deps,
		table.WithTitle(source.Title()),
		table.WithColumns(source.Columns()),
		table.WithRows([]table.Row{}),
		table.WithMaxWidth(maxWidth),
		table.WithKeyShortHints(m.shortHints()...),
		table.WithKeyLongHints(m.longHints()...),
	)

	return m
}

func (m *Model[T]) Init() tea.Cmd { return m.RefreshCmd }

func (m *Model[T]) Theme() *theme.Theme { return m.theme }

func (m *Model[T]) SetTheme(th *theme.Theme) {
	m.theme = th
	if m.table != nil {
		m.table.SetTheme(th)
	}
}

func (m *Model[T]) ID() ui.ComponentID { return m.id }
func (m *Model[T]) Name() string       { return m.source.Title() }
func (m *Model[T]) OnClose() tea.Cmd   { return nil }
func (m *Model[T]) Mode() env.Mode     { return m.table.Mode() }
