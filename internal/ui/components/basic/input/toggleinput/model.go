package toggleinput

import (
	"slices"

	"github.com/MohsenBg/bgscan/internal/ui/components/basic/input"
	"github.com/MohsenBg/bgscan/internal/ui/shared/keys"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
	"github.com/MohsenBg/bgscan/internal/ui/theme"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
)

// Option configures a Model at construction time.
type Option func(*Model)

// Model is the boolean toggle implementation of [input.Input].
type Model struct {
	id   ui.ComponentID
	name string

	env *ui.Env

	title    string
	errorMsg string

	value    bool
	affirm   string
	negate   string
	huhInput *huh.Confirm
	readOnly bool
	focused  bool

	validationFunc func(value bool) error

	onChange func(bool) tea.Cmd
	onSubmit func(bool) tea.Cmd
}

func New(
	env *ui.Env,
	title string,
	options ...Option,
) input.Input[bool] {
	m := &Model{
		id:     ui.NewComponentID(),
		name:   "toggle",
		env:    env,
		title:  title,
		affirm: "Yes",
		negate: "No",
	}

	m.huhInput = huh.NewConfirm().
		Value(&m.value)

	m.huhInput.WithKeyMap(m.keyMap())
	m.huhInput.WithTheme(theme.NewHuhTheme(&m.env.Theme))
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
		m.focused = true
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

func (m *Model) Init() tea.Cmd { return nil }

func (m *Model) ID() ui.ComponentID { return m.id }

func (m *Model) Name() string { return m.name }

func (m *Model) Mode() ui.KeyMode { return ui.KeyModeNormal }

// Title returns the input prompt as raw data for the dialog frame.
func (m *Model) Title() string { return m.title }

func (m *Model) DialogHints() keys.Hints {
	k := m.env.Keys.ToggleInput
	tl := k.Toggle

	tl.Keys = slices.Clone(tl.Keys)
	tl.Keys[0], tl.Keys[1] = tl.Keys[1], tl.Keys[0]

	hints := append(
		keys.Hints{},
		keys.Combine("move", tl, k.Toggle),
	)

	hints = append(hints, keys.Short(k.Submit)...)

	return hints
}

func (m *Model) Width() int {
	return min(50, m.env.Layout.Body().Width)
}

func (m *Model) CloseCmd() tea.Cmd {
	return func() tea.Msg {
		return ui.CloseComponentMsg{ID: m.ID()}
	}
}

func (m *Model) OnClose() tea.Cmd { return nil }

// rebuild recreates the huh field from the live theme: huh's WithTheme is
// set-once and ignores later calls, so a fresh field is built preserving
// value, labels and focus.
func (m *Model) rebuild() {
	m.huhInput = huh.NewConfirm().
		Title(m.title).
		Value(&m.value)
	m.huhInput.WithKeyMap(m.keyMap())
	m.huhInput.WithTheme(theme.NewHuhTheme(&m.env.Theme))
	inp := m.huhInput.
		Affirmative(m.affirm).
		Negative(m.negate).
		WithWidth(m.Width())
	m.huhInput = inp.(*huh.Confirm)
	if m.readOnly {
		m.huhInput.Blur()
	} else if m.focused {
		m.huhInput.Focus()
	}
}

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

// AppendOnSubmit chains fn after any previously registered onSubmit
// callback instead of replacing it.
func (m *Model) AppendOnSubmit(fn func() tea.Cmd) {
	prev := m.onSubmit
	m.onSubmit = func(value bool) tea.Cmd {
		if prev == nil {
			return fn()
		}
		return tea.Sequence(prev(value), fn())
	}
}

func (m *Model) setReadOnly(ro bool) {
	m.readOnly = ro
	if ro {
		m.focused = false
		m.huhInput.Blur()
	}
}

func (m *Model) validation() error {
	if m.validationFunc == nil {
		return nil
	}
	return m.validationFunc(m.Value())
}

// keyMap translates the shared keys into the huh keymap.
func (m *Model) keyMap() *huh.KeyMap {
	k := m.env.Keys.ToggleInput
	km := huh.NewDefaultKeyMap()
	km.Confirm.Prev.SetKeys(k.Prev.Keys...)
	km.Confirm.Next.SetKeys(k.Next.Keys...)
	km.Confirm.Submit.SetKeys(k.Submit.Keys...)
	km.Confirm.Toggle.SetKeys(k.Toggle.Keys...)
	km.Confirm.Accept.SetKeys(k.Accept.Keys...)
	km.Confirm.Reject.SetKeys(k.Reject.Keys...)
	return km
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
