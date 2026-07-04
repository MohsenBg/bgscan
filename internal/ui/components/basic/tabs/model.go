package tabs

import (
	"bgscan/internal/ui/shared/env"
	"bgscan/internal/ui/shared/layout"
	"bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

type Tab[T any] struct {
	Label string
	Value T
}

func NewTab[T any](label string, value T) Tab[T] {
	return Tab[T]{
		Label: label,
		Value: value,
	}
}

type Model[T any] struct {
	layout      *layout.Layout
	id          ui.ComponentID
	name        string
	tabs        []Tab[T]
	onSelectTab func(idx int, tab Tab[T]) tea.Cmd
	idx         int
	maxWidth    int
}

func New[T any](layout *layout.Layout, tabs []Tab[T], onSelectTab func(idx int, tab Tab[T]) tea.Cmd) *Model[T] {
	return &Model[T]{
		layout:      layout,
		id:          ui.NewComponentID(),
		name:        "tabs",
		tabs:        tabs,
		idx:         0,
		maxWidth:    90,
		onSelectTab: onSelectTab,
	}
}

func (m *Model[T]) SetMaxWidth(width int) {
	m.maxWidth = width
}

// Mode implements
func (m *Model[T]) Mode() env.Mode {
	return env.NormalMode
}

// Init implements the BubbleTea initialization interface.
func (m *Model[T]) Init() tea.Cmd {
	return nil
}

// ID returns the component unique identifier.
func (m *Model[T]) ID() ui.ComponentID {
	return m.id
}

// Name returns the component name.
func (m *Model[T]) Name() string {
	return m.name
}

// OnClose is called when the component is removed from the UI.
func (m *Model[T]) OnClose() tea.Cmd {
	return nil
}

func (m *Model[T]) SelectTab(idx int) *Tab[T] {
	if idx >= 0 && len(m.tabs) > idx {
		m.idx = idx
		return &m.tabs[idx]
	}
	return nil
}

func (m *Model[T]) CurrentTab() *Tab[T] {
	if m.idx >= 0 && len(m.tabs) > m.idx {
		return &m.tabs[m.idx]
	}
	return nil
}

func (m *Model[T]) selectTabCmd() tea.Cmd {
	tab := m.CurrentTab()
	if tab != nil {
		return m.onSelectTab(m.idx, *tab)
	}
	return nil
}

func (m *Model[T]) NextTab() {
	if m.idx+1 < len(m.tabs) {
		m.idx++
		return
	}
	m.idx = 0
}

func (m *Model[T]) BackTab() {
	if m.idx-1 < 0 {
		m.idx = len(m.tabs) - 1
		return
	}
	m.idx--
}
