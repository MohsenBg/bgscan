package ipviewer

import (
	"bgscan/internal/core/result"
	"bgscan/internal/ui/components/basic/table"
	"bgscan/internal/ui/shared/env"
	"bgscan/internal/ui/shared/layout"
	"bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

type Model struct {
	id     ui.ComponentID
	name   string
	maxRow uint32
	width  int
	height int
	table  ui.Component
	rows   []table.Row
	schema result.ResultSchema
	layout *layout.Layout
}

type Option func(*Model)

func WithWidth(w int) Option {
	return func(m *Model) { m.width = w }
}

func WithHeight(h int) Option {
	return func(m *Model) { m.height = h }
}

func WithMaxRow(max uint32) Option {
	return func(m *Model) { m.maxRow = max }
}

func (m *Model) ID() ui.ComponentID { return m.id }
func (m *Model) Name() string       { return m.name }
func (m *Model) OnClose() tea.Cmd   { return nil }
func (m *Model) Mode() env.Mode     { return env.NormalMode }
func (m *Model) Init() tea.Cmd      { return nil }

func New(l *layout.Layout, name string, rows []result.Result, schema result.ResultSchema, opts ...Option) *Model {
	cols := make([]table.Column, 0, len(schema.Columns))
	for _, col := range schema.Columns {
		cols = append(cols, table.Column{Title: col.Name, Width: col.Width})
	}

	m := &Model{
		id:     ui.NewComponentID(),
		name:   name,
		maxRow: 10000,
		layout: l,
		schema: schema,
	}

	for _, opt := range opts {
		opt(m)
	}

	tableOpts := []table.Option{
		table.WithColumns(cols),
		table.WithRows([]table.Row{}),
	}
	if m.width > 0 {
		tableOpts = append(tableOpts, table.WithMaxWidth(m.width))
	}
	if m.height > 0 {
		tableOpts = append(tableOpts, table.WithMaxHeight(m.height))
	}

	m.table = table.New(l, tableOpts...)
	m.updateRows(rows)
	return m
}

func (m *Model) SetRows(rows []result.Result) {
	m.updateRows(rows)
}

func (m *Model) Table() *table.Model {
	if t, ok := m.table.(*table.Model); ok {
		return t
	}
	return nil
}

func (m *Model) SetWidth(w int) {
	m.width = w
	if t := m.Table(); t != nil {
		t.SetMaxWidth(w)
	}
}

func (m *Model) SetHeight(h int) {
	m.height = h
	if t := m.Table(); t != nil {
		t.SetMaxHeight(h)
	}
}

func (m *Model) updateRows(rows []result.Result) {
	limit := min(len(rows), int(m.maxRow))
	newRows := make([]table.Row, 0, limit)
	for _, r := range rows[:limit] {
		newRows = append(newRows, r.ToRecord())
	}

	if len(newRows) == len(m.rows) && slicesEqual(newRows, m.rows) {
		return
	}
	m.rows = newRows
	if t := m.Table(); t != nil {
		t.SetRows(newRows)
	}
}

func slicesEqual(a, b []table.Row) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if len(a[i]) != len(b[i]) {
			return false
		}
		for j := range a[i] {
			if a[i][j] != b[i][j] {
				return false
			}
		}
	}
	return true
}
