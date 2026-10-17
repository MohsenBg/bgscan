package inspector

import (
	"fmt"
	"io"
	"strings"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/help"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/input"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/tabs"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
)

// helpHeight is the vertical space reserved for the short help footer.
const helpHeight = 3

type FieldInput interface {
	input.Dialog
	Value() any
	SetValue(any)
}

// Field is one inspectable property, rendered as one row in the list.
type Field struct {
	Name        string
	Description string
	Group       string
	Input       FieldInput
	Visible     func() bool
	Format      func(any) string

	snapshot *any
}

// value returns the last committed display value. Input.Value() is not read
// here because it reflects live, uncommitted edits.
func (f Field) value() string {
	if f.snapshot == nil {
		return ""
	}
	v := *f.snapshot
	if f.Format != nil {
		return f.Format(v)
	}
	return fmt.Sprint(v)
}

func (f Field) visible() bool {
	return f.Visible == nil || f.Visible()
}

type FieldItem struct {
	Field Field
}

func (i FieldItem) FilterValue() string { return i.Field.Name }

// fieldDelegate renders "Name  value" rows with an optional description line.
type fieldDelegate struct {
	env *ui.Env
}

func (d fieldDelegate) Height() int                             { return 2 }
func (d fieldDelegate) Spacing() int                            { return 0 }
func (d fieldDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d fieldDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(FieldItem)
	if !ok {
		return
	}

	th := d.env.Theme
	f := item.Field

	name := f.Name
	if index == m.Index() {
		name = selectedFieldNameStyle(&th).Render("▶ " + name)
	} else {
		name = fieldNameStyle(&th).Render(name)
	}

	value := valueStyle(&th).Render(f.value())
	gap := max(m.Width()-lipgloss.Width(name)-lipgloss.Width(value), 1)
	line := lipgloss.JoinHorizontal(lipgloss.Top, name, strings.Repeat(" ", gap), value)
	_, _ = fmt.Fprint(w, paddingCell().Render(line))
}

// Model is a tabbed list of fields grouped by Field.Group.
type Model struct {
	id   ui.ComponentID
	name string
	env  *ui.Env

	Title string

	groups map[string][]Field
	order  []string

	tabs ui.Component
	list list.Model
	help ui.Component

	maxWidth       int
	widthOverride  int
	heightOverride int
}

type Option func(*Model)

func WithWidth(width int) Option {
	return func(m *Model) {
		if width > 0 {
			m.widthOverride = width
		}
	}
}

func WithHeight(height int) Option {
	return func(m *Model) {
		if height > 0 {
			m.heightOverride = height
		}
	}
}

// New builds an inspector from a flat field list, grouped by Field.Group
// into tabs. Groups are ordered by first appearance for a stable tab order.
func New(env *ui.Env, name string, fields []Field, opts ...Option) *Model {
	m := &Model{
		id:       ui.NewComponentID(),
		name:     name,
		env:      env,
		maxWidth: 60,
	}
	for _, opt := range opts {
		opt(m)
	}

	for i := range fields {
		f := &fields[i]
		if f.Input == nil {
			continue
		}
		v := f.Input.Value()
		f.snapshot = &v
		f.Input.AppendOnSubmit(func() tea.Cmd {
			m.Refresh()
			return nil
		})
	}

	m.groups = make(map[string][]Field)
	for _, f := range fields {
		if _, seen := m.groups[f.Group]; !seen {
			m.order = append(m.order, f.Group)
		}
		m.groups[f.Group] = append(m.groups[f.Group], f)
	}

	tbs := make([]tabs.Tab[[]Field], 0, len(m.order))
	for _, group := range m.order {
		tbs = append(tbs, tabs.Tab[[]Field]{Label: group, Value: m.groups[group]})
	}

	tb := tabs.New(env, tbs, func(_ int, tab tabs.Tab[[]Field]) tea.Cmd {
		return func() tea.Msg {
			return tabChangeMsg{Group: tab.Label, Fields: tab.Value}
		}
	}, tabs.WithSlideAccent[[]Field]())
	tb.SetMaxWidth(m.maxWidth)
	m.tabs = tb

	if len(tbs) > 0 {
		m.Title = tbs[0].Label
		m.list = newFieldList(tbs[0].Value, m.env, m.Width(), m.Height())
	}

	m.help = help.New(
		env, name,
		help.HelpSet{Short: m.ShortHints(), Long: m.LongHints()},
		help.WithShortHelpWidth(m.Width()),
	)
	return m
}

func newFieldList(fields []Field, env *ui.Env, width, height int) list.Model {
	lm := list.New(visibleItems(fields), fieldDelegate{env: env}, width, height)
	lm.DisableQuitKeybindings()
	lm.SetShowStatusBar(false)
	lm.SetShowTitle(false)
	lm.SetShowHelp(false)
	lm.SetFilteringEnabled(false)
	return lm
}

func visibleItems(fields []Field) []list.Item {
	items := make([]list.Item, 0, len(fields))
	for _, f := range fields {
		if f.visible() {
			items = append(items, FieldItem{Field: f})
		}
	}
	return items
}

func (m *Model) Init() tea.Cmd {
	if m.tabs == nil {
		return nil
	}
	return m.tabs.Init()
}
func (m *Model) ID() ui.ComponentID { return m.id }
func (m *Model) Name() string       { return m.name }
func (m *Model) OnClose() tea.Cmd   { return nil }
func (m *Model) Mode() ui.KeyMode   { return ui.KeyModeNormal }
func (m *Model) Fields() []Field    { return m.groups[m.Title] }
func (m *Model) SelectedField() (Field, bool) {
	item, ok := m.list.SelectedItem().(FieldItem)
	if !ok {
		return Field{}, false
	}
	return item.Field, true
}

func (m *Model) Width() int {
	if m.widthOverride > 0 {
		return m.widthOverride
	}
	return min(m.maxWidth, m.env.Layout.Body().Width)
}

func (m *Model) Height() int {
	total := m.heightOverride
	if total <= 0 {
		total = m.env.Layout.Body().Height
	}

	taken := 4 // vertical padding
	if len(m.groups) > 0 {
		taken += lipgloss.Height(m.tabs.View())
	}
	if m.Title != "" {
		taken += lipgloss.Height(m.Title)
	}
	if m.help != nil {
		taken += helpHeight
	}

	return max(15, total-taken)
}

func (m *Model) SetWidth(width int) {
	m.widthOverride = width
	if tb, ok := m.tabs.(*tabs.Model[[]Field]); ok {
		tb.SetMaxWidth(width)
	}
	m.list.SetWidth(width)
	if h, ok := m.help.(*help.Model); ok {
		h.SetShortWidth(width)
	}
}

func (m *Model) SetHeight(height int) {
	m.heightOverride = height
	m.list.SetHeight(m.Height())
}

func (m *Model) CloseCmd() tea.Cmd {
	return func() tea.Msg {
		return ui.CloseComponentMsg{ID: m.ID()}
	}
}

// Refresh re-pulls Value() from every Input and re-renders.
func (m *Model) Refresh() tea.Cmd {
	for _, f := range m.Fields() {
		*f.snapshot = f.Input.Value()
	}
	return m.list.SetItems(visibleItems(m.groups[m.Title]))
}
