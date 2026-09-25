package table

import (
	"cmp"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/MohsenBg/bgscan/internal/logger"
	"github.com/MohsenBg/bgscan/internal/ui/shared/env"
	"github.com/MohsenBg/bgscan/internal/ui/shared/layout"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
	"github.com/MohsenBg/bgscan/internal/ui/theme"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/dustin/go-humanize"
)

type (
	Column = table.Column
	Row    = table.Row
)

// tableHPad must match the horizontal padding in tableViewStyle
// (Padding(0, 1, 0, 1) = 2 chars).
const tableHPad = 4

// Model is a Bubble Tea table with responsive layout, key bindings, sorting
// and filtering. All methods are safe for concurrent use.
type Model struct {
	mu sync.RWMutex

	id     ui.ComponentID
	name   string
	Title  string
	Layout *layout.Layout
	theme  *theme.Theme
	log    logger.Set

	Help     help.Model
	FullHelp bool

	BubbleTable table.Model
	Keys        KeyMap

	originalCols []table.Column
	paddingY     int
	maxWidth     int // 0 means unlimited
	maxHeight    int // 0 means unlimited

	// sortCol indexes originalCols; -1 means load order.
	sortCol      int
	sortAsc      bool
	unsortable   map[int]bool
	originalRows []Row

	filtering bool
	filter    string

	pendingCols []table.Column
	pendingRows []table.Row
}

type Option func(*Model)

func WithTitle(title string) Option {
	return func(m *Model) { m.Title = title }
}

func WithColumns(cols []table.Column) Option {
	return func(m *Model) { m.pendingCols = cols }
}

func WithRows(rows []table.Row) Option {
	return func(m *Model) { m.pendingRows = rows }
}

func WithPaddingY(padding int) Option {
	return func(m *Model) { m.paddingY = padding }
}

// WithMaxWidth caps the width (0 means unlimited).
func WithMaxWidth(w int) Option {
	return func(m *Model) { m.maxWidth = w }
}

// WithMaxHeight caps the height (0 means unlimited).
func WithMaxHeight(h int) Option {
	return func(m *Model) { m.maxHeight = h }
}

func WithKeyBindings(keys ...ActionKey) Option {
	return func(m *Model) { m.Keys = defaultKeys(keys...) }
}

// WithUnsortableColumns excludes columns from sort cycling and SetSort.
func WithUnsortableColumns(cols ...int) Option {
	return func(m *Model) {
		if m.unsortable == nil {
			m.unsortable = map[int]bool{}
		}
		for _, c := range cols {
			m.unsortable[c] = true
		}
	}
}

func New(deps ui.Deps, opts ...Option) *Model {
	m := &Model{
		id:      ui.NewComponentID(),
		name:    "table",
		Layout:  deps.Layout,
		theme:   deps.Theme,
		log:     deps.Log,
		Help:    help.New(),
		Keys:    defaultKeys(),
		sortCol: -1,
		sortAsc: true,
	}

	for _, opt := range opts {
		opt(m)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	cols := m.pendingCols
	if cols == nil {
		cols = []table.Column{}
	}
	rows := m.pendingRows
	if rows == nil {
		rows = []table.Row{}
	}

	m.originalCols = slices.Clone(cols)
	m.originalRows = cloneRows(rows)
	m.BubbleTable = table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(max(1, len(rows))),
		table.WithWidth(m.tableWidthLocked()),
	)
	m.BubbleTable.SetStyles(m.tableStyles())
	m.updateTableSizeLocked()

	return m
}

func (m *Model) Theme() *theme.Theme { return m.theme }

// SetTheme swaps the palette and rebuilds the cached table styles.
func (m *Model) SetTheme(th *theme.Theme) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.theme = th
	m.BubbleTable.SetStyles(m.tableStyles())
}

func (m *Model) Init() tea.Cmd      { return nil }
func (m *Model) ID() ui.ComponentID { return m.id }
func (m *Model) Name() string       { return m.name }
func (m *Model) OnClose() tea.Cmd   { return nil }
func (m *Model) Mode() env.Mode {
	if !m.filtering && m.filter == "" {
		return env.NormalMode
	}
	return env.ManagedMode
}

// SetPaddingY sets the vertical padding.
func (m *Model) SetPaddingY(padding int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.paddingY = padding
	m.updateTableSizeLocked()
}

// SetMaxWidth sets an optional maximum width (0 means unlimited).
func (m *Model) SetMaxWidth(w int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.maxWidth = w
	m.updateTableSizeLocked()
}

