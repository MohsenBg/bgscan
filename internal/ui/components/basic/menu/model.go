// Package menu contains the selection menus used by bgscan.
package menu

import (
	"time"

	"github.com/MohsenBg/bgscan/internal/ui/shared/env"
	"github.com/MohsenBg/bgscan/internal/ui/shared/layout"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
	"github.com/MohsenBg/bgscan/internal/ui/theme"

	tea "charm.land/bubbletea/v2"
)

const (
	// Maximum menu size.
	maxMenuWidth  = 60
	maxMenuHeight = 25

	// Fallback size when no layout is attached and no fixed size is set.
	fallbackWidth  = 40
	fallbackHeight = 12

	// Minimum width for the sidebar layout.
	minSidebarWidth = 50

	// Lines used by everything except the item rows:
	// title, blank, (items), blank, help.
	chromeRows = 4

	// Blank lines between two item rows.
	rowGap = 1

	// Lines taken by one item row including its gap.
	rowStride = 1 + rowGap

	// Fewest item lines we ever render.
	minVisibleLines = 1
)

// MenuItem is a selectable menu row.
type MenuItem struct {
	icon     string
	title    string
	desc     string
	shortcut string
	action   tea.Cmd
}

// NewMenuItem creates a menu item.
func NewMenuItem(icon, title, shortcut string, action tea.Cmd) MenuItem {
	return MenuItem{icon: icon, title: title, shortcut: shortcut, action: action}
}

// WithDescription sets the item's description.
func (i MenuItem) WithDescription(desc string) MenuItem {
	i.desc = desc
	return i
}

func (i MenuItem) Title() string       { return i.title }
func (i MenuItem) Icon() string        { return i.icon }
func (i MenuItem) Description() string { return i.desc }
func (i MenuItem) Shortcut() string    { return i.shortcut }
func (i MenuItem) Action() tea.Cmd     { return i.action }

// Option configures a menu.
type Option func(*Model)

// WithIcon shows or hides the icon column.
func WithIcon(show bool) Option {
	return func(m *Model) { m.showIcon = show }
}

// WithShortcut shows or hides shortcuts.
func WithShortcut(show bool) Option {
	return func(m *Model) { m.showShortcut = show }
}

// WithWidth sets a fixed menu width.
func WithWidth(width int) Option {
	return func(m *Model) { m.width = width; m.widthAuto = false }
}

// WithHeight sets a fixed menu height.
func WithHeight(height int) Option {
	return func(m *Model) { m.height = height; m.heightAuto = false }
}

// Model is a selectable menu.
type Model struct {
	id       ui.ComponentID
	name     string
	title    string
	items    []MenuItem
	index    int
	onSelect func(MenuItem) tea.Cmd
	Layout   *layout.Layout
	theme    *theme.Theme

	showIcon     bool
	showShortcut bool

	width      int
	height     int
	widthAuto  bool
	heightAuto bool

	// enterAt is when the menu opened. Only the entrance fade-in uses
	// it; selection and activation are instant.
	enterAt time.Time
}

// New creates a menu.
func New(items []MenuItem, title string, deps ui.Deps, options ...Option) *Model {
	m := &Model{
		id:           ui.NewComponentID(),
		name:         "menu",
		title:        title,
		items:        items,
		Layout:       deps.Layout,
		theme:        deps.Theme,
		showIcon:     true,
		showShortcut: true,
		widthAuto:    true,
		heightAuto:   true,
		enterAt:      time.Now(),
	}
	for _, option := range options {
		option(m)
	}
	return m
}

// Init restarts the entrance animation.
func (m *Model) Init() tea.Cmd {
	m.enterAt = time.Now()
	return m.nextFrame()
}

func (m *Model) ID() ui.ComponentID { return m.id }
func (m *Model) Mode() env.Mode     { return env.NormalMode }
func (m *Model) Name() string       { return m.name }
func (m *Model) OnClose() tea.Cmd   { return nil }

// Theme returns the current theme.
func (m *Model) Theme() *theme.Theme { return m.theme }

// SetTheme updates the theme.
func (m *Model) SetTheme(theme *theme.Theme) { m.theme = theme }

// SetOnSelect sets the selection callback.
func (m *Model) SetOnSelect(fn func(MenuItem) tea.Cmd) { m.onSelect = fn }

// GetSelected returns the selected item.
func (m *Model) GetSelected() (MenuItem, bool) {
	if m.index < 0 || m.index >= len(m.items) {
		return MenuItem{}, false
	}
	return m.items[m.index], true
}

// SetItems replaces the menu items and resets the entrance animation.
func (m *Model) SetItems(items []MenuItem) tea.Cmd {
	m.items = items
	if m.index >= len(m.items) {
		m.index = max(len(m.items)-1, 0)
	}
	m.enterAt = time.Now()
	return m.nextFrame()
}

// hasDescriptions reports whether any item has a description.
func (m *Model) hasDescriptions() bool {
	for _, item := range m.items {
		if item.desc != "" {
			return true
		}
	}
	return false
}

// effectiveSize returns the menu's outer size in terminal cells.
// The height is at least the chrome plus one item row.
func (m *Model) effectiveSize() (width, height int) {
	width, height = m.width, m.height
	if m.widthAuto && m.Layout != nil {
		width = min(m.Layout.BodyContentWidth(), maxMenuWidth)
	}
	if m.heightAuto && m.Layout != nil {
		height = min(m.Layout.BodyContentHeight(), maxMenuHeight)
	}
	if width <= 0 {
		width = fallbackWidth
	}
	if height <= 0 {
		height = fallbackHeight
	}
	return width, max(height, chromeRows+minVisibleLines)
}

// visibleLines returns the number of lines available for item rows.
func visibleLines(height int) int {
	return max(height-chromeRows, minVisibleLines)
}

// visibleItems returns how many items fit on one page.
func (m *Model) visibleItems() int {
	_, height := m.effectiveSize()
	return itemsThatFit(visibleLines(height))
}

// itemsThatFit returns how many rows (with gaps between them) fit in lines.
func itemsThatFit(lines int) int {
	return max((lines+rowGap)/rowStride, 1)
}

// itemsHeight returns the lines needed to show n rows with gaps.
func itemsHeight(n int) int {
	if n == 0 {
		return 0
	}
	return n*rowStride - rowGap
}
