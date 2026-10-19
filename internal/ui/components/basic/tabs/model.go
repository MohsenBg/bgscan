package tabs

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
)

// Tab holds a label and associated value.
type Tab[T any] struct {
	Label string
	Value T
}

func NewTab[T any](label string, value T) Tab[T] {
	return Tab[T]{Label: label, Value: value}
}

// Model is a generic tab row with optional underline or animated indicator.
type Model[T any] struct {
	env *ui.Env
	id  ui.ComponentID

	tabs        []Tab[T]
	idx         int
	onSelectTab func(idx int, tab Tab[T]) tea.Cmd

	activeStyle   *lipgloss.Style
	inactiveStyle *lipgloss.Style
	separator     string
	maxWidth      int
	underline     bool

	customActive   bool
	customInactive bool

	animated   bool
	slide      float64
	slideFrom  float64
	slideTo    float64
	slideStart time.Time
	title      string
}

type Option[T any] func(*Model[T])

func WithActiveStyle[T any](s lipgloss.Style) Option[T] {
	return func(m *Model[T]) {
		m.activeStyle = &s
		m.customActive = true
	}
}

func WithInactiveStyle[T any](s lipgloss.Style) Option[T] {
	return func(m *Model[T]) {
		m.inactiveStyle = &s
		m.customInactive = true
	}
}

func WithSeparator[T any](sep string) Option[T] {
	return func(m *Model[T]) { m.separator = sep }
}

func WithMaxWidth[T any](width int) Option[T] {
	return func(m *Model[T]) { m.SetMaxWidth(width) }
}

func WithUnderline[T any](enabled bool) Option[T] {
	return func(m *Model[T]) { m.underline = enabled }
}

func WithTitle[T any](title string) Option[T] {
	return func(m *Model[T]) { m.title = title }
}

func (m *Model[T]) SetTitle(title string) {
	m.title = title
}

// WithSlideAccent switches to the animated chevron style.
func WithSlideAccent[T any]() Option[T] {
	return func(m *Model[T]) {
		if m.env.Settings.Animated {
			m.animated = true
			m.separator = chevronSep
		}
	}
}

func New[T any](
	env *ui.Env,
	tabs []Tab[T],
	onSelectTab func(idx int, tab Tab[T]) tea.Cmd,
	opts ...Option[T],
) *Model[T] {
	m := &Model[T]{
		env:         env,
		id:          ui.NewComponentID(),
		tabs:        tabs,
		onSelectTab: onSelectTab,
		separator:   " ",
		animated:    false,
		underline:   true,
	}

	for _, opt := range opts {
		opt(m)
	}

	m.applyDefaultStyles()

	if m.animated {
		m.slide = float64(m.tabOffset(m.idx))
		m.slideFrom = m.slide
		m.slideTo = m.slide
	}

	return m
}

func (m *Model[T]) ID() ui.ComponentID { return m.id }
func (m *Model[T]) Name() string       { return "tabs" }
func (m *Model[T]) Mode() ui.KeyMode   { return ui.KeyModeNormal }
func (m *Model[T]) OnClose() tea.Cmd   { return nil }

func (m *Model[T]) Init() tea.Cmd {
	return m.tick()
}

func (m *Model[T]) CurrentTab() *Tab[T] {
	if m.idx < 0 || m.idx >= len(m.tabs) {
		return nil
	}
	return &m.tabs[m.idx]
}

func (m *Model[T]) SetMaxWidth(width int) {
	if width < 0 {
		width = 0
	}
	m.maxWidth = width
}

// SelectTab sets the active tab by index.
func (m *Model[T]) SelectTab(idx int) *Tab[T] {
	if idx < 0 || idx >= len(m.tabs) {
		return nil
	}
	m.idx = idx
	m.startSlide()
	return &m.tabs[idx]
}

func (m *Model[T]) NextTab() {
	if len(m.tabs) == 0 {
		return
	}
	m.idx = (m.idx + 1) % len(m.tabs)
	m.startSlide()
}

func (m *Model[T]) BackTab() {
	if len(m.tabs) == 0 {
		return
	}
	m.idx--
	if m.idx < 0 {
		m.idx = len(m.tabs) - 1
	}
	m.startSlide()
}

func (m *Model[T]) selectTabCmd() tea.Cmd {
	tab := m.CurrentTab()
	if tab == nil || m.onSelectTab == nil {
		return nil
	}
	return m.onSelectTab(m.idx, *tab)
}

// rebuild refreshes theme-derived default styles. Custom styles are preserved.
func (m *Model[T]) rebuild() {
	if !m.customActive {
		s := m.defaultActiveStyle()
		m.activeStyle = &s
	}
	if !m.customInactive {
		s := m.defaultInactiveStyle()
		m.inactiveStyle = &s
	}
}

func (m *Model[T]) applyDefaultStyles() {
	if m.activeStyle == nil {
		s := m.defaultActiveStyle()
		m.activeStyle = &s
	}
	if m.inactiveStyle == nil {
		s := m.defaultInactiveStyle()
		m.inactiveStyle = &s
	}
}

func (m *Model[T]) styleFor(i int) lipgloss.Style {
	if i == m.idx {
		if m.activeStyle != nil {
			return *m.activeStyle
		}
		return m.defaultActiveStyle()
	}

	if m.inactiveStyle != nil {
		return *m.inactiveStyle
	}
	return m.defaultInactiveStyle()
}

func (m *Model[T]) frameWidth() int {
	width := m.env.Layout.Body().Width
	if m.maxWidth > 0 && width > m.maxWidth {
		width = m.maxWidth
	}
	return width
}
