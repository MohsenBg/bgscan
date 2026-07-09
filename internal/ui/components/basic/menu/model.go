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

func NewMenuItem(icon, title, shortcut string, action tea.Cmd) MenuItem {
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

func (d ItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d ItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(MenuItem)
	if !ok {
		return
	}

	var leftSection, rightSection string

	titleText := item.title
	if index == m.Index() {
		leftSection += selectedIconStyle().Render(item.icon)
		titleText = selectedItemTitleStyle().Render(titleText)
	} else {
		leftSection += iconStyle().Render(item.icon)
		titleText = itemTitleStyle().Render(titleText)
	}
	leftSection += titleText

	if d.showShortcut && item.shortcut != "" {
		rightSection += shortcutStyle().Render(item.shortcut)
	}

	gap := max(m.Width()-lipgloss.Width(leftSection)-lipgloss.Width(rightSection), 1)

	line := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftSection,
		strings.Repeat(" ", gap),
		rightSection,
	)

	_, err := fmt.Fprint(w, PaddingCell().Render(line))
	if err != nil {
		logger.UIError("Error while rendering menu: %v", err)
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

type KeyMap struct {
	ExecuteShortcut string // was tea.KeyMsg
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		ExecuteShortcut: "",
	}
}

func New(items []MenuItem, title string, layout *layout.Layout) *Model {
	listItems := make([]list.Item, len(items))
	for i, item := range items {
		listItems[i] = item
	}

	width := min(layout.BodyContentWidth(), 50)
	height := min(layout.BodyContentHeight(), 20)

	delegate := NewItemDelegate(true, true)
	l := list.New(listItems, delegate, width, height)

	l.Title = title
	l.Styles.Title = titleStyle()

	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(true)
	l.SetFilteringEnabled(false)
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(
				key.WithKeys(env.KeyEnter),
				key.WithHelp(env.KeyEnter, "select"),
			),
		}
	}

	l.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(
				key.WithKeys(env.KeyEnter),
				key.WithHelp(env.KeyEnter, "select"),
			),
		}
	}

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

func (m *Model) Init() tea.Cmd      { return nil }
func (m *Model) ID() ui.ComponentID { return m.id }
func (m *Model) Name() string       { return m.name }
func (m *Model) OnClose() tea.Cmd   { return nil }

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
	return m.List.SetItems(listItems)
}

func (m *Model) Mode() env.Mode {
	return env.NormalMode
}
