package table

import (
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	"bgscan/internal/ui/shared/env"
	"bgscan/internal/ui/shared/layout"
	"bgscan/internal/ui/shared/ui"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// Aliases for upstream table types
type (
	Column = table.Column
	Row    = table.Row
)

// tableHPad must match the horizontal padding in tableViewStyle (Padding(0, 1, 0, 1) = 2 chars).
const tableHPad = 4

// Model wraps a Bubble Tea table with responsive layout, key bindings, and concurrency safety.
type Model struct {
	mu sync.RWMutex

	id     ui.ComponentID
	name   string
	Title  string
	Layout *layout.Layout

	Help     help.Model
	FullHelp bool

	BubbleTable table.Model
	Keys        KeyMap

	originalCols []table.Column
	paddingY     int
	maxWidth     int // 0 means unlimited

	// Pending fields for Functional Options Pattern
	pendingCols []table.Column
	pendingRows []table.Row
}

// --- Functional Options Pattern ---

// Option configures the table Model.
type Option func(*Model)

// WithTitle sets the table title.
func WithTitle(title string) Option {
	return func(m *Model) { m.Title = title }
}

// WithColumns sets the initial columns.
func WithColumns(cols []table.Column) Option {
	return func(m *Model) { m.pendingCols = cols }
}

// WithRows sets the initial rows.
func WithRows(rows []table.Row) Option {
	return func(m *Model) { m.pendingRows = rows }
}

// WithPaddingY sets the vertical padding.
func WithPaddingY(padding int) Option {
	return func(m *Model) { m.paddingY = padding }
}

// WithMaxWidth sets an optional maximum width for the table (0 = unlimited).
func WithMaxWidth(w int) Option {
	return func(m *Model) { m.maxWidth = w }
}

// WithKeyBindings appends custom key bindings to the default ones.
func WithKeyBindings(keys ...ActionKey) Option {
	return func(m *Model) { m.Keys = defaultKeys(keys...) }
}

// New creates a new table model using the Functional Options Pattern.
func New(lay *layout.Layout, opts ...Option) *Model {
	m := &Model{
		id:     ui.NewComponentID(),
		name:   "table",
		Layout: lay,
		Help:   help.New(),
		Keys:   defaultKeys(),
	}

	// Apply options
	for _, opt := range opts {
		opt(m)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Fallback to empty slices if not provided
	cols := m.pendingCols
	if cols == nil {
		cols = []table.Column{}
	}
	rows := m.pendingRows
	if rows == nil {
		rows = []table.Row{}
	}

	m.originalCols = slices.Clone(cols)
	m.BubbleTable = table.New(
		table.WithColumns(cols), // original widths; real scaling happens on first WindowSizeMsg
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(max(1, len(rows))),
		table.WithWidth(m.tableWidthLocked()),
	)
	m.BubbleTable.SetStyles(tableStyles())
	m.updateTableSizeLocked()

	return m
}

// --- Component Interface ---

func (m *Model) Init() tea.Cmd      { return nil }
func (m *Model) ID() ui.ComponentID { return m.id }
func (m *Model) Name() string       { return m.name }
func (m *Model) OnClose() tea.Cmd   { return nil }
func (m *Model) Mode() env.Mode     { return env.NormalMode }

// --- Public Mutators (Thread-safe) ---

func (m *Model) SetPaddingY(padding int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.paddingY = padding
	m.updateTableSizeLocked()
}

func (m *Model) SetMaxWidth(w int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.maxWidth = w
	m.updateTableSizeLocked()
}

func (m *Model) AppendRow(row table.Row) {
	m.mu.Lock()
	defer m.mu.Unlock()
	rows := append(slices.Clone(m.BubbleTable.Rows()), slices.Clone(row))
	m.BubbleTable.SetRows(rows)
	m.updateTableSizeLocked()
}

func (m *Model) SetRows(rows []table.Row) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.BubbleTable.SetRows(cloneRows(rows))
	m.updateTableSizeLocked()
}

// --- Helper Functions ---

func NewRowTime(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.Format("2006-01-02 15:04:05")
}

func cloneRows(rows []table.Row) []table.Row {
	out := make([]table.Row, len(rows))
	for i, row := range rows {
		out[i] = slices.Clone(row)
	}
	return out
}

// scaledColumns distributes the available content width across columns
// proportionally to their original widths, accounting for the bubbles table's
// internal per-cell padding (Padding(0,1) each side = 2 chars/col) and
// inter-column borders (numCols-1 chars).
func (m *Model) scaledColumns() []table.Column {
	src := m.originalCols
	if len(src) == 0 {
		return nil
	}
	out := slices.Clone(src)

	total := 0
	for _, c := range src {
		total += c.Width
	}
	if total <= 0 {
		return out
	}

	nCols := len(src)
	// Each column rendered width = colWidth + 2 (cell padding both sides).
	// Separators between columns = nCols-1.
	// So: rendered = sum(colWidths) + 2*nCols + (nCols-1)
	// We want: sum(colWidths) = tableW - 2*nCols - (nCols-1)
	overhead := 2*nCols + (nCols - 1)
	target := max(nCols, m.tableWidthLocked()-overhead)

	assigned := 0
	for i, c := range src {
		w := int(float64(target) * float64(c.Width) / float64(total))
		out[i].Width = max(1, w)
		assigned += out[i].Width
	}
	// Give rounding remainder to last column.
	if diff := target - assigned; diff != 0 {
		out[len(out)-1].Width = max(1, out[len(out)-1].Width+diff)
	}
	return out
}

func (m *Model) updateTableSizeLocked() {
	if m.Layout == nil || m.Layout.Body.Height == 0 || m.Layout.Body.Width == 0 {
		return
	}
	if len(m.originalCols) == 0 {
		return
	}

	scaled := m.scaledColumns()
	m.BubbleTable.SetColumns(scaled)

	// Calculate available height
	helpHeight := lipgloss.Height(m.renderHelpView())
	titleHeight := lipgloss.Height(m.renderTitle())
	height := max(1, m.Layout.Body.Height-helpHeight-titleHeight-m.paddingY)

	m.BubbleTable.SetHeight(height)
	m.BubbleTable.SetWidth(m.tableWidthLocked())
}

func (m *Model) tableWidthLocked() int {
	if m.Layout == nil || m.Layout.Body.Width == 0 {
		return 78 // sensible fallback matching default 80-col terminal minus tableHPad
	}
	w := max(10, m.Layout.Body.Width-tableHPad)
	if m.maxWidth > 0 {
		w = min(w, m.maxWidth)
	}
	return w
}

// --- Key Bindings ---

type ActionKey struct {
	Keys      []string
	ShortHelp string
	FullHelp  string
	Cmd       tea.Cmd
}

var arrowSymbols = map[string]string{
	"up": "↑", "down": "↓", "left": "←", "right": "→",
}

func NewKey(keys []string, shortHelp, fullHelp string, cmd tea.Cmd) ActionKey {
	ks := make([]string, len(keys))
	for i, k := range keys {
		if symbol, ok := arrowSymbols[k]; ok {
			ks[i] = symbol
		} else {
			ks[i] = k
		}
	}

	kstr := strings.Join(ks, "/")
	if kstr == "" {
		kstr = "?"
	}
	if shortHelp != "" {
		shortHelp = fmt.Sprintf("%s %s", kstr, shortHelp)
	}
	if fullHelp != "" {
		fullHelp = fmt.Sprintf("%s %s", kstr, fullHelp)
	}

	return ActionKey{Keys: keys, ShortHelp: shortHelp, FullHelp: fullHelp, Cmd: cmd}
}

type KeyMap struct{ Actions []ActionKey }

func (k *KeyMap) Add(a ActionKey) { k.Actions = append(k.Actions, a) }

func (k KeyMap) Check(msg tea.Msg) tea.Cmd {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return nil
	}
	keyStr := keyMsg.String()
	for _, a := range k.Actions {
		if slices.Contains(a.Keys, keyStr) {
			return a.Cmd
		}
	}
	return nil
}

