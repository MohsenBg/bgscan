package crud

import (
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/table"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

type MsgRefresh struct{}

// Model is a generic CRUD controller backed by a sortable table. The
// available commands derive entirely from hooks: zero-value Hooks means a
// read-only view.
type Model[T any] struct {
	id     ui.ComponentID
	env    *ui.Env
	table  *table.Model
	source Source[T]
	hooks  Hooks[T]

	items    []T
	itemsMap map[string]T
	actions  []actionEntry[T]
}

// New creates a CRUD model; maxWidth bounds the rendered table width.
func New[T any](
	env *ui.Env,
	source Source[T],
	hooks Hooks[T],
	maxWidth int,
) *Model[T] {
	m := &Model[T]{
		id:       ui.NewComponentID(),
		env:      env,
		source:   source,
		hooks:    hooks,
		itemsMap: make(map[string]T),
		actions:  make([]actionEntry[T], 0, 5),
	}

	// The verbs have to exist before the table is built, because the
	// table's help lines are drawn from them.
	m.bindActions(*env.Keys)

	m.table = table.New(
		env,
		table.WithTitle(source.Title()),
		table.WithColumns(source.Columns()),
		table.WithRows([]table.Row{}),
		table.WithMaxWidth(maxWidth),
		table.WithKeyShortHints(m.shortHints()...),
		table.WithKeyLongHints(m.longHints()...),
	)

	return m
}

func (m *Model[T]) Init() tea.Cmd { return tea.Batch(m.RefreshCmd, m.table.Init()) }

func (m *Model[T]) ID() ui.ComponentID { return m.id }
func (m *Model[T]) Name() string       { return m.source.Title() }
func (m *Model[T]) OnClose() tea.Cmd   { return nil }
func (m *Model[T]) Mode() ui.KeyMode   { return m.table.Mode() }
