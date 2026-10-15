package selectinput

import (
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/input"
	"github.com/MohsenBg/bgscan/internal/ui/shared/keys"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
	"github.com/MohsenBg/bgscan/internal/ui/theme"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
)

// Option configures a Model at construction time.
type Option[T comparable] func(*Model[T])

// Model is the select-based implementation of [input.Input].
type Model[T comparable] struct {
	id   ui.ComponentID
	name string

	env *ui.Env

	title    string
	errorMsg string

	value    T
	options  []huh.Option[T]
	huhInput *huh.Select[T]
	readOnly bool
	focused  bool

	validationFunc func(value T) error

	onChange func(T) tea.Cmd
	onSubmit func(T) tea.Cmd
}

func New[T comparable](
	env *ui.Env,
	title string,
	options ...Option[T],
) input.Input[T] {
	m := &Model[T]{
		id:    ui.NewComponentID(),
		name:  "select",
		env:   env,
		title: title,
	}

	m.huhInput = huh.NewSelect[T]().
		Title(title).
		Value(&m.value)

	m.huhInput.WithKeyMap(m.keyMap())
	m.huhInput.WithTheme(theme.NewHuhTheme(&m.env.Theme))
	for _, opt := range options {
		opt(m)
	}
	inp := m.huhInput.Options(m.options...).WithWidth(m.Width())
	m.huhInput = inp.(*huh.Select[T])

	return m
}

// WithOptions sets the selectable options.
func WithOptions[T comparable](opts ...huh.Option[T]) Option[T] {
	return func(m *Model[T]) {
		m.options = opts
	}
}

// WithValue sets the initial selected value.
func WithValue[T comparable](value T) Option[T] {
	return func(m *Model[T]) {
		m.value = value
	}
}

// WithValidation sets the function used to validate the input's value.
func WithValidation[T comparable](fn func(T) error) Option[T] {
	return func(m *Model[T]) {
		m.validationFunc = fn
	}
}

// WithFocus focuses the input on creation.
func WithFocus[T comparable]() Option[T] {
	return func(m *Model[T]) {
		m.focused = true
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
func WithOnChange[T comparable](fn func(T) tea.Cmd) Option[T] {
	return func(m *Model[T]) {
		m.onChange = fn
	}
}

// WithOnSubmit registers a callback invoked when the value is submitted.
func WithOnSubmit[T comparable](fn func(T) tea.Cmd) Option[T] {
	return func(m *Model[T]) {
		m.onSubmit = fn
	}
}

func (m *Model[T]) Init() tea.Cmd { return nil }

func (m *Model[T]) ID() ui.ComponentID { return m.id }

func (m *Model[T]) Name() string { return m.name }

func (m *Model[T]) Mode() ui.KeyMode { return ui.KeyModeNormal }

// Title returns the raw dialog-frame title.
func (m *Model[T]) Title() string { return m.title }

func (m *Model[T]) DialogHints() keys.Hints {
	return DialogHints(m.env.Keys.SelectInput)
}

// DialogHints returns the frame hints for a select with the given keys.
func DialogHints(k keys.SelectInputKeys) keys.Hints {
	return append(
		keys.Hints{keys.Combine("move", k.Up, k.Down)},
		keys.Short(k.Submit)...,
	)
}

// keyMap translates the shared keys into the huh select keymap.
func (m *Model[T]) keyMap() *huh.KeyMap {
	k := m.env.Keys.SelectInput
	km := huh.NewDefaultKeyMap()
	km.Select.Submit.SetKeys(k.Submit.Keys...)
	km.Select.Prev.SetKeys(k.Prev.Keys...)
	km.Select.Next.SetKeys(k.Next.Keys...)
	km.Select.Up.SetKeys(k.Up.Keys...)
	km.Select.Down.SetKeys(k.Down.Keys...)
	km.Select.HalfPageUp.SetKeys(k.HalfPageUp.Keys...)
	km.Select.HalfPageDown.SetKeys(k.HalfPageDown.Keys...)
	km.Select.GotoTop.SetKeys(k.GotoTop.Keys...)
	km.Select.GotoBottom.SetKeys(k.GotoBottom.Keys...)
	km.Select.Left.SetKeys(k.Left.Keys...)
	km.Select.Right.SetKeys(k.Right.Keys...)
	km.Select.Filter.SetKeys(k.Filter.Keys...)
	km.Select.SetFilter.SetKeys(k.SetFilter.Keys...)
	km.Select.ClearFilter.SetKeys(k.ClearFilter.Keys...)
	return km
}

func (m *Model[T]) Width() int {
	return min(50, m.env.Layout.Body().Width)
}

func (m *Model[T]) CloseCmd() tea.Cmd {
	return func() tea.Msg {
		return ui.CloseComponentMsg{ID: m.ID()}
	}
}

func (m *Model[T]) OnClose() tea.Cmd { return nil }

// rebuild recreates the huh field from the live theme: huh's WithTheme is
// set-once and ignores later calls, so a fresh field is built preserving
// value, options and focus.
func (m *Model[T]) rebuild() {
	m.huhInput = huh.NewSelect[T]().
		Title(m.title).
		Value(&m.value)
	m.huhInput.WithKeyMap(m.keyMap())
	m.huhInput.WithTheme(theme.NewHuhTheme(&m.env.Theme))
	inp := m.huhInput.Options(m.options...).WithWidth(m.Width())
	m.huhInput = inp.(*huh.Select[T])
	if m.readOnly {
		m.huhInput.Blur()
	} else if m.focused {
		m.huhInput.Focus()
	}
}

func (m *Model[T]) Value() T { return m.value }

func (m *Model[T]) SetValue(value T) {
	m.value = value
	m.huhInput = m.huhInput.Value(&m.value)
}

func (m *Model[T]) ReadOnly() bool { return m.readOnly }

func (m *Model[T]) SetReadOnly(ro bool) { m.setReadOnly(ro) }

func (m *Model[T]) OnValidate(fn func(T) error) { m.validationFunc = fn }

func (m *Model[T]) OnChange(fn func(T) tea.Cmd) { m.onChange = fn }

func (m *Model[T]) OnSubmit(fn func(T) tea.Cmd) { m.onSubmit = fn }

// AppendOnSubmit chains fn after any previously registered onSubmit
// callback instead of replacing it.
func (m *Model[T]) AppendOnSubmit(fn func() tea.Cmd) {
	prev := m.onSubmit
	m.onSubmit = func(value T) tea.Cmd {
		if prev == nil {
			return fn()
		}
		return tea.Sequence(prev(value), fn())
	}
}

func (m *Model[T]) setReadOnly(ro bool) {
	m.readOnly = ro
	if ro {
		m.focused = false
		m.huhInput.Blur()
	}
}

func (m *Model[T]) validation() error {
	if m.validationFunc == nil {
		return nil
	}
	return m.validationFunc(m.Value())
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
