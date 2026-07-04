package toggleinput

import (
	"bgscan/internal/ui/components/basic/input"
	"bgscan/internal/ui/shared/env"
	"bgscan/internal/ui/shared/layout"
	"bgscan/internal/ui/shared/ui"
	"bgscan/internal/ui/theme"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
)

// Option configures a Model at construction time.
type Option func(*Model)

// Model is the boolean toggle implementation of [input.Input].
type Model struct {
	// Component identity
	id   ui.ComponentID
	name string

	// Layout reference
	layout *layout.Layout

	// UI content
	title    string
	errorMsg string

	// Input field
	value    bool
	affirm   string
	negate   string
	huhInput *huh.Confirm
	readOnly bool

	// Validation
	validationFunc func(value bool) error

	// Callbacks
	onChange func(bool) tea.Cmd
	onSubmit func(bool) tea.Cmd
}

// New creates a new toggle input component.
func New(
	l *layout.Layout,
	title string,
	options ...Option,
) input.Input[bool] {
	m := &Model{
		id:     ui.NewComponentID(),
		name:   "toggle",
		layout: l,
		title:  title,
		affirm: "Yes",
		negate: "No",
	}

	m.huhInput = huh.NewConfirm().
		Title(title).
		Value(&m.value)

	m.huhInput.WithKeyMap(huh.NewDefaultKeyMap())
	m.huhInput.WithTheme(theme.NewHuhTheme())
	for _, opt := range options {
		opt(m)
	}

	inp := m.huhInput.
		Affirmative(m.affirm).
		Negative(m.negate).
		WithWidth(m.Width())
	m.huhInput = inp.(*huh.Confirm)

	return m
}

// --- Options -----------------------------------------------------------

// WithValue sets the initial toggle state.
func WithValue(value bool) Option {
	return func(m *Model) {
		m.value = value
	}
}

// WithLabels sets the affirmative/negative button labels.
func WithLabels(affirm, negate string) Option {
	return func(m *Model) {
		m.affirm = affirm
		m.negate = negate
	}
}

// WithValidation sets the function used to validate the input's value.
func WithValidation(fn func(bool) error) Option {
	return func(m *Model) {
		m.validationFunc = fn
	}
}

// WithFocus focuses the input on creation.
func WithFocus() Option {
	return func(m *Model) {
		m.huhInput.Focus()
	}
}

// WithReadOnly sets the initial read-only state of the input.
func WithReadOnly(ro bool) Option {
	return func(m *Model) {
		m.setReadOnly(ro)
	}
}

// WithOnChange registers a callback invoked whenever the value changes.
func WithOnChange(fn func(bool) tea.Cmd) Option {
	return func(m *Model) {
		m.onChange = fn
	}
}

// WithOnSubmit registers a callback invoked when the value is submitted.
func WithOnSubmit(fn func(bool) tea.Cmd) Option {
	return func(m *Model) {
		m.onSubmit = fn
	}
}

// --- ui.Component --------------------------------------------------------

func (m *Model) Init() tea.Cmd { return nil }

func (m *Model) ID() ui.ComponentID { return m.id }

func (m *Model) Name() string { return m.name }

func (m *Model) Mode() env.Mode { return env.NormalMode }

func (m *Model) Width() int {
	if m.layout == nil {
		return 50
	}
	return min(50, m.layout.Body.Width)
}

func (m *Model) CloseCmd() tea.Cmd {
	return func() tea.Msg {
		return ui.CloseComponentMsg{ID: m.ID()}
	}
}

func (m *Model) OnClose() tea.Cmd { return nil }

// --- input.Input[bool] --------------------------------------------------------

func (m *Model) Value() bool { return m.value }

func (m *Model) SetValue(value bool) {
	m.value = value
	m.huhInput = m.huhInput.Value(&m.value)
}

func (m *Model) ReadOnly() bool { return m.readOnly }

func (m *Model) SetReadOnly(ro bool) { m.setReadOnly(ro) }

func (m *Model) OnValidate(fn func(bool) error) { m.validationFunc = fn }

func (m *Model) OnChange(fn func(bool) tea.Cmd) { m.onChange = fn }

func (m *Model) OnSubmit(fn func(bool) tea.Cmd) { m.onSubmit = fn }

// AppendOnSubmit implements [input.Input]. It chains fn after any
// previously registered onSubmit callback rather than replacing it.
func (m *Model) AppendOnSubmit(fn func() tea.Cmd) {
	prev := m.onSubmit
	m.onSubmit = func(value bool) tea.Cmd {
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
		m.huhInput.Blur()
	}
}

func (m *Model) validation() error {
	if m.validationFunc == nil {
		return nil
	}
	return m.validationFunc(m.Value())
}

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
