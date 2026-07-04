package crud

import (
	"bgscan/internal/ui/components/basic/table"
	"bgscan/internal/ui/shared/env"
	"bgscan/internal/ui/shared/layout"
	"bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

type (
	AddFunc[T any] func() (T, error)
	MsgRefresh     struct{}
)

type Model[T any] struct {
	id       ui.ComponentID
	name     string
	layout   *layout.Layout
	table    *table.Model
	provider Provider[T]
	canAdd   bool

	items    []T
	itemsMap map[string]T
}

func New[T any](
	name string,
	l *layout.Layout,
	provider Provider[T],
	maxWidth int,
	canAdd bool,
) *Model[T] {
	m := &Model[T]{
		id:       ui.NewComponentID(),
		name:     name,
		layout:   l,
		provider: provider,
		canAdd:   canAdd,
		itemsMap: make(map[string]T),
	}

	m.table = table.New(
		l,
		table.WithTitle(provider.Title()),
		table.WithColumns(provider.Columns()),
		table.WithRows([]table.Row{}),
		table.WithMaxWidth(maxWidth),
	)

	m.configureKeymaps()

	return m
}

func (m *Model[T]) Init() tea.Cmd {
	return m.RefreshCmd
}

func (m *Model[T]) ID() ui.ComponentID { return m.id }
func (m *Model[T]) Name() string       { return m.name }
func (m *Model[T]) OnClose() tea.Cmd   { return nil }
func (m *Model[T]) Mode() env.Mode     { return env.NormalMode }

// CanAdd returns whether the Add operation is available
func (m *Model[T]) CanAdd() bool {
	return m.canAdd
}
