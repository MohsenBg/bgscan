package form

import (
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/inspector"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/notice"
	"github.com/MohsenBg/bgscan/internal/ui/shared/env"
	"github.com/MohsenBg/bgscan/internal/ui/shared/layout"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type Focus int

const (
	FocusInspector Focus = iota
	focusCount
)

type (
	Option func(*Model)

	ValidationFunc func(*Model) error
)

func WithName(name string) Option {
	return func(m *Model) {
		m.name = name
	}
}

func WithWidth(width int) Option {
	return func(m *Model) {
		m.width = width
	}
}

func WithHeight(height int) Option {
	return func(m *Model) {
		m.height = height
	}
}

func WithValidation(fn ValidationFunc) Option {
	return func(m *Model) {
		m.validation = fn
	}
}

func WithSave(fn tea.Cmd) Option {
	return func(m *Model) {
		m.onSave = fn
	}
}

func WithCancel(fn tea.Cmd) Option {
	return func(m *Model) {
		m.onCancel = fn
	}
}

type Model struct {
	id        ui.ComponentID
	layout    *layout.Layout
	inspector *inspector.Model

	name   string
	width  int
	height int

	focus Focus

	validation ValidationFunc
	onSave     tea.Cmd
	onCancel   tea.Cmd
}

func New(
	layout *layout.Layout,
	inspector *inspector.Model,
	opts ...Option,
) *Model {
	m := &Model{
		id:        ui.NewComponentID(),
		layout:    layout,
		inspector: inspector,
		height:    layout.BodyContentHeight(),
		width:     layout.BodyContentWidth(),
	}

	for _, opt := range opts {
		opt(m)
	}

	m.recalculateSize()
	return m
}

func (m *Model) ID() ui.ComponentID {
	return m.id
}

func (m *Model) Name() string {
	return m.name
}

func (m *Model) SetName(name string) {
	m.name = name
}

func (m *Model) Width() int {
	return m.width
}

func (m *Model) SetWidth(width int) {
	m.width = width
}

func (m *Model) Height() int {
	return m.height
}

func (m *Model) SetHeight(height int) {
	m.height = height
}

func (m *Model) Focus() Focus {
	return m.focus
}

func (m *Model) Init() tea.Cmd {
	if m.inspector == nil {
		return nil
	}

	return m.inspector.Init()
}

func (m *Model) Mode() env.Mode {
	return env.ManagedMode
}

func (m *Model) OnClose() tea.Cmd {
	if m.inspector == nil {
		return nil
	}

	return m.inspector.OnClose()
}

func (m *Model) Validate() error {
	if m.validation == nil {
		return nil
	}

	return m.validation(m)
}

func (m *Model) Save() tea.Cmd {
	if err := m.Validate(); err != nil {
		return notice.NewNoticeCmd(
			m.layout,
			"Validation Error",
			err.Error(),
			notice.NOTICE_ERROR,
		)
	}

	if m.onSave == nil {
		return nil
	}

	return m.onSave
}

func (m *Model) Cancel() tea.Cmd {
	if m.onCancel == nil {
		return nil
	}

	return m.onCancel
}

func (m *Model) NextFocus() {
	m.focus = (m.focus + 1) % focusCount
}

func (m *Model) PrevFocus() {
	m.focus = (m.focus - 1 + focusCount) % focusCount
}

const (
	minHeight          = 20
	minWidth           = 40
	minInspectorHeight = 15
)

func (m *Model) recalculateSize() {
	height := max(minHeight, m.height)
	width := max(minWidth, m.width)

	m.height = height
	m.width = width

	titleHeight := lipgloss.Height(m.renderTitle())
	hintsHeight := lipgloss.Height(m.renderKeyHints())
	spacing := 2

	inspectorHeight := height - titleHeight - hintsHeight - spacing
	m.inspector.SetHeight(max(minInspectorHeight, inspectorHeight))
	m.inspector.SetWidth(width)
}
