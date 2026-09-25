package textinput

import (
	"strings"

	"github.com/MohsenBg/bgscan/internal/logger"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/input"
	"github.com/MohsenBg/bgscan/internal/ui/shared/env"
	"github.com/MohsenBg/bgscan/internal/ui/shared/layout"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
	"github.com/MohsenBg/bgscan/internal/ui/theme"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

type Option func(*Model)

// Model is the single-line text [input.Input].
type Model struct {
	id   ui.ComponentID
	name string

	layout *layout.Layout

	theme *theme.Theme
	log   logger.Set

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

func New(
	deps ui.Deps,
	title string,
	options ...Option,
) input.Input[string] {
	ti := textinput.New()
	m := &Model{
		id:                ui.NewComponentID(),
		name:              "input",
		layout:            deps.Layout,
		theme:             deps.Theme,
		log:               deps.Log,
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

func WithPlaceholder(p string) Option {
	return func(m *Model) {
		m.placeholder = p
		m.textinput.Placeholder = p
	}
}

func WithValue(value string) Option {
	return func(m *Model) {
		m.textinput.SetValue(value)
	}
}

func WithValidation(fn func(string) error) Option {
	return func(m *Model) {
		m.validationFunc = fn
	}
}

func WithCharLimit(limit int) Option {
	return func(m *Model) {
		m.textinput.CharLimit = limit
	}
}

func WithFocus() Option {
	return func(m *Model) {
		m.textinput.Focus()
	}
}

func WithReadOnly(ro bool) Option {
	return func(m *Model) {
		m.setReadOnly(ro)
	}
}

func WithOnChange(fn func(string) tea.Cmd) Option {
	return func(m *Model) {
		m.onChange = fn
	}
}

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

func (m *Model) Mode() env.Mode {
	return env.InputMode
}

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

func (m *Model) OnClose() tea.Cmd {
	return nil
}

func (m *Model) Theme() *theme.Theme { return m.theme }

func (m *Model) SetTheme(th *theme.Theme) { m.theme = th }

func (m *Model) Value() string {
	return m.textinput.Value()
}

func (m *Model) SetValue(value string) {
	m.textinput.SetValue(value)
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
		m.textinput.Blur()
	}
}

func (m *Model) validation(value string) error {
	if m.validationFunc == nil {
		return nil
	}
	return m.validationFunc(value)
}

// submit trims, validates, then invokes onSubmit.
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
