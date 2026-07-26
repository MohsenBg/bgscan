package crud

import (
	"bgscan/internal/ui/components/basic/table"
	tea "charm.land/bubbletea/v2"
)

// Provider defines the hook system for the generic CRUD controller.
type Provider[T any] interface {
	Title() string
	Columns() []table.Column
	Load() ([]T, error)
	RenderRow(item T) table.Row
	Identity(item T) string

	// Optional Operations
	OnSelect(item T) (tea.Cmd, bool)
	OnDelete(item T) (tea.Cmd, bool)
	OnRename(item T, newName string) (tea.Cmd, bool)
	OnAdd(item T) (tea.Cmd, bool)
}
