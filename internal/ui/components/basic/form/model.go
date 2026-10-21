package form

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/MohsenBg/bgscan/internal/ui/components/basic/inspector"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/notice"
	"github.com/MohsenBg/bgscan/internal/ui/shared/keys"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
)

const (
	minHeight          = 20
	minWidth           = 40
	minInspectorHeight = 15
)

type (
	// Option configures a Model.
	Option func(*Model)

	// ValidationFunc validates the whole form.
	ValidationFunc func(*Model) error
)

func WithName(name string) Option {
	return func(m *Model) { m.name = name }
}

func WithWidth(width int) Option {
	return func(m *Model) { m.width = width }
}

func WithHeight(height int) Option {
	return func(m *Model) { m.height = height }
}

func WithValidation(fn ValidationFunc) Option {
	return func(m *Model) { m.validation = fn }
}

func WithSave(cmd tea.Cmd) Option {
	return func(m *Model) { m.onSave = cmd }
}

func WithCancel(cmd tea.Cmd) Option {
	return func(m *Model) { m.onCancel = cmd }
}

// Model is a titled, sized frame around an inspector with save/cancel actions.
type Model struct {
	id         ui.ComponentID
	env        *ui.Env
	inspector  *inspector.Model
	name       string
	width      int
	height     int
	validation ValidationFunc
	onSave     tea.Cmd
	onCancel   tea.Cmd
}

func New(env *ui.Env, insp *inspector.Model, opts ...Option) *Model {
	m := &Model{
		id:        ui.NewComponentID(),
		env:       env,
		inspector: insp,
		width:     env.Layout.Body().Width,
		height:    env.Layout.Body().Height,
	}

	for _, opt := range opts {
		opt(m)
	}

	m.recalculateSize()
	return m
}

func (m *Model) ID() ui.ComponentID { return m.id }
func (m *Model) Name() string       { return m.name }
func (m *Model) Mode() ui.KeyMode   { return ui.KeyModeManage }

// Title returns the raw form heading for dialog frames.
func (m *Model) Title() string { return m.name }

func (m *Model) SetName(name string) { m.name = name }

func (m *Model) Width() int { return m.width }
func (m *Model) SetWidth(width int) {
	m.width = width
	m.recalculateSize()
}

func (m *Model) Height() int { return m.height }
func (m *Model) SetHeight(height int) {
	m.height = height
	m.recalculateSize()
}

func (m *Model) Init() tea.Cmd {
	if m.inspector == nil {
		return nil
	}
	return m.inspector.Init()
}

func (m *Model) OnClose() tea.Cmd {
	if m.inspector == nil {
		return nil
	}
	return m.inspector.OnClose()
}

// DialogHints returns save hints for a dialog frame.
func DialogHints(formKeys keys.FormKeys) keys.Hints {
	return keys.Hints{formKeys.Save.ShortItem()}
}

// Validate runs the form validator, if any.
func (m *Model) Validate() error {
	if m.validation == nil {
		return nil
	}
	return m.validation(m)
}

// Save validates the form and invokes the save command.
func (m *Model) Save() tea.Cmd {
	if err := m.Validate(); err != nil {
		return notice.NewNoticeCmd(
			m.env,
			"Validation Error",
			err.Error(),
			notice.NOTICE_ERROR,
		)
	}
	return m.onSave
}

// Cancel invokes the cancel command, if any.
func (m *Model) Cancel() tea.Cmd {
	return m.onCancel
}

// Inspector exposes the wrapped inspector.
func (m *Model) Inspector() *inspector.Model {
	return m.inspector
}

func (m *Model) recalculateSize() {
	m.width = max(minWidth, m.width)
	m.height = max(minHeight, m.height)

	if m.inspector == nil {
		return
	}

	const spacing = 2
	titleHeight := lipgloss.Height(m.renderTitle())
	inspectorHeight := max(minInspectorHeight, m.height-titleHeight-spacing)

	m.inspector.SetWidth(m.width)
	m.inspector.SetHeight(inspectorHeight)
}

// formKeysActive prevents the form from stealing keys while the inspector
// is in an editing or otherwise non-normal mode.
func (m *Model) formKeysActive() bool {
	if m.inspector == nil {
		return true
	}

	switch m.inspector.Mode() {
	case ui.KeyModeNormal, ui.KeyModeManage:
		return true
	default:
		return false
	}
}
