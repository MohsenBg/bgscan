package input

import (
	"bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

// Dialog is the subset of Input needed to host a component as a dialog:
// identity plus the ability to chain a callback after submit succeeds,
// without depending on the input's value type.
type Dialog interface {
	ui.Component
	AppendOnSubmit(fn func() tea.Cmd)
}

// Input is a generic, focusable form field component that holds a value of
// type T, supports validation, and exposes change/submit callbacks.
type Input[T any] interface {
	Dialog

	// Value returns the current value of the input.
	Value() T
	// SetValue sets the current value of the input.
	SetValue(value T)

	// ReadOnly reports whether the input is currently read-only.
	ReadOnly() bool
	// SetReadOnly sets whether the input is read-only.
	SetReadOnly(bool)

	// OnValidate registers the function used to validate the input's value.
	OnValidate(fn func(T) error)

	// OnChange registers a callback invoked whenever the value changes.
	OnChange(fn func(T) tea.Cmd)
	// OnSubmit registers a callback invoked when the value is submitted.
	OnSubmit(fn func(T) tea.Cmd)
}