func (k KeyMap) ShortHelp() []key.Binding {
	var bindings []key.Binding
	for _, a := range k.Actions {
		if a.ShortHelp == "" {
			continue
		}
		bindings = append(bindings, key.NewBinding(
			key.WithKeys(a.Keys...),
			key.WithHelp(a.ShortHelp, ""),
		))
	}
	return bindings
}

func (k KeyMap) FullHelp(width int) [][]key.Binding {
	if len(k.Actions) == 0 {
		return nil
	}
	colCount := 3
	if width > 90 {
		colCount = 4
	}

	cols := make([][]key.Binding, colCount)
	for i, a := range k.Actions {
		if a.FullHelp == "" {
			continue
		}
		binding := key.NewBinding(
			key.WithKeys(a.Keys...),
			key.WithHelp("", a.FullHelp),
		)
		cols[i%colCount] = append(cols[i%colCount], binding)
	}
	return cols
}

func defaultKeys(extra ...ActionKey) KeyMap {
	km := KeyMap{}
	km.Add(NewKey([]string{"up", "k"}, "up", "Move up", nil))
	km.Add(NewKey([]string{"down", "j"}, "down", "Move down", nil))
	km.Add(NewKey([]string{"b", "pgup"}, "", "Page up", nil))
	km.Add(NewKey([]string{"f", "pgdown", " "}, "", "Page down", nil))
	km.Add(NewKey([]string{"u", "ctrl+u"}, "", "½ page up", nil))
	km.Add(NewKey([]string{"d", "ctrl+d"}, "", "½ page down", nil))
	km.Add(NewKey([]string{"home", "g"}, "", "Go to start", nil))
	km.Add(NewKey([]string{"end", "G"}, "", "Go to end", nil))

	for _, k := range extra {
		km.Add(k)
	}

	km.Add(NewKey([]string{"?"}, "help", "Toggle help", nil))
	km.Add(NewKey([]string{"q", "esc"}, "quit", "Quit", nil))
	return km
}

// SetKeys replaces the current key bindings with the provided action keys.
// It merges them with the default navigation keys.
func (m *Model) SetKeys(keys ...ActionKey) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Keys = defaultKeys(keys...)
}
