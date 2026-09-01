package textinput

import (
	"strings"

	"github.com/MohsenBg/bgscan/internal/ui/components/basic/input"
	"github.com/MohsenBg/bgscan/internal/ui/shared/env"
	"github.com/MohsenBg/bgscan/internal/ui/shared/layout"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

type Option func(*Model)

// Model is the single-line text implementation of [input.Input].
type Model struct {
	id   ui.ComponentID
	name string

	layout *layout.Layout

	title       string
	placeholder string
	errorMsg    string

	textinput textinput.Model
	readOnly  bool

	validationFunc    func(value string) error
	dynamicValidation bool

	onChange func(string) tea.Cmd
	onSubmit func(string) tea.Cmd
}

// New creates a single-line text input component.
func New(
	layout *layout.Layout,
	title string,
	options ...Option,
) input.Input[string] {
	ti := textinput.New()
	m := &Model{
		id:                ui.NewComponentID(),
		name:              "input",
		layout:            layout,
		title:             title,
		textinput:         ti,
		dynamicValidation: false,
	}

	m.textinput.SetWidth(m.Width())
	for _, opt := range options {
		opt(m)
	}

	return m
}

// WithPlaceholder sets the placeholder text shown when the input is empty.
func WithPlaceholder(p string) Option {
	return func(m *Model) {
		m.placeholder = p
		m.textinput.Placeholder = p
	}
}

// WithValue sets the initial value of the input.
func WithValue(value string) Option {
	return func(m *Model) {
		m.textinput.SetValue(value)
	}
}

// WithValidation sets the function used to validate the input's value.
func WithValidation(fn func(string) error) Option {
	return func(m *Model) {
		m.validationFunc = fn
	}
}

// WithCharLimit sets the maximum number of characters the input will accept.
func WithCharLimit(limit int) Option {
	return func(m *Model) {
		m.textinput.CharLimit = limit
	}
}

// WithFocus focuses the input on creation.
func WithFocus() Option {
	return func(m *Model) {
		m.textinput.Focus()
	}
}

// WithReadOnly sets the initial read-only state of the input.
func WithReadOnly(ro bool) Option {
	return func(m *Model) {
		m.setReadOnly(ro)
	}
}

// WithOnChange registers a callback invoked whenever the value changes.
func WithOnChange(fn func(string) tea.Cmd) Option {
	return func(m *Model) {
		m.onChange = fn
	}
}

// WithOnSubmit registers a callback invoked when the value is submitted.
func WithOnSubmit(fn func(string) tea.Cmd) Option {
	return func(m *Model) {
		m.onSubmit = fn
	}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

// ID returns the component identifier.
func (m *Model) ID() ui.ComponentID {
	return m.id
}

// Name returns the component name.
func (m *Model) Name() string {
	return m.name
}

// Mode returns the input mode used by this component.
func (m *Model) Mode() env.Mode {
	return env.InputMode
}

// Width calculates the maximum width of the input field.
func (m *Model) Width() int {
	if m.layout == nil {
		return 50
	}
	return min(50, m.layout.Body.Width)
}

// CloseCmd returns a command that closes this component.
func (m *Model) CloseCmd() tea.Cmd {
	return func() tea.Msg {
		return ui.CloseComponentMsg{ID: m.ID()}
	}
}

// OnClose is called when the component is removed.
func (m *Model) OnClose() tea.Cmd {
	return nil
}

func (m *Model) Value() string {
	return m.textinput.Value()
}

// SetValue implements [input.Input].
func (m *Model) SetValue(value string) {
	m.textinput.SetValue(value)
}

// ReadOnly implements [input.Input].
func (m *Model) ReadOnly() bool {
	return m.readOnly
}

// SetReadOnly implements [input.Input].
func (m *Model) SetReadOnly(ro bool) {
	m.setReadOnly(ro)
}

func (m *Model) OnValidate(fn func(string) error) {
	m.validationFunc = fn
}

// OnChange implements [input.Input].
func (m *Model) OnChange(fn func(string) tea.Cmd) {
	m.onChange = fn
}

// OnSubmit implements [input.Input].
func (m *Model) OnSubmit(fn func(string) tea.Cmd) {
	m.onSubmit = fn
}

// AppendOnSubmit implements [input.Input]. It chains fn after any
// previously registered onSubmit callback rather than replacing it.
func (m *Model) AppendOnSubmit(fn func() tea.Cmd) {
	prev := m.onSubmit
	m.onSubmit = func(value string) tea.Cmd {
		if prev == nil {
			return fn()
		}
		return tea.Sequence(prev(value), fn())
	}
}

func (m *Model) setReadOnly(ro bool) {
	m.readOnly = ro
	if ro {
		m.textinput.Blur()
	}
}

// validation runs the configured validation function against the given
// value and returns an error if validation fails.
func (m *Model) validation(value string) error {
	if m.validationFunc == nil {
		return nil
	}
	return m.validationFunc(value)
}

// submit trims whitespace from the current value, validates the trimmed
// value and, if valid, invokes onSubmit with it.
func (m *Model) submit() tea.Cmd {
	value := strings.TrimSpace(m.Value())
	m.SetValue(value)
	if err := m.validation(value); err != nil {
		m.errorMsg = err.Error()
		return nil
	}
	m.errorMsg = ""
	if m.onSubmit != nil {
		return m.onSubmit(value)
	}
	return nil
}
