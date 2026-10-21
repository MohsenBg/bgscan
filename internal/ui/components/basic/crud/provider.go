package crud

import (
	tea "charm.land/bubbletea/v2"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/table"
)

// Source provides data and rendering for a CRUD table.
type Source[T any] interface {
	Title() string
	Name() string
	Columns() []table.Column
	Load() ([]T, error)

	// RenderRow converts an item into a table row.
	// The first cell must be the item's identity string.
	RenderRow(item T) table.Row

	// Identity returns a stable unique key for the item.
	Identity(item T) string
}

// Action defines a custom keybinding for the table.
type Action[T any] struct {
	Key       []string
	Name      string
	Help      string
	NeedsItem bool
	Handler   func(item T) tea.Cmd
}

// Hooks configures optional mutation behaviors.
// Nil fields disable the corresponding built-in actions.
type Hooks[T any] struct {
	OnSelect func(item T) tea.Cmd
	OnDelete func(item T) tea.Cmd
	OnRename func(item T, newName string) tea.Cmd
	OnAdd    func() tea.Cmd
	Custom   []Action[T]
}
