package textarea

import (
	"strings"

	"github.com/MohsenBg/bgscan/internal/ui/components/basic/input"
	"github.com/MohsenBg/bgscan/internal/ui/shared/keys"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
)

type Option func(*Model)

// Model is the multi-line string implementation of [input.Input].
type Model struct {
	id   ui.ComponentID
	name string

	env *ui.Env

	title       string
	placeholder string
	errorMsg    string

	textarea textarea.Model
	readOnly bool
	height   int

	validationFunc    func(value string) error
	dynamicValidation bool

	onChange func(string) tea.Cmd
	onSubmit func(string) tea.Cmd

	allowNewline bool
}

func New(
	env *ui.Env,
	title string,
	options ...Option,
) input.Input[string] {
	ta := textarea.New()
	m := &Model{
		id:                ui.NewComponentID(),
		name:              "textarea",
		env:               env,
		title:             title,
		textarea:          ta,
		height:            3,
		dynamicValidation: false,
		allowNewline:      true,
	}

	m.textarea.KeyMap = m.keyMap()
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

func WithPlaceholder(p string) Option {
	return func(m *Model) {
		m.placeholder = p
		m.textarea.Placeholder = p
	}
}

func WithValue(value string) Option {
	return func(m *Model) {
		m.textarea.SetValue(value)
	}
}

func WithValidation(fn func(string) error) Option {
	return func(m *Model) {
		m.validationFunc = fn
	}
}

func WithCharLimit(limit int) Option {
	return func(m *Model) {
		m.textarea.CharLimit = limit
	}
}

func WithHeight(height int) Option {
	return func(m *Model) {
		m.height = height
	}
}

// WithNewlines controls whether Shift+Enter inserts a newline.
// When disabled, newlines cannot be entered at all.
func WithNewlines(enabled bool) Option {
	return func(m *Model) {
		m.allowNewline = enabled
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

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) ID() ui.ComponentID {
	return m.id
}

func (m *Model) Name() string {
	return m.name
}

func (m *Model) Mode() ui.KeyMode {
	return ui.KeyModeInput
}

// Title returns the input prompt as raw data for the dialog frame.
func (m *Model) Title() string { return m.title }

func (m *Model) DialogHints() keys.Hints {
	return DialogHints(m.env.Keys.TextArea)
}

// DialogHints returns the frame hints for a text area with the given keys.
func DialogHints(k keys.TextAreaKeys) keys.Hints {
	return keys.Short(k.Submit)
}

// keyMap translates the shared keys into the bubbles textarea keymap.
func (m *Model) keyMap() textarea.KeyMap {
	k := m.env.Keys.TextArea
	defaultKeyMap := textarea.DefaultKeyMap()
	defaultKeyMap.CharacterForward.SetKeys(k.CharacterForward.Keys...)
	defaultKeyMap.CharacterBackward.SetKeys(k.CharacterBackward.Keys...)
	defaultKeyMap.WordForward.SetKeys(k.WordForward.Keys...)
	defaultKeyMap.WordBackward.SetKeys(k.WordBackward.Keys...)
	defaultKeyMap.LineNext.SetKeys(k.LineNext.Keys...)
	defaultKeyMap.LinePrevious.SetKeys(k.LinePrevious.Keys...)
	defaultKeyMap.DeleteWordBackward.SetKeys(k.DeleteWordBackward.Keys...)
	defaultKeyMap.DeleteWordForward.SetKeys(k.DeleteWordForward.Keys...)
	defaultKeyMap.DeleteAfterCursor.SetKeys(k.DeleteAfterCursor.Keys...)
	defaultKeyMap.DeleteBeforeCursor.SetKeys(k.DeleteBeforeCursor.Keys...)
	defaultKeyMap.InsertNewline.SetKeys(k.InsertNewline.Keys...)
	defaultKeyMap.DeleteCharacterBackward.SetKeys(k.DeleteCharacterBackward.Keys...)
	defaultKeyMap.DeleteCharacterForward.SetKeys(k.DeleteCharacterForward.Keys...)
	defaultKeyMap.LineStart.SetKeys(k.LineStart.Keys...)
	defaultKeyMap.LineEnd.SetKeys(k.LineEnd.Keys...)
	defaultKeyMap.PageUp.SetKeys(k.PageUp.Keys...)
	defaultKeyMap.PageDown.SetKeys(k.PageDown.Keys...)
	defaultKeyMap.Paste.SetKeys(k.Paste.Keys...)
	defaultKeyMap.InputBegin.SetKeys(k.InputBegin.Keys...)
	defaultKeyMap.InputEnd.SetKeys(k.InputEnd.Keys...)
	defaultKeyMap.CapitalizeWordForward.SetKeys(k.CapitalizeWordForward.Keys...)
	defaultKeyMap.LowercaseWordForward.SetKeys(k.LowercaseWordForward.Keys...)
	defaultKeyMap.UppercaseWordForward.SetKeys(k.UppercaseWordForward.Keys...)
	defaultKeyMap.TransposeCharacterBackward.SetKeys(k.TransposeCharacterBackward.Keys...)
	defaultKeyMap.SelectCharacterForward.SetKeys(k.SelectCharacterForward.Keys...)
	defaultKeyMap.SelectCharacterBackward.SetKeys(k.SelectCharacterBackward.Keys...)
	defaultKeyMap.SelectWordForward.SetKeys(k.SelectWordForward.Keys...)
	defaultKeyMap.SelectWordBackward.SetKeys(k.SelectWordBackward.Keys...)
	defaultKeyMap.SelectLineUp.SetKeys(k.SelectLineUp.Keys...)
	defaultKeyMap.SelectLineDown.SetKeys(k.SelectLineDown.Keys...)
	defaultKeyMap.SelectAll.SetKeys(k.SelectAll.Keys...)
	defaultKeyMap.CopySelection.SetKeys(k.CopySelection.Keys...)
	return defaultKeyMap
}

func (m *Model) Width() int {
	return min(50, m.env.Layout.Body().Width)
}

func (m *Model) CloseCmd() tea.Cmd {
	return func() tea.Msg {
		return ui.CloseComponentMsg{ID: m.ID()}
	}
}

func (m *Model) OnClose() tea.Cmd {
	return nil
}

func (m *Model) Value() string {
	return m.textarea.Value()
}

func (m *Model) SetValue(value string) {
	m.textarea.SetValue(value)
}

func (m *Model) ReadOnly() bool {
	return m.readOnly
}

func (m *Model) SetReadOnly(ro bool) {
	m.setReadOnly(ro)
}

func (m *Model) OnValidate(fn func(string) error) {
	m.validationFunc = fn
}

func (m *Model) OnChange(fn func(string) tea.Cmd) {
	m.onChange = fn
}

func (m *Model) OnSubmit(fn func(string) tea.Cmd) {
	m.onSubmit = fn
}

// AppendOnSubmit chains fn after any previously registered onSubmit
// callback instead of replacing it.
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
		m.textarea.Blur()
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
