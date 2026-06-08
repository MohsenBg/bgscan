package ipviewer

import (
	"bgscan/internal/core/result"
	"bgscan/internal/ui/components/basic/table"
	"bgscan/internal/ui/shared/env"
	"bgscan/internal/ui/shared/layout"
	"bgscan/internal/ui/shared/ui"

	tea "github.com/charmbracelet/bubbletea"
)

var ipFullListColumns = []table.Column{
	{Title: "IP", Width: 20},
	{Title: "Latency", Width: 15},
	{Title: "Download", Width: 15},
	{Title: "Upload", Width: 15},
}

var ipShortListColumns = []table.Column{
	{Title: "IP", Width: 80},
	{Title: "Latency", Width: 20},
}

type ViewMode string

const (
	ShortView ViewMode = "ShortView"
	FullView  ViewMode = "FullView"
)

// Model represents the main logs menu component.
type Model struct {
	id       ui.ComponentID
	viewMode ViewMode
	name     string
	maxRow   uint32
	table    ui.Component
	rows     []table.Row
	layout   *layout.Layout
}

// ID returns the component's unique identifier.
func (m *Model) ID() ui.ComponentID {
	return m.id
}

// Name returns the component's display name.
func (m *Model) Name() string {
	return m.name
}

// OnClose is called when the component is closed.
func (m *Model) OnClose() tea.Cmd {
	return nil
}

// New creates and returns a new logs menu model.
func New(l *layout.Layout, name string, rows []result.IPScanResult, mode ViewMode) *Model {
	cols := ipFullListColumns
	if mode == ShortView {
		cols = ipShortListColumns
	}
	t := table.New(name, cols, []table.Row{}, l)

	m := &Model{
		id:       ui.NewComponentID(),
		name:     name,
		maxRow:   10_000,
		viewMode: mode,
		layout:   l,
		table:    t,
	}

	m.updateRows(rows)
	return m
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) SetRows(rows []result.IPScanResult) {
	m.updateRows(rows)
}
func (m *Model) Table() *table.Model {
	if t, ok := m.table.(*table.Model); ok {
		return t
	}
	return nil
}

func (m *Model) Mode() env.Mode {
	return env.NormalMode
}
