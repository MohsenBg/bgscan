package crud

import (
	"bgscan/internal/ui/components/basic/table"

	tea "charm.land/bubbletea/v2"
)

// Provider defines the hook system for the generic CRUD controller.
type Provider[T any] interface {
	// Title returns the table title.
	Title() string
	// Columns returns the table column definitions.
	Columns() []table.Column
	// Load returns the current set of items.
	Load() ([]T, error)
	// RenderRow renders a single item as a table row.
	RenderRow(item T) table.Row
	// Identity returns a stable key for item, used for selection and diffing.
	Identity(item T) string

	// OnSelect is called when an item is chosen. The bool reports whether
	// the operation handled the selection (true) or should fall through.
	OnSelect(item T) (tea.Cmd, bool)
	// OnDelete is called when an item is deleted.
	OnDelete(item T) (tea.Cmd, bool)
	// OnRename is called when an item is renamed to newName.
	OnRename(item T, newName string) (tea.Cmd, bool)
	// OnAdd is called when a new item is added.
	OnAdd(item T) (tea.Cmd, bool)
}
