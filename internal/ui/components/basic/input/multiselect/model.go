package multiselect

import (
	"errors"

	"github.com/MohsenBg/bgscan/internal/ui/components/basic/input"
	"github.com/MohsenBg/bgscan/internal/ui/shared/env"
	"github.com/MohsenBg/bgscan/internal/ui/shared/layout"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
	"github.com/MohsenBg/bgscan/internal/ui/theme"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
)

// Option configures a Model at construction time.
type Option[T comparable] func(*Model[T])

// Model is the multi-select implementation of [input.Input], where the
// value is a slice of selected options.
type Model[T comparable] struct {
	id   ui.ComponentID
	name string

	layout *layout.Layout

	title    string
	errorMsg string

	value    []T
	options  []huh.Option[T]
	limit    int
	huhInput *huh.MultiSelect[T]
	readOnly bool

	validationFunc func(value []T) error

	onChange func([]T) tea.Cmd
	onSubmit func([]T) tea.Cmd

	required bool
}

// New creates a new multi-select input component.
func New[T comparable](
	l *layout.Layout,
	title string,
	options ...Option[T],
) input.Input[[]T] {
	m := &Model[T]{
		id:     ui.NewComponentID(),
		name:   "multiselect",
		layout: l,
		title:  title,
	}

	m.huhInput = huh.NewMultiSelect[T]().
		Title(title).
		Value(&m.value)

	for _, opt := range options {
		opt(m)
	}

	inp := m.huhInput.Options(m.options...).WithWidth(m.Width())
	m.huhInput = inp.(*huh.MultiSelect[T])
	m.huhInput.WithKeyMap(huh.NewDefaultKeyMap())
	m.huhInput.WithTheme(theme.NewHuhTheme())

	if m.limit > 0 {
		m.huhInput = m.huhInput.Limit(m.limit)
	}

	return m
}

// WithOptions sets the selectable options.
func WithOptions[T comparable](opts ...huh.Option[T]) Option[T] {
	return func(m *Model[T]) {
		m.options = opts
	}
}

// WithValue sets the initially selected values.
func WithValue[T comparable](value []T) Option[T] {
	return func(m *Model[T]) {
		m.value = value
	}
}

// WithLimit sets the maximum number of selectable options. 0 means no limit.
func WithLimit[T comparable](limit int) Option[T] {
	return func(m *Model[T]) {
		m.limit = limit
	}
}

// WithValidation sets the function used to validate the input's value.
func WithValidation[T comparable](fn func([]T) error) Option[T] {
	return func(m *Model[T]) {
		m.validationFunc = fn
	}
}

// WithFocus focuses the input on creation.
func WithFocus[T comparable]() Option[T] {
	return func(m *Model[T]) {
		m.huhInput.Focus()
	}
}

// WithReadOnly sets the initial read-only state of the input.
func WithReadOnly[T comparable](ro bool) Option[T] {
	return func(m *Model[T]) {
		m.setReadOnly(ro)
	}
}

// WithOnChange registers a callback invoked whenever the value changes.
func WithOnChange[T comparable](fn func([]T) tea.Cmd) Option[T] {
	return func(m *Model[T]) {
		m.onChange = fn
	}
}

// WithOnSubmit registers a callback invoked when the value is submitted.
func WithOnSubmit[T comparable](fn func([]T) tea.Cmd) Option[T] {
	return func(m *Model[T]) {
		m.onSubmit = fn
	}
}

func (m *Model[T]) Init() tea.Cmd { return nil }

func (m *Model[T]) ID() ui.ComponentID { return m.id }

func (m *Model[T]) Name() string { return m.name }

func (m *Model[T]) Mode() env.Mode { return env.NormalMode }

func (m *Model[T]) Width() int {
	if m.layout == nil {
		return 50
	}
	return min(50, m.layout.Body.Width)
}

func (m *Model[T]) CloseCmd() tea.Cmd {
	return func() tea.Msg {
		return ui.CloseComponentMsg{ID: m.ID()}
	}
}

func (m *Model[T]) OnClose() tea.Cmd { return nil }

func (m *Model[T]) Value() []T { return m.value }

func (m *Model[T]) SetValue(value []T) {
	m.value = value
	m.huhInput = m.huhInput.Value(&m.value)
}

func (m *Model[T]) ReadOnly() bool { return m.readOnly }

func (m *Model[T]) SetReadOnly(ro bool) { m.setReadOnly(ro) }

func (m *Model[T]) OnValidate(fn func([]T) error) { m.validationFunc = fn }

func (m *Model[T]) OnChange(fn func([]T) tea.Cmd) { m.onChange = fn }

func (m *Model[T]) OnSubmit(fn func([]T) tea.Cmd) { m.onSubmit = fn }

// AppendOnSubmit implements [input.Input]. It chains fn after any
// previously registered onSubmit callback rather than replacing it.
func (m *Model[T]) AppendOnSubmit(fn func() tea.Cmd) {
	prev := m.onSubmit
	m.onSubmit = func(value []T) tea.Cmd {
		if prev == nil {
			return fn()
		}
		return tea.Sequence(prev(value), fn())
	}
}

func (m *Model[T]) setReadOnly(ro bool) {
	m.readOnly = ro
	if ro {
		m.huhInput.Blur()
	}
}

func (m *Model[T]) validation() error {
	if m.required && len(m.value) == 0 {
		return errors.New("please select at least one option")
	}

	if m.validationFunc != nil {
		if err := m.validationFunc(m.value); err != nil {
			return err
		}
	}

	return nil
}

func WithRequired[T comparable]() Option[T] {
	return func(m *Model[T]) {
		m.required = true
	}
}

func (m *Model[T]) submit() tea.Cmd {
	if err := m.validation(); err != nil {
		m.errorMsg = err.Error()
		return nil
	}
	m.errorMsg = ""
	if m.onSubmit != nil {
		return m.onSubmit(m.Value())
	}
	return nil
}

// slicesEqual is a small helper for change detection since []T is not
// directly comparable with !=.
func slicesEqual[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
