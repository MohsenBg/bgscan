// Package menu contains the selection menus used by bgscan.
package menu

import (
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/MohsenBg/bgscan/internal/ui/components/basic/help"
	"github.com/MohsenBg/bgscan/internal/ui/shared/keys"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
)

const (
	maxMenuWidth    = 65
	maxMenuHeight   = 25
	fallbackWidth   = 40
	fallbackHeight  = 12
	minSidebarWidth = 40
	minMenuWidth    = 15
	minMenuHeight   = 12

	helpRows        = 1
	rowGap          = 1
	rowStride       = 1 + rowGap
	minVisibleLines = 1

	contentPadding = 0

	iconColumnWidth         = 3
	keycapHorizontalPadding = 2
	shortcutGapWidth        = 1
	helpSeparatorWidth      = 3

	rowPrefixWidth   = 1
	rowLeadWidth     = 1
	rowSuffixWidth   = 1
	rowOverheadWidth = rowPrefixWidth + rowSuffixWidth
	rowChromeWidth   = rowOverheadWidth + rowLeadWidth
	minRowWidth      = 10
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
	return MenuItem{
		icon:     icon,
		title:    title,
		shortcut: shortcut,
		action:   action,
	}
}

// WithDescription sets the item description.
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
	return func(m *Model) {
		m.showIcon = show
	}
}

// WithShortcut shows or hides shortcuts.
func WithShortcut(show bool) Option {
	return func(m *Model) {
		m.showShortcut = show
	}
}

// WithWidth sets a fixed menu width.
func WithWidth(width int) Option {
	return func(m *Model) {
		m.width = max(1, width)
	}
}

// WithHeight sets a fixed menu height.
func WithHeight(height int) Option {
	return func(m *Model) {
		m.height = max(1, height)
	}
}

// WithHelp shows or hides the built-in help bar.
// Disable it when a parent, such as a dialog frame, owns the help.
func WithHelp(show bool) Option {
	return func(m *Model) {
		m.showHelp = show
	}
}

// WithBackHint controls whether the built-in back hint is shown.
// Disable it when a parent, such as a dialog frame, provides its own back hint.
func WithBackHint(show bool) Option {
	return func(m *Model) {
		m.showBackHint = show
	}
}

// Model is a selectable menu.
type Model struct {
	id       ui.ComponentID
	name     string
	title    string
	items    []MenuItem
	index    int
	onSelect func(MenuItem) tea.Cmd

	env *ui.Env

	showIcon     bool
	showShortcut bool
	showHelp     bool
	showBackHint bool

	width  int
	height int

	help ui.Component

	// enterAt is used only for the entrance animation.
	// Selection itself is instant.
	enterAt time.Time
}

// New creates a menu.
func New(items []MenuItem, title string, env *ui.Env, options ...Option) *Model {
	m := &Model{
		id:           ui.NewComponentID(),
		name:         "menu",
		title:        title,
		items:        items,
		env:          env,
		showIcon:     true,
		showShortcut: true,
		showHelp:     true,
		showBackHint: true,
		enterAt:      time.Now(),
	}

	for _, option := range options {
		option(m)
	}

	if m.showHelp {
		w, _ := m.effectiveSize()
		innerWidth := max(1, w-2*contentPadding)

		m.help = help.New(
			env,
			title,
			help.HelpSet{
				Short: m.ShortHints(),
				Long:  m.LongHints(),
			},
			help.WithShortHelpWidth(innerWidth),
			help.WithShortHelpHeight(helpRows),
		)
	}

	return m
}

// Init restarts the entrance animation.
func (m *Model) Init() tea.Cmd {
	m.enterAt = time.Now()
	return m.nextFrame()
}

func (m *Model) ID() ui.ComponentID { return m.id }
func (m *Model) Mode() ui.KeyMode   { return ui.KeyModeNormal }
func (m *Model) Name() string       { return m.name }
func (m *Model) OnClose() tea.Cmd   { return nil }

// Title returns the menu heading.
func (m *Model) Title() string { return m.title }

// SetOnSelect sets the selection callback.
func (m *Model) SetOnSelect(fn func(MenuItem) tea.Cmd) {
	m.onSelect = fn
}

