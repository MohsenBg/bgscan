package menu

import (
	"fmt"
	"io"
	"strings"

	"bgscan/internal/logger"
	"bgscan/internal/ui/shared/env"
	"bgscan/internal/ui/shared/layout"
	"bgscan/internal/ui/shared/ui"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// MenuItem represents a menu item that implements the list.Item interface.
type MenuItem struct {
	icon     string
	title    string
	shortcut string
	action   func() tea.Cmd
}

func (i MenuItem) FilterValue() string    { return i.title }
func (i MenuItem) Title() string          { return i.title }
func (i MenuItem) Icon() string           { return i.icon }
func (i MenuItem) Shortcut() string       { return i.shortcut }
func (i MenuItem) Action() func() tea.Cmd { return i.action }

// NewMenuItem creates a new menu item.
func NewMenuItem(
	icon string,
	title string,
	shortcut string,
	action tea.Cmd,
) MenuItem {
	return MenuItem{
		icon:     icon,
		title:    title,
		shortcut: shortcut,
		action:   func() tea.Cmd { return action },
	}
}

// ItemDelegate handles rendering of menu items.
type ItemDelegate struct {
	showIcon     bool
	showShortcut bool
}

func NewItemDelegate(showIcon, showShortcut bool) ItemDelegate {
	return ItemDelegate{
		showIcon:     showIcon,
		showShortcut: showShortcut,
	}
}

func (d ItemDelegate) Height() int  { return 2 }
func (d ItemDelegate) Spacing() int { return 0 }

func (d ItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd {
	return nil
}

func (d ItemDelegate) Render(
	w io.Writer,
	m list.Model,
	index int,
	listItem list.Item,
) {
	item, ok := listItem.(MenuItem)
	if !ok {
		return
	}

	selected := index == m.Index()

	var left string

	// Icon column.
	if d.showIcon {
		if selected {
			left += selectedIconStyle().Render(item.icon)
		} else {
			left += iconStyle().Render(item.icon)
		}
	}

	// Title.
	if selected {
		left += selectedItemTitleStyle().Render(item.title)
	} else {
		left += itemTitleStyle().Render(item.title)
	}

	// Shortcut.
	var right string
	if d.showShortcut && item.shortcut != "" {
		right = shortcutStyle().Render(item.shortcut)
	}

	// Keep the shortcut aligned to the right edge.
	gap := max(
		m.Width()-lipgloss.Width(left)-lipgloss.Width(right),
		1,
	)

	line := lipgloss.JoinHorizontal(
		lipgloss.Top,
		left,
		strings.Repeat(" ", gap),
		right,
	)

	_, err := fmt.Fprint(w, PaddingCell().Render(line))
	if err != nil {
		logger.UIError("Error while rendering menu: %v", err)
	}
}

// Option configures a menu.
type Option func(*Model)

// WithIcon enables or disables the icon column.
func WithIcon(enabled bool) Option {
	return func(m *Model) {
		m.showIcon = enabled
	}
}

// WithShortcut enables or disables the shortcut column.
func WithShortcut(enabled bool) Option {
	return func(m *Model) {
		m.showShortcut = enabled
	}
}

// WithWidth sets the menu width.
//
// A value <= 0 keeps the automatically calculated width.
func WithWidth(width int) Option {
	return func(m *Model) {
		if width > 0 {
			m.width = width
			m.widthAuto = false
		}
	}
}

// WithHeight sets the menu height.
//
// A value <= 0 keeps the automatically calculated height.
func WithHeight(height int) Option {
	return func(m *Model) {
		if height > 0 {
			m.height = height
			m.heightAuto = false
		}
	}
}

// Model represents the menu component state.
type Model struct {
	id       ui.ComponentID
	name     string
	List     list.Model
	onSelect func(MenuItem) tea.Cmd
	Layout   *layout.Layout
	items    []MenuItem

	showIcon     bool
	showShortcut bool

	width  int
	height int

	widthAuto  bool
	heightAuto bool
}

func New(
	items []MenuItem,
	title string,
	layout *layout.Layout,
	options ...Option,
) *Model {
	m := &Model{
		id:           ui.NewComponentID(),
		name:         "menu",
		items:        items,
		Layout:       layout,
		showIcon:     true,
		showShortcut: true,
		width:        layout.BodyContentWidth(),
		height:       layout.BodyContentHeight(),
		widthAuto:    true,
		heightAuto:   true,
	}

	// Apply options before creating the list delegate.
	for _, option := range options {
		option(m)
	}

	listItems := make([]list.Item, len(items))
	for i, item := range items {
		listItems[i] = item
	}

	delegate := NewItemDelegate(
		m.showIcon,
		m.showShortcut,
	)

	m.List = list.New(
		listItems,
		delegate,
		m.width,
		m.height,
	)

	m.List.Title = title
	m.List.Styles.Title = titleStyle()

	m.List.SetShowStatusBar(false)
	m.List.SetFilteringEnabled(false)
	m.List.SetShowHelp(true)
	m.List.DisableQuitKeybindings()

	m.List.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(
				key.WithKeys(env.KeyEnter),
				key.WithHelp(env.KeyEnter, "select"),
			),
		}
	}

	m.List.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(
				key.WithKeys(env.KeyEnter),
				key.WithHelp(env.KeyEnter, "select"),
			),
		}
	}

	m.updateMenuLayout()

	return m
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

func (m *Model) OnClose() tea.Cmd {
	return nil
}

func (m *Model) SetOnSelect(fn func(MenuItem) tea.Cmd) {
	m.onSelect = fn
}

func (m *Model) GetSelected() (MenuItem, bool) {
	item, ok := m.List.SelectedItem().(MenuItem)
	return item, ok
}

func (m *Model) SetItems(items []MenuItem) tea.Cmd {
	listItems := make([]list.Item, len(items))

	for i, item := range items {
		listItems[i] = item
	}

	m.items = items

	return m.List.SetItems(listItems)
}

func (m *Model) Mode() env.Mode {
	return env.NormalMode
}
