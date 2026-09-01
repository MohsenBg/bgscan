package tabs

import (
	"github.com/MohsenBg/bgscan/internal/ui/shared/env"
	"github.com/MohsenBg/bgscan/internal/ui/shared/layout"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// Tab holds a label and associated value.
type Tab[T any] struct {
	Label string
	Value T
}

// NewTab constructs a Tab.
func NewTab[T any](label string, value T) Tab[T] {
	return Tab[T]{Label: label, Value: value}
}

// Model is the tabs component.
type Model[T any] struct {
	layout *layout.Layout
	id     ui.ComponentID
	name   string

	tabs        []Tab[T]
	idx         int
	onSelectTab func(idx int, tab Tab[T]) tea.Cmd

	// styling options
	activeStyle   *lipgloss.Style
	inactiveStyle *lipgloss.Style
	separator     string
	alignment     lipgloss.Position
	maxWidth      int
	underline     bool
}

// Option configures the tabs Model.
type Option[T any] func(*Model[T])

// WithActiveStyle sets the style for the selected tab.
func WithActiveStyle[T any](s lipgloss.Style) Option[T] {
	return func(m *Model[T]) { m.activeStyle = &s }
}

// WithInactiveStyle sets the style for unselected tabs.
func WithInactiveStyle[T any](s lipgloss.Style) Option[T] {
	return func(m *Model[T]) { m.inactiveStyle = &s }
}

// WithSeparator sets the string placed between tabs (default: " ").
func WithSeparator[T any](sep string) Option[T] {
	return func(m *Model[T]) { m.separator = sep }
}

// WithAlignment sets the horizontal alignment of the tab row.
func WithAlignment[T any](align lipgloss.Position) Option[T] {
	return func(m *Model[T]) { m.alignment = align }
}

// WithMaxWidth sets the maximum width of the tabs row (0 = no limit).
func WithMaxWidth[T any](w int) Option[T] {
	return func(m *Model[T]) { m.maxWidth = w }
}

// WithUnderline enables an underline border under the active tab (default: true).
func WithUnderline[T any](u bool) Option[T] {
	return func(m *Model[T]) { m.underline = u }
}

// New creates a tabs model with the given options.
func New[T any](layout *layout.Layout, tabs []Tab[T], onSelectTab func(idx int, tab Tab[T]) tea.Cmd, opts ...Option[T]) *Model[T] {
	m := &Model[T]{
		layout:      layout,
		id:          ui.NewComponentID(),
		name:        "tabs",
		tabs:        tabs,
		idx:         0,
		onSelectTab: onSelectTab,
		maxWidth:    0, // unlimited by default
		alignment:   lipgloss.Left,
		separator:   " ",
		underline:   true,
	}
	// apply options before setting defaults
	for _, opt := range opts {
		opt(m)
	}
	// set default styles if not provided
	if m.activeStyle == nil {
		s := defaultActiveStyle()
		m.activeStyle = &s
	}
	if m.inactiveStyle == nil {
		s := defaultInactiveStyle()
		m.inactiveStyle = &s
	}
	return m
}

// ID returns the component's unique identifier.
func (m *Model[T]) ID() ui.ComponentID { return m.id }

// Name returns the component's name.
func (m *Model[T]) Name() string { return m.name }

// Mode returns the component's mode.
func (m *Model[T]) Mode() env.Mode { return env.NormalMode }

// Init is a no‑op for the tabs component.
func (m *Model[T]) Init() tea.Cmd { return nil }

// OnClose is called when the component is removed.
func (m *Model[T]) OnClose() tea.Cmd { return nil }

// CurrentTab returns the currently selected tab.
func (m *Model[T]) CurrentTab() *Tab[T] {
	if m.idx >= 0 && m.idx < len(m.tabs) {
		return &m.tabs[m.idx]
	}
	return nil
}

func (m *Model[T]) SetMaxWidth(width int) {
	m.maxWidth = width
}

// SelectTab sets the current index and returns the selected tab (or nil if out of range).
func (m *Model[T]) SelectTab(idx int) *Tab[T] {
	if idx >= 0 && idx < len(m.tabs) {
		m.idx = idx
		return &m.tabs[idx]
	}
	return nil
}

// NextTab moves to the next tab (wraps around).
func (m *Model[T]) NextTab() {
	if len(m.tabs) == 0 {
		return
	}
	m.idx = (m.idx + 1) % len(m.tabs)
}

// BackTab moves to the previous tab (wraps around).
func (m *Model[T]) BackTab() {
	if len(m.tabs) == 0 {
		return
	}
	m.idx--
	if m.idx < 0 {
		m.idx = len(m.tabs) - 1
	}
}

// selectTabCmd returns the command to call the onSelectTab callback.
func (m *Model[T]) selectTabCmd() tea.Cmd {
	if tab := m.CurrentTab(); tab != nil && m.onSelectTab != nil {
		return m.onSelectTab(m.idx, *tab)
	}
	return nil
}