// SetMaxHeight sets an optional maximum height (0 means unlimited).
func (m *Model) SetMaxHeight(h int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.maxHeight = h
	m.updateTableSizeLocked()
}

func (m *Model) AppendRow(row table.Row) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.originalRows = append(m.originalRows, slices.Clone(row))
	m.refreshDisplayLocked(false)
}

// SetRows replaces all rows, keeping the active sort + filter and the
// cursor position (live scan refreshes must not jump).
func (m *Model) SetRows(rows []table.Row) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.originalRows = cloneRows(rows)
	m.refreshDisplayLocked(false)
}

// SortState reports the active sort column (-1 means load order) and
// direction.
func (m *Model) SortState() (col int, asc bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.sortCol, m.sortAsc
}

// SetSort sorts by col. Out-of-range or unsortable columns clear the sort.
func (m *Model) SetSort(col int, asc bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if col < 0 || col >= len(m.originalCols) || m.unsortable[col] {
		m.clearSortLocked()
		return
	}
	m.sortCol, m.sortAsc = col, asc
	m.refreshDisplayLocked(true)
}

// ClearSort restores load order, keeping the cursor position.
func (m *Model) ClearSort() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.clearSortLocked()
}

func (m *Model) clearSortLocked() {
	m.sortCol = -1
	m.refreshDisplayLocked(false)
}

// cycleSortColumnLocked advances to the next sortable column, wrapping to
// load order after the last one. Caller must hold m.mu.
func (m *Model) cycleSortColumnLocked() {
	next := m.sortCol + 1
	for next < len(m.originalCols) && m.unsortable[next] {
		next++
	}
	if next >= len(m.originalCols) {
		m.clearSortLocked()
		return
	}
	m.sortCol, m.sortAsc = next, true
	m.refreshDisplayLocked(true)
}

// toggleSortDirectionLocked flips the direction, starting at the first
// sortable column (descending) when no sort is active. Caller holds m.mu.
func (m *Model) toggleSortDirectionLocked() {
	if m.sortCol < 0 {
		m.sortCol = 0
		for m.sortCol < len(m.originalCols) && m.unsortable[m.sortCol] {
			m.sortCol++
		}
		if m.sortCol >= len(m.originalCols) {
			return
		}
		m.sortAsc = false
	} else {
		m.sortAsc = !m.sortAsc
	}
	m.refreshDisplayLocked(true)
}

// Filter returns the active filter text ("" means unfiltered).
func (m *Model) Filter() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.filter
}

// Filtering reports whether keystrokes are editing the filter.
func (m *Model) Filtering() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.filtering
}

// SetFilter replaces the filter text, keeping the cursor.
func (m *Model) SetFilter(s string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.filter = s
	m.refreshDisplayLocked(false)
}

// ClearFilter removes the filter, keeping the cursor.
func (m *Model) ClearFilter() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.filter, m.filtering = "", false
	m.refreshDisplayLocked(false)
}

// refreshDisplayLocked pushes the sorted + filtered snapshot to the table.
// toTop jumps to the first row for explicit user sorts; data refreshes pass
// false to preserve the cursor. Caller holds m.mu.
func (m *Model) refreshDisplayLocked(toTop bool) {
	m.BubbleTable.SetRows(m.displayRowsLocked())
	if toTop {
		m.BubbleTable.GotoTop()
	}
	m.updateTableSizeLocked()
}

// displayRowsLocked returns the load-order snapshot with the active sort
// (stable) then filter applied. Caller holds m.mu.
func (m *Model) displayRowsLocked() []Row {
	rows := cloneRows(m.originalRows)
	if m.sortCol >= 0 {
		col, asc := m.sortCol, m.sortAsc
		slices.SortStableFunc(rows, func(a, b Row) int {
			c := compareCells(cellAt(a, col), cellAt(b, col))
			if !asc {
				c = -c
			}
			return c
		})
	}
	if m.filter != "" {
		needle := strings.ToLower(m.filter)
		rows = slices.DeleteFunc(rows, func(r Row) bool {
			return !rowMatches(r, needle)
		})
	}
	return rows
}

// rowMatches reports whether any cell contains the (already lowered)
// needle, case-insensitively.
func rowMatches(r Row, needle string) bool {
	for _, c := range r {
		if strings.Contains(strings.ToLower(c), needle) {
			return true
		}
	}
	return false
}

func cellAt(r Row, col int) string {
	if col < 0 || col >= len(r) {
		return ""
	}
	return r[col]
}

