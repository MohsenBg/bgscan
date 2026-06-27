package menu

import (
	"fmt"
	"io"
	"strings"

	"bgscan/internal/logger"
	"bgscan/internal/ui/shared/env"
	"bgscan/internal/ui/shared/layout"
	"bgscan/internal/ui/shared/ui"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// MenuItem represents a menu item that implements the list.Item interface.
// It contains display information and an optional action to execute.
type MenuItem struct {
	icon     string
	title    string
	shortcut string
	action   func() tea.Cmd
}

// FilterValue implements list.Item interface for filtering functionality.
func (i MenuItem) FilterValue() string { return i.title }

// Title returns the display title of the menu item.
func (i MenuItem) Title() string { return i.title }

// Icon returns the icon/emoji displayed before the title.
func (i MenuItem) Icon() string { return i.icon }

// Shortcut returns the keyboard shortcut key for this item.
func (i MenuItem) Shortcut() string { return i.shortcut }

// Action returns the command to execute when this item is selected.
func (i MenuItem) Action() func() tea.Cmd { return i.action }

// NewMenuItem creates a new menu item with the specified properties.
func NewMenuItem(icon, title, shortcut string, action tea.Cmd) MenuItem {
	return MenuItem{
		icon:     icon,
		title:    title,
		shortcut: shortcut,
		action:   func() tea.Cmd { return action },
	}
}

// ItemDelegate handles the rendering of menu items in the list.
type ItemDelegate struct {
	showIcon     bool
	showShortcut bool
}

// NewItemDelegate creates a new delegate with display options.
func NewItemDelegate(showIcon, showShortcut bool) ItemDelegate {
	return ItemDelegate{
		showIcon:     showIcon,
		showShortcut: showShortcut,
	}
}

// Height returns the height of each item in terminal lines.
func (d ItemDelegate) Height() int { return 2 }

// Spacing returns the number of lines between items.
func (d ItemDelegate) Spacing() int { return 0 }

// Update handles messages for the delegate (currently unused).
func (d ItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

// Render draws a single menu item to the writer.
func (d ItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(MenuItem)
	if !ok {
		return
	}

	var leftSection, rightSection string

	// Add title with appropriate style based on selection
	titleText := item.title
	if index == m.Index() {
		leftSection += selectedIconStyle().Render(item.icon)
		titleText = selectedItemTitleStyle().Render(titleText)
	} else {
		leftSection += iconStyle().Render(item.icon)
		titleText = itemTitleStyle().Render(titleText)
	}
	leftSection += titleText

	// Add shortcut if enabled
	if d.showShortcut && item.shortcut != "" {
		rightSection += shortcutStyle().Render(item.shortcut)
	}

	// Calculate gap to space out left and right sections
	gap := max(m.Width()-lipgloss.Width(leftSection)-lipgloss.Width(rightSection), 1)

	// Join sections with spacing
	line := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftSection,
		strings.Repeat(" ", gap),
		rightSection,
	)

	if _, err := fmt.Fprint(w, PaddingCell().Render(line)); err != nil {
		logger.UIError("error rendering menu item %v", err)
	}
}

// Model represents the menu component state.
type Model struct {
	id       ui.ComponentID
	name     string
	List     list.Model
	onSelect func(MenuItem) tea.Cmd
	Layout   *layout.Layout
	keyMap   KeyMap
	items    []MenuItem
}

// KeyMap defines keyboard shortcuts for the menu.
type KeyMap struct {
	ExecuteShortcut tea.KeyMsg
}

// DefaultKeyMap returns the default key bindings for the menu.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		ExecuteShortcut: tea.KeyMsg{Type: tea.KeyRunes},
	}
}

// New creates a new menu model with the given items and dimensions.
func New(items []MenuItem, title string, layout *layout.Layout) *Model {
	// Convert MenuItem slice to list.Item slice
	listItems := make([]list.Item, len(items))
	for i, item := range items {
		listItems[i] = item
	}

	width := min(layout.BodyContentWidth(), 50)
	height := min(layout.BodyContentHeight(), 20)

	// Create list with custom delegate
	delegate := NewItemDelegate(true, true)
	l := list.New(listItems, delegate, width, height)

	// Configure list appearance
	l.Title = title
	l.Styles.Title = titleStyle()

	// Configure list behavior
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(true)

	m := &Model{
		id:     ui.NewComponentID(),
		name:   "menu",
		List:   l,
		items:  items,
		keyMap: DefaultKeyMap(),
		Layout: layout,
	}
	m.updateMenuLayout()
	return m
}

// Init initializes the menu component.
func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) ID() ui.ComponentID {
	return m.id
}

func (m *Model) Name() string {
	return m.name
}

func (m *Model) OnClose() tea.Cmd {
	return nil
}

// SetOnSelect sets the callback function to execute when an item is selected.
func (m *Model) SetOnSelect(fn func(MenuItem) tea.Cmd) {
	m.onSelect = fn
}

// GetSelected returns the currently selected menu item.
func (m *Model) GetSelected() (MenuItem, bool) {
	item, ok := m.List.SelectedItem().(MenuItem)
	return item, ok
}

// SetItems updates the menu with a new set of items.
func (m *Model) SetItems(items []MenuItem) tea.Cmd {
	listItems := make([]list.Item, len(items))
	for i, item := range items {
		listItems[i] = item
	}
	return m.List.SetItems(listItems)
}

func (m *Model) Mode() env.Mode {
	return env.NormalMode
}
