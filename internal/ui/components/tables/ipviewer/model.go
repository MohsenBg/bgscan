// Package ipviewer shows scan results in a table built from a result
// schema. It colors metric cells by tier, caps the rows shown at maxRow,
// and copies the selected row's host value to the clipboard when the
// user presses enter.
package ipviewer

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/MohsenBg/bgscan/internal/core/result"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/table"
	"github.com/MohsenBg/bgscan/internal/ui/shared/keys"
	"github.com/MohsenBg/bgscan/internal/ui/shared/meter"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

// Model is the result table component. It wraps a basic table and keeps
// the rendered rows so unchanged results skip a table update.
type Model struct {
	id     ui.ComponentID
	name   string
	maxRow uint32
	width  int
	height int
	table  ui.Component
	rows   []table.Row
	schema result.ResultSchema
	env    *ui.Env

	shortHints []keys.Hint
	longHints  []keys.Hint

	filterEnabled bool
}

// Option changes a Model at creation time. Options are passed to New.
type Option func(*Model)

// WithWidth sets the table width. A zero width leaves the table at its
// default width.
func WithWidth(w int) Option {
	return func(m *Model) { m.width = w }
}

// WithHeight sets the table height. A zero height leaves the table at
// its default height.
func WithHeight(h int) Option {
	return func(m *Model) { m.height = h }
}

// WithMaxRow sets the maximum number of rows shown. A zero max is
// ignored, so the current limit stays in place.
func WithMaxRow(max uint32) Option {
	return func(m *Model) {
		if max > 0 {
			m.maxRow = max
		}
	}
}

// WithShortHints adds entries to the table's short help line.
func WithShortHints(hints ...keys.Hint) Option {
	return func(m *Model) { m.shortHints = hints }
}

// WithLongHints adds entries to the table's full help dialog.
func WithLongHints(hints ...keys.Hint) Option {
	return func(m *Model) { m.longHints = hints }
}

// WithFiltering sets whether the user can filter the table rows.
// Filtering stays off unless WithFiltering(true) is passed.
func WithFiltering(enabled bool) Option {
	return func(m *Model) { m.filterEnabled = enabled }
}

// ID returns the component ID assigned at creation.
func (m *Model) ID() ui.ComponentID { return m.id }

// Name returns the component display name.
func (m *Model) Name() string { return m.name }

// OnClose runs when the component closes. It has nothing to clean up and
// returns nil.
func (m *Model) OnClose() tea.Cmd { return nil }

// Mode returns the key mode of the underlying table.
func (m *Model) Mode() ui.KeyMode { return m.table.Mode() }

// Init starts the underlying table and returns its first command.
func (m *Model) Init() tea.Cmd { return m.table.Init() }

// New creates a result table with columns taken from schema and fills it
// with rows. The row cap comes from the environment settings, or the
// default cap when the setting is zero, unless WithMaxRow overrides it.
func New(env *ui.Env, name string, rows []result.Result, schema result.ResultSchema, opts ...Option) *Model {
	cols := make([]table.Column, 0, len(schema.Columns))
	for _, col := range schema.Columns {
		cols = append(cols, table.Column{Title: col.Name, Width: col.Width})
	}

	maxRows := env.Settings.MaxRow
	if maxRows == 0 {
		maxRows = ui.DefaultMaxRow
	}

	m := &Model{
		id:     ui.NewComponentID(),
		name:   name,
		maxRow: maxRows,
		env:    env,
		schema: schema,
	}

	for _, opt := range opts {
		opt(m)
	}

	tableOpts := []table.Option{
		table.WithColumns(cols),
		table.WithRows([]table.Row{}),
	}
	if len(cols) > 0 {
		types := make([]result.ColumnType, len(cols))
		for i, c := range schema.Columns {
			types[i] = c.ColumnType
		}
		tableOpts = append(tableOpts, table.WithColumnTypes(types...))
	}
	if m.width > 0 {
		tableOpts = append(tableOpts, table.WithMaxWidth(m.width))
	}
	if m.height > 0 {
		tableOpts = append(tableOpts, table.WithMaxHeight(m.height))
	}
	if m.name != "" {
		tableOpts = append(tableOpts, table.WithTitle(m.name))
	}
	if len(m.shortHints) > 0 {
		tableOpts = append(tableOpts, table.WithKeyShortHints(m.shortHints...))
	}
	if len(m.longHints) > 0 {
		tableOpts = append(tableOpts, table.WithKeyLongHints(m.longHints...))
	}
	if !m.filterEnabled {
		tableOpts = append(tableOpts, table.WithFiltering(false))
	}

	m.table = table.New(env, tableOpts...)
	m.updateRows(rows)
	return m
}

// SetRows replaces the results shown in the table. Rows past the row
// cap are dropped, and a call with rows equal to the current ones does
// nothing.
func (m *Model) SetRows(rows []result.Result) {
	m.updateRows(rows)
}

// Table returns the underlying table model, or nil when the wrapped
// component is not a table model.
func (m *Model) Table() *table.Model {
	if t, ok := m.table.(*table.Model); ok {
		return t
	}
	return nil
}

// SetWidth sets the table width and forwards it to the underlying table.
func (m *Model) SetWidth(w int) {
	m.width = w
	if t := m.Table(); t != nil {
		t.SetWidth(w)
	}
}

// SetHeight sets the table height and forwards it to the underlying
// table.
func (m *Model) SetHeight(h int) {
	m.height = h
	if t := m.Table(); t != nil {
		t.SetHeight(h)
	}
}

func (m *Model) updateRows(rows []result.Result) {
	limit := min(len(rows), int(m.maxRow))
	newRows := make([]table.Row, 0, limit)
	for _, r := range rows[:limit] {
		newRows = append(newRows, m.colorRow(r.ToRecord()))
	}

	if len(newRows) == len(m.rows) && slicesEqual(newRows, m.rows) {
		return
	}
	m.rows = newRows
	if t := m.Table(); t != nil {
		t.SetRows(newRows)
	}
}

// colorRow rewrites metric cells (duration, bandwidth) in place with their
// tier color. Cells with no rule or no measurement are left untouched.
func (m *Model) colorRow(rec []string) table.Row {
	for i, cell := range rec {
		if i >= len(m.schema.Columns) {
			break
		}

		c := m.env.Theme.TierColor(
			meter.Resolve(m.schema.Name, m.schema.Columns[i].ColumnType, cell),
		)
		if c == nil {
			continue
		}

		styled := lipgloss.NewStyle().Foreground(c).Render(cell)
		rec[i] = keepRowBackground(styled)
	}

	return rec
}

// keepRowBackground swaps the styled value's trailing full reset for SGR 39
// (default foreground). The table wraps the whole row in the selected style
// and emits its background once, so a mid-row full reset would turn the
// highlight off for every cell after the colored one. SGR 39 ends only the
// value's own color and leaves the wrapper's background active.
func keepRowBackground(styled string) string {
	styled = strings.TrimSuffix(styled, ansi.ResetStyle)
	styled = strings.TrimSuffix(styled, "\x1b[0m")

	return styled + "\x1b[39m"
}

// ipColumn returns the schema index of the results' key column.
// Falls back to 0; every schema declares that column first.
func (m *Model) ipColumn() int {
	for i, c := range m.schema.Columns {
		if c.ColumnType == result.ColumnHost {
			return i
		}
	}

	return 0
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
