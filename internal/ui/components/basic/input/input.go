// input.go
package input

import (
	tea "charm.land/bubbletea/v2"

	"github.com/MohsenBg/bgscan/internal/ui/shared/keys"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
)

// Dialog defines the interface required to host an input component inside a dialog frame.
// It decouples the dialog manager from the specific input value type.
type Dialog interface {
	ui.Component

	// AppendOnSubmit chains a callback to run after a successful submission.
	// This is used by the dialog manager to close the overlay automatically.
	AppendOnSubmit(fn func() tea.Cmd)

	// DialogHints returns the key hints to display in the dialog frame (excluding Cancel).
	DialogHints() keys.Hints
}

// Input is a generic form field component that holds a value of type T.
// It supports validation, read-only mode, and exposes change/submit callbacks.
type Input[T any] interface {
	Dialog

	// Value returns the current input value.
	Value() T

	// SetValue updates the current input value.
	SetValue(value T)

	// ReadOnly reports whether the input is currently read-only.
	ReadOnly() bool

	// SetReadOnly toggles the read-only state.
	SetReadOnly(bool)

	// OnValidate registers the validation function for the input value.
	OnValidate(fn func(T) error)

	// OnChange registers a callback invoked whenever the value changes.
	OnChange(fn func(T) tea.Cmd)

	// OnSubmit registers a callback invoked when the user submits the value.
	OnSubmit(fn func(T) tea.Cmd)
}
