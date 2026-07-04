package textarea

import (
	"bgscan/internal/ui/components/basic/input"
	"bgscan/internal/ui/shared/env"
	"bgscan/internal/ui/shared/layout"
	"bgscan/internal/ui/shared/ui"

	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
)

// Option configures a Model at construction time.
type Option func(*Model)

// Model is the multi-line string implementation of [input.Input].
type Model struct {
	// Component identity
	id   ui.ComponentID
	name string

	// Layout reference
	layout *layout.Layout

	// UI content
	title       string
	placeholder string
	errorMsg    string

	// Input field
	textarea textarea.Model
	readOnly bool
	height   int

	// Validation
	validationFunc    func(value string) error
	dynamicValidation bool

	// Callbacks
	onChange func(string) tea.Cmd
	onSubmit func(string) tea.Cmd
}

// New creates a new multi-line text input component.
func New(
	layout *layout.Layout,
	title string,
	options ...Option,
) input.Input[string] {
	ta := textarea.New()
	m := &Model{
		id:                ui.NewComponentID(),
		name:              "textarea",
		layout:            layout,
		title:             title,
		textarea:          ta,
		height:            3,
		dynamicValidation: false,
	}

	m.textarea.SetWidth(m.Width())
	for _, opt := range options {
		opt(m)
	}

	m.textarea.DynamicHeight = true
	m.textarea.MinHeight = 3
	m.textarea.MaxHeight = m.height
	m.textarea.SetHeight(m.height)

	return m
}

// --- Options ---------------------------------------------------------------

// WithPlaceholder sets the placeholder text shown when the input is empty.
func WithPlaceholder(p string) Option {
	return func(m *Model) {
		m.placeholder = p
		m.textarea.Placeholder = p
	}
}

// WithValue sets the initial value of the input.
func WithValue(value string) Option {
	return func(m *Model) {
		m.textarea.SetValue(value)
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
		m.textarea.CharLimit = limit
	}
}

// WithHeight sets the maximum height of the textarea.
func WithHeight(height int) Option {
	return func(m *Model) {
		m.height = height
	}
}

// WithFocus focuses the input on creation.
func WithFocus() Option {
	return func(m *Model) {
		m.textarea.Focus()
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

// --- ui.Component ------------------------------------------------------------

// Init initializes the component.
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

// --- input.Input[string] ----------------------------------------------------

// Value implements [input.Input].
func (m *Model) Value() string {
	return m.textarea.Value()
}

// SetValue implements [input.Input].
func (m *Model) SetValue(value string) {
	m.textarea.SetValue(value)
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

// --- internal helpers --------------------------------------------------------

func (m *Model) setReadOnly(ro bool) {
	m.readOnly = ro
	if ro {
		m.textarea.Blur()
	}
}

// validation runs the configured validation function against the
// current value and returns an error if validation fails.
func (m *Model) validation() error {
	if m.validationFunc == nil {
		return nil
	}
	return m.validationFunc(m.Value())
}

// submit validates the current value and, if valid, invokes onSubmit.
func (m *Model) submit() tea.Cmd {
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
