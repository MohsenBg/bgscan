package crud

import (
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/table"

	tea "charm.land/bubbletea/v2"
)

// Source supplies the read-side data for a CRUD model.
type Source[T any] interface {
	// Title is the display name of the collection.
	Title() string
	// Columns describes the table columns.
	Columns() []table.Column
	// Load returns the current set of items.
	Load() ([]T, error)
	// RenderRow renders one item as a table row.
	//
	// The first cell of the returned Row MUST be the item's Identity,
	// because selection lookup uses it.
	RenderRow(item T) table.Row
	// Identity returns a stable key for item.
	Identity(item T) string
}

// Action is a custom keybinding registered in addition to the built-in
// select / add / delete / rename verbs.
//
// If NeedsItem is true, Handler receives the currently selected item.
// If NeedsItem is false, Handler receives the zero value of T and should
// ignore it.
type Action[T any] struct {
	Key       []string
	Name      string
	Help      string
	NeedsItem bool
	Handler   func(item T) tea.Cmd
}

// Hooks holds the optional mutation behaviors for a CRUD model.
//
// A nil field means the capability is not supported: the keybinding is not
// registered and the action is not dispatched. This replaces the previous
// "call with zero value and check bool" probing hack.
type Hooks[T any] struct {
	// OnSelect is invoked when the user picks an item.
	OnSelect func(item T) tea.Cmd
	// OnDelete is invoked after the user confirms deletion.
	OnDelete func(item T) tea.Cmd
	// OnRename is invoked after the user submits a new name.
	OnRename func(item T, newName string) tea.Cmd
	// OnAdd is invoked when the user triggers "add".
	OnAdd func() tea.Cmd
	// Custom holds extra keybindings registered on the table.
	Custom []Action[T]
}