// compareCells orders rendered cells: numerically, then by byte size
// ("10 MB" > "9 MB"), then lexicographically.
func compareCells(a, b string) int {
	a, b = strings.TrimSpace(a), strings.TrimSpace(b)
	if fa, err := strconv.ParseFloat(a, 64); err == nil {
		fb, err := strconv.ParseFloat(b, 64)
		if err != nil {
			return -1
		}
		return cmp.Compare(fa, fb)
	} else if _, err := strconv.ParseFloat(b, 64); err == nil {
		return 1
	}
	if pa, err := humanize.ParseBytes(a); err == nil {
		pb, err := humanize.ParseBytes(b)
		if err != nil {
			return -1
		}
		return cmp.Compare(pa, pb)
	} else if _, err := humanize.ParseBytes(b); err == nil {
		return 1
	}
	return strings.Compare(a, b)
}

// NewRowTime renders t as "2006-01-02 15:04:05", or "-" when t is zero.
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

// scaledColumns distributes the content width across columns proportionally
// to their original widths, accounting for bubbles' per-cell padding (2
// chars/col) and inter-column borders (nCols-1 chars).
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
	// Rendered width = sum(colWidths) + 2*nCols (padding) + (nCols-1).
	overhead := 2*nCols + (nCols - 1)
	target := max(nCols, m.tableWidthLocked()-overhead)

	// Titles are much shorter than their column widths, so the 2-cell
	// marker never clips.
	if m.sortCol >= 0 && m.sortCol < len(out) {
		if m.sortAsc {
			out[m.sortCol].Title += " ▲"
		} else {
			out[m.sortCol].Title += " ▼"
		}
	}

	assigned := 0
	for i, c := range src {
		w := int(float64(target) * float64(c.Width) / float64(total))
		out[i].Width = max(1, w)
		assigned += out[i].Width
	}
	// Rounding remainder goes to the last column.
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

	m.BubbleTable.SetColumns(m.scaledColumns())

	helpHeight := lipgloss.Height(m.renderHelpView())
	titleHeight := lipgloss.Height(m.renderTitle())
	filterHeight := 1
	height := max(1, m.tableHeightLocked()-helpHeight-titleHeight-filterHeight-m.paddingY)
	m.BubbleTable.SetHeight(height)
	m.BubbleTable.SetWidth(m.tableWidthLocked())
}

func (m *Model) tableWidthLocked() int {
	if m.Layout == nil || m.Layout.Body.Width == 0 {
		return 48
	}
	w := max(10, m.Layout.Body.Width-tableHPad)
	if m.maxWidth > 0 {
		w = min(w, m.maxWidth)
	}
	return w
}

func (m *Model) tableHeightLocked() int {
	if m.Layout == nil || m.Layout.Body.Height == 0 {
		return 15
	}
	h := max(15, m.Layout.Body.Height)
	if m.maxHeight > 0 {
		h = min(h, m.maxHeight)
	}
	return h
}

// ActionKey binds a set of keys to a command and its help text.
type ActionKey struct {
	Keys      []string
	ShortHelp string
	FullHelp  string
	Cmd       tea.Cmd
}

var arrowSymbols = map[string]string{
	"up": "↑", "down": "↓", "left": "←", "right": "→",
}

// NewKey creates an ActionKey, replacing arrow key names with their symbols and
// prefixing the help strings with the joined key list.
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

// KeyMap is an ordered set of ActionKey bindings.
type KeyMap struct{ Actions []ActionKey }

// Add appends a binding.
func (k *KeyMap) Add(a ActionKey) { k.Actions = append(k.Actions, a) }

// Check returns the command bound to the given key message, or nil if none
// matches.
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

// ShortHelp implements help.KeyMap.
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

// FullHelp implements help.KeyMap.
func (k KeyMap) FullHelp(width int) [][]key.Binding {
	if len(k.Actions) == 0 {
		return nil
	}
	colCount := 3
	if width > 100 {
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

	km.Add(NewKey([]string{"s"}, "sort", "Change sort field", nil))
	km.Add(NewKey([]string{"o"}, "order", "Toggle sort order", nil))
	km.Add(NewKey([]string{"/"}, "filter", "Filter rows", nil))

	for _, k := range extra {
		km.Add(k)
	}

	km.Add(NewKey([]string{"?"}, "help", "Toggle help", nil))
	km.Add(NewKey([]string{"q", "esc"}, "quit", "Quit", nil))
	return km
}

// SetKeys replaces the current bindings with the provided actions merged with
// the default navigation keys.
func (m *Model) SetKeys(keys ...ActionKey) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Keys = defaultKeys(keys...)
}