// GetSelected returns the selected item.
func (m *Model) GetSelected() (MenuItem, bool) {
	if m.index < 0 || m.index >= len(m.items) {
		return MenuItem{}, false
	}

	return m.items[m.index], true
}

// SetItems replaces the menu items and restarts the entrance animation.
func (m *Model) SetItems(items []MenuItem) tea.Cmd {
	m.items = items

	if m.index >= len(m.items) {
		m.index = max(len(m.items)-1, 0)
	}

	m.enterAt = time.Now()
	return m.nextFrame()
}

func (m *Model) hasDescriptions() bool {
	for _, item := range m.items {
		if item.desc != "" {
			return true
		}
	}

	return false
}

// effectiveSize returns the total menu size including outer padding.
func (m *Model) effectiveSize() (width, height int) {
	availWidth := fallbackWidth
	availHeight := fallbackHeight

	if w := m.env.Layout.Body().Width; w > 0 {
		if m.width > 0 {
			availWidth = m.width
		} else {
			availWidth = w
		}
	}

	if h := m.env.Layout.Body().Height; h > 0 && m.height <= 0 {
		if m.height > 0 {
			availHeight = m.height
		} else {
			availHeight = h
		}
	}

	width = min(availWidth, maxMenuWidth)
	width = max(width, minMenuWidth, 2*contentPadding+1)

	desiredHeight := max(m.chromeRows()+itemsHeight(len(m.items)), minMenuHeight)

	height = min(desiredHeight, maxMenuHeight, availHeight)
	// height = max(availHeight, m.chromeRows()+minVisibleLines)

	return width, height
}

// chromeRows counts non-item rows: title, spacers, and help.
func (m *Model) chromeRows() int {
	rows := 2 // head spacer + foot spacer

	if m.title != "" {
		rows++
	}

	if m.showHelp {
		rows += helpRows
	}

	return rows
}

func (m *Model) visibleLines(height int) int {
	return max(height-m.chromeRows(), minVisibleLines)
}

func (m *Model) visibleItems() int {
	_, height := m.effectiveSize()
	return itemsThatFit(m.visibleLines(height))
}

func itemsThatFit(lines int) int {
	return max((lines+rowGap)/rowStride, 1)
}

func itemsHeight(n int) int {
	if n == 0 {
		return 0
	}

	return n*rowStride - rowGap
}

func (m *Model) handleResize() {
	if m.help == nil {
		return
	}

	h, ok := m.help.(*help.Model)
	if !ok {
		return
	}

	w, _ := m.effectiveSize()
	h.SetShortWidth(max(1, w-2*contentPadding))
}

// ShortHints are the compact hints shown under the menu.
func (m *Model) ShortHints() keys.Hints {
	k := m.env.Keys.Menu
	g := m.env.Keys.Global

	hints := keys.Hints{
		keys.Combine("move", k.Up, k.Down),
	}

	back := g.Back
	back.Hidden = !m.showBackHint
	back.Disabled = !m.showBackHint

	hints = append(hints, keys.Short(
		k.Select,
		back,
		g.Quit,
	)...)

	if m.showHelp {
		hints = append(hints, keys.Short(g.Help)...)
	}

	return hints
}

// LongHints are everything the menu handles, for the full help dialog.
func (m *Model) LongHints() keys.Hints {
	k := m.env.Keys.Menu
	g := m.env.Keys.Global

	back := g.Back
	back.Hidden = !m.showBackHint
	back.Disabled = !m.showBackHint

	hints := keys.Long(
		k.Up,
		k.Down,
		k.Select,
		back,
	)

	for _, item := range m.items {
		hints = append(hints, keys.Hint{
			Key:   item.shortcut,
			Label: "open " + item.title,
		})
	}

	hints = append(hints, keys.Long(g.Quit, g.Help)...)

	return hints
}

func DialogHints(k keys.MenuKeys) keys.Hints {
	hints := keys.Hints{
		keys.Combine("move", k.Up, k.Down),
	}

	hints = append(hints, keys.Short(
		k.Select,
	)...)

	return hints
}
