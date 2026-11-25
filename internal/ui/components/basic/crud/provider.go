package crud

import (
	tea "charm.land/bubbletea/v2"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/table"
)

// Source provides data and rendering for a CRUD table.
type Source[T any] interface {
	// Title returns the heading shown above the table.
	Title() string

	// Name returns the singular name of one item, used in prompts.
	Name() string

	// Columns returns the table columns.
	Columns() []table.Column

	// Load returns all items to show in the table.
	Load() ([]T, error)

	// RenderRow converts an item into a table row.
	// The first cell must be the item's identity string.
	RenderRow(item T) table.Row

	// Identity returns a stable unique key for the item.
	Identity(item T) string
}

// Action defines a custom keybinding for the table.
type Action[T any] struct {
	Key  []string // keys that trigger the action
	Name string   // action name, also used by TriggerActionMsg
	Help string   // help text shown in the footer hints

	// NeedsItem makes the model resolve the selected item and pass it
	// to Handler. Without it, Handler gets the zero value.
	NeedsItem bool
	Handler   func(item T) tea.Cmd
}

// Hooks configures optional mutation behaviors.
// Nil fields disable the corresponding built-in actions.
type Hooks[T any] struct {
	OnSelect func(item T) tea.Cmd
	OnDelete func(item T) tea.Cmd // runs after the user confirms
	OnRename func(item T, newName string) tea.Cmd
	OnAdd    func() tea.Cmd
	Custom   []Action[T]
}
