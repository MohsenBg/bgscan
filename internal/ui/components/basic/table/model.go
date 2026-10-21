package table

import (
	"cmp"
	"fmt"
	"net/netip"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/table"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/dustin/go-humanize"

	"github.com/MohsenBg/bgscan/internal/core/result"
	"github.com/MohsenBg/bgscan/internal/core/speedtest"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/help"
	"github.com/MohsenBg/bgscan/internal/ui/shared/keys"
	"github.com/MohsenBg/bgscan/internal/ui/shared/meter"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
)

type (
	Column = table.Column
	Row    = table.Row
)

const (
	paddingVertical   = 2
	paddingHorizontal = 4
	helpHeight        = 3
	titleHeight       = 3
	tableMinHeight    = 25
	tableMinWidth     = 30

	markerGlyph   = "▐ "
	cellPadLeft   = 2
	markerOpacity = 0.8

	maxFilterChar = 35
)

// Model is a Bubble Tea table with responsive layout, key bindings, sorting,
// filtering, loading, empty state, and safe concurrent access.
type Model struct {
	mu   sync.RWMutex
	id   ui.ComponentID
	name string

	title  string
	env    *ui.Env
	help   ui.Component
	bubble table.Model

	additionalShortHints []keys.Hint
	additionalLongHints  []keys.Hint

	originalCols []Column
	originalRows []Row
	colTypes     []result.ColumnType

	width     int // 0 means use layout
	height    int // 0 means use layout
	maxWidth  int // 0 means unlimited
	maxHeight int // 0 means unlimited

	sortCol    int // -1 means load order
	sortAsc    bool
	unsortable map[int]bool

	selPrefix string
	marker    string

	emptyMessage  string
	filterEnabled bool
	filtering     bool
	filter        textinput.Model

	spinner        spinner.Model
	loading        bool
	loadingMessage string
}

type Option func(*Model)

func WithTitle(title string) Option {
	return func(m *Model) { m.title = title }
}

func WithColumns(cols []Column) Option {
	return func(m *Model) { m.originalCols = slices.Clone(cols) }
}

// WithColumnTypes assigns a semantic type per column, by index.
func WithColumnTypes(types ...result.ColumnType) Option {
	return func(m *Model) { m.colTypes = slices.Clone(types) }
}

func WithRows(rows []Row) Option {
	return func(m *Model) { m.originalRows = cloneRows(rows) }
}

// WithWidth sets an explicit width. Zero means use layout.
func WithWidth(w int) Option {
	return func(m *Model) { m.width = w }
}

// WithHeight sets an explicit height. Zero means use layout.
func WithHeight(h int) Option {
	return func(m *Model) { m.height = h }
}

// WithMaxWidth caps the width. Zero means unlimited.
func WithMaxWidth(w int) Option {
	return func(m *Model) { m.maxWidth = w }
}

// WithMaxHeight caps the height. Zero means unlimited.
func WithMaxHeight(h int) Option {
	return func(m *Model) { m.maxHeight = h }
}

func WithKeyShortHints(hints ...keys.Hint) Option {
	return func(m *Model) { m.additionalShortHints = slices.Clone(hints) }
}

func WithKeyLongHints(hints ...keys.Hint) Option {
	return func(m *Model) { m.additionalLongHints = slices.Clone(hints) }
}

// WithEmptyMessage sets the message displayed when the table has no rows.
func WithEmptyMessage(msg string) Option {
	return func(m *Model) { m.emptyMessage = msg }
}

// WithLoading sets the initial loading state (defaults to false).
func WithLoading(loading bool) Option {
	return func(m *Model) { m.loading = loading }
}

// WithLoadingMessage sets the message displayed while loading.
func WithLoadingMessage(message string) Option {
	return func(m *Model) { m.loadingMessage = message }
}

// WithFiltering enables or disables table filtering.
func WithFiltering(enabled bool) Option {
	return func(m *Model) { m.filterEnabled = enabled }
}

// WithUnsortableColumns excludes columns from sort cycling and SetSort.
func WithUnsortableColumns(cols ...int) Option {
	return func(m *Model) {
		if m.unsortable == nil {
			m.unsortable = make(map[int]bool, len(cols))
		}
		for _, c := range cols {
			m.unsortable[c] = true
		}
	}
}

func New(env *ui.Env, opts ...Option) *Model {
	filter := textinput.New()
	filter.Prompt = ""
	filter.CharLimit = maxFilterChar

	m := &Model{
		id:             ui.NewComponentID(),
		name:           "table",
		env:            env,
		sortCol:        -1,
		sortAsc:        true,
		filterEnabled:  true,
		filter:         filter,
		emptyMessage:   "No data to display.",
		loadingMessage: "Loading records please wait!",
		spinner:        spinner.New(spinner.WithSpinner(spinner.Points)),
	}

	for _, opt := range opts {
		opt(m)
	}

	m.help = help.New(
		env,
		m.title,
		help.HelpSet{
			Short: m.shortHits(m.additionalShortHints...),
			Long:  m.longHits(m.additionalLongHints...),
		},
	)

	m.bubble = table.New(
		table.WithColumns(m.originalCols),
		table.WithRows(m.originalRows),
		table.WithFocused(true),
		table.WithHeight(max(1, len(m.originalRows))),
		table.WithWidth(m.tableWidthLocked()),
	)

	m.updateTableSizeLocked()
	m.applyStylesLocked()

	return m
}

func (m *Model) Init() tea.Cmd {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.spinner.Tick
}

func (m *Model) ID() ui.ComponentID { return m.id }
func (m *Model) Name() string       { return m.name }
func (m *Model) OnClose() tea.Cmd   { return nil }

func (m *Model) Mode() ui.KeyMode {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.loading || m.filtering || m.filterValue() != "" {
		return ui.KeyModeManage
	}
	return ui.KeyModeNormal
}

// Title returns the table title.
func (m *Model) Title() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.title
}

// SetTitle sets the table title.
func (m *Model) SetTitle(title string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.title = title
}

// SelectedRow returns a copy of the selected visible row.
func (m *Model) SelectedRow() (Row, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.loading {
		return nil, false
	}

	rows := m.bubble.Rows()
	idx := m.bubble.Cursor()
	if idx < 0 || idx >= len(rows) {
		return nil, false
	}

	return slices.Clone(rows[idx]), true
}

// SelectedIndex returns the selected visible row index, or -1 when empty/loading.
func (m *Model) SelectedIndex() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.loading || len(m.bubble.Rows()) == 0 {
		return -1
	}
	return m.bubble.Cursor()
}

// VisibleRowCount returns the number of rows currently visible after
// filtering and sorting. It returns 0 while loading.
func (m *Model) VisibleRowCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.loading {
		return 0
	}
	return len(m.bubble.Rows())
}

// TotalRowCount returns the total number of original rows.
func (m *Model) TotalRowCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.originalRows)
}

// GotoTop moves the cursor to the first visible row.
func (m *Model) GotoTop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.bubble.GotoTop()
}

// GotoBottom moves the cursor to the last visible row.
func (m *Model) GotoBottom() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.bubble.GotoBottom()
}

// Loading reports whether the table is currently in a loading state.
func (m *Model) Loading() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.loading
}

func (m *Model) EnableLoading()  { m.SetLoading(true) }
func (m *Model) DisableLoading() { m.SetLoading(false) }

// SetLoading sets the loading state explicitly.
func (m *Model) SetLoading(loading bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.loading != loading {
		m.loading = loading
		m.updateTableSizeLocked()
	}
}

// SetWidth sets an explicit width. Zero means use layout.
func (m *Model) SetWidth(w int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.width = w
	m.updateTableSizeLocked()
}

// SetHeight sets an explicit height. Zero means use layout.
func (m *Model) SetHeight(h int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.height = h
	m.updateTableSizeLocked()
}

// SetMaxWidth sets an optional maximum width. Zero means unlimited.
func (m *Model) SetMaxWidth(w int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.maxWidth = w
	m.updateTableSizeLocked()
}

// SetMaxHeight sets an optional maximum height. Zero means unlimited.
func (m *Model) SetMaxHeight(h int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.maxHeight = h
	m.updateTableSizeLocked()
}

func (m *Model) AppendRow(row Row) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.originalRows = append(m.originalRows, slices.Clone(row))
	m.refreshDisplayLocked(false)
}

// SetRows replaces all rows, keeping the active sort, filter, and cursor
// position. This is important for live scan refreshes.
func (m *Model) SetRows(rows []Row) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.originalRows = cloneRows(rows)
	m.refreshDisplayLocked(false)
}

// SortState reports the active sort column and direction (-1 means load order).
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

// ClearSort restores load order while preserving cursor position.
func (m *Model) ClearSort() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.clearSortLocked()
}

func (m *Model) clearSortLocked() {
	m.sortCol = -1
	m.sortAsc = true
	m.refreshDisplayLocked(false)
}

// nextSortableCol returns the first sortable column index >= start, or -1 if none.
func (m *Model) nextSortableCol(start int) int {
	for col := max(0, start); col < len(m.originalCols); col++ {
		if !m.unsortable[col] {
			return col
		}
	}
	return -1
}

// cycleSortColumnLocked advances to the next sortable column, wrapping to
// load order after the last one. It must be called with m.mu held.
func (m *Model) cycleSortColumnLocked() {
	next := m.nextSortableCol(m.sortCol + 1)
	if next < 0 {
		m.clearSortLocked()
		return
	}

	m.sortCol, m.sortAsc = next, true
	m.refreshDisplayLocked(true)
}

// toggleSortDirectionLocked flips the sort direction. When no sort is active,
// it starts at the first sortable column in descending order.
func (m *Model) toggleSortDirectionLocked() {
	if m.sortCol < 0 {
		col := m.nextSortableCol(0)
		if col < 0 {
			return
		}
		m.sortCol, m.sortAsc = col, false
	} else {
		m.sortAsc = !m.sortAsc
	}

	m.refreshDisplayLocked(true)
}

func (m *Model) Filtering() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.filtering
}

// ClearFilter removes the filter while preserving cursor position.
func (m *Model) ClearFilter() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.clearFilterLocked()
}

func (m *Model) clearFilterLocked() {
	m.filter.SetValue("")
	m.filtering = false
	m.hideCursor()
	m.refreshDisplayLocked(false)
}

// refreshDisplayLocked pushes the sorted and filtered snapshot to the table.
func (m *Model) refreshDisplayLocked(resetCursor bool) {
	m.bubble.SetRows(m.displayRowsLocked())
	if resetCursor {
		m.bubble.GotoTop()
	}
	m.updateTableSizeLocked()
}

// displayRowsLocked returns the load-order snapshot with active filter and
// sort applied. Sort order is stable. It must be called with m.mu held.
func (m *Model) displayRowsLocked() []Row {
	needle := strings.ToLower(m.filterValue())
	if m.sortCol < 0 && needle == "" {
		return m.originalRows
	}

	// Filter before sorting to avoid sorting rows that will be discarded.
	var rows []Row
	if needle != "" {
		rows = make([]Row, 0, len(m.originalRows))
		for _, r := range m.originalRows {
			if rowMatches(r, needle) {
				rows = append(rows, r)
			}
		}
	} else {
		rows = slices.Clone(m.originalRows)
	}

	if m.sortCol >= 0 {
		col, asc := m.sortCol, m.sortAsc
		compare := m.comparator(col)

		slices.SortStableFunc(rows, func(a, b Row) int {
			c := compare(cellAt(a, col), cellAt(b, col))
			if !asc {
				return -c
			}
			return c
		})
	}

	return rows
}

// rowMatches reports whether any cell contains needle case-insensitively.
func rowMatches(r Row, needle string) bool {
	for _, c := range r {
		if strings.Contains(strings.ToLower(ansi.Strip(c)), needle) {
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

// compareCells orders rendered cells numerically, then by byte size
// (for example, "10 MB" > "9 MB"), then lexicographically.
func compareCells(a, b string) int {
	a = strings.TrimSpace(ansi.Strip(a))
	b = strings.TrimSpace(ansi.Strip(b))

	if a == b {
		return 0
	}

	fa, errA := strconv.ParseFloat(a, 64)
	fb, errB := strconv.ParseFloat(b, 64)
	switch {
	case errA == nil && errB == nil:
		return cmp.Compare(fa, fb)
	case errA == nil:
		return -1 // Numbers sort before non-numbers.
	case errB == nil:
		return 1
	}

	pa, errA := humanize.ParseBytes(a)
	pb, errB := humanize.ParseBytes(b)
	switch {
	case errA == nil && errB == nil:
		return cmp.Compare(pa, pb)
	case errA == nil:
		return -1
	case errB == nil:
		return 1
	}

	return cmp.Compare(a, b)
}

// columnType reports the semantic type of a column.
func (m *Model) columnType(col int) result.ColumnType {
	if col < 0 || col >= len(m.colTypes) {
		return result.ColumnText
	}
	return m.colTypes[col]
}

// comparator picks the sort function for a column by its declared type.
func (m *Model) comparator(col int) func(a, b string) int {
	switch m.columnType(col) {
	case result.ColumnIP:
		return sortIP
	case result.ColumnDuration:
		return sortDuration
	case result.ColumnBandwidth:
		return sortBandwidth
	default:
		return compareCells
	}
}

// typedCompare builds a comparator ordering cells by parsed values.
func typedCompare[T any](
	parse func(string) (T, bool),
	order func(T, T) int,
) func(a, b string) int {
	return func(a, b string) int {
		sa := strings.TrimSpace(ansi.Strip(a))
		sb := strings.TrimSpace(ansi.Strip(b))

		pa, oka := parse(sa)
		pb, okb := parse(sb)

		switch {
		case oka && okb:
			return order(pa, pb)
		case oka:
			return -1
		case okb:
			return 1
		default:
			return compareCells(sa, sb)
		}
	}
}

var (
	sortIP = typedCompare(
		func(s string) (netip.Addr, bool) {
			p, err := netip.ParseAddr(s)
			return p, err == nil
		},
		netip.Addr.Compare,
	)

	sortDuration = typedCompare(
		meter.Duration,
		cmp.Compare[time.Duration],
	)

	sortBandwidth = typedCompare(
		func(s string) (speedtest.BitsPerSec, bool) {
			v, err := speedtest.ParseBitsPerSec(s)
			return v, err == nil
		},
		cmp.Compare[speedtest.BitsPerSec],
	)
)

// NewRowTime renders t as "2006-01-02 15:04:05", or "-" when t is zero.
func NewRowTime(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.Format(time.DateTime)
}

func cloneRows(rows []Row) []Row {
	if rows == nil {
		return nil
	}
	out := make([]Row, len(rows))
	for i, row := range rows {
		out[i] = slices.Clone(row)
	}
	return out
}

// scaledColumns distributes content width across columns proportionally.
// It must be called with m.mu held.
func (m *Model) scaledColumns() []Column {
	if len(m.originalCols) == 0 {
		return nil
	}

	out := slices.Clone(m.originalCols)

	total := 0
	for _, c := range out {
		total += c.Width
	}
	if total <= 0 {
		return out
	}

	if m.sortCol >= 0 && m.sortCol < len(out) {
		if m.sortAsc {
			out[m.sortCol].Title += " ▲"
		} else {
			out[m.sortCol].Title += " ▼"
		}
	}

	overhead := (cellPadLeft+1)*len(out) + m.frameH()
	target := max(len(out), m.tableWidthLocked()-overhead)

	assigned := 0
	for i, c := range out {
		out[i].Width = max(1, (target*c.Width)/total)
		assigned += out[i].Width
	}

	if diff := target - assigned; diff != 0 {
		last := len(out) - 1
		out[last].Width = max(1, out[last].Width+diff)
	}

	return out
}

func (m *Model) updateTableSizeLocked() {
	if len(m.originalCols) == 0 {
		return
	}

	m.bubble.SetColumns(m.scaledColumns())

	rows := len(m.bubble.Rows())
	if m.loading {
		rows = 0
	}

	chromeHeight := helpHeight + titleHeight + paddingVertical*2
	var total int

	switch {
	case m.height > 0:
		total = m.height
	case m.env.Layout.Body().Height > 0:
		total = min(
			max(tableMinHeight, rows+2+chromeHeight),
			m.env.Layout.Body().Height,
		)
	default:
		total = max(tableMinHeight, rows+chromeHeight)
	}

	if m.maxHeight > 0 {
		total = min(total, m.maxHeight)
	}

	height := max(1, total-chromeHeight-m.frameV())
	m.bubble.SetHeight(height)

	width := m.tableWidthLocked()
	m.bubble.SetWidth(m.contentWidthLocked())

	if h, ok := m.help.(*help.Model); ok {
		h.SetShortWidth(width)
	}
}

func (m *Model) contentWidthLocked() int {
	return max(1, m.tableWidthLocked()-m.frameH())
}

func (m *Model) tableWidthLocked() int {
	var w int
	switch {
	case m.width > 0:
		w = m.width
	case m.env.Layout.Body().Width > 0:
		w = max(10, m.env.Layout.Body().Width-paddingHorizontal)
	default:
		w = tableMinWidth
	}

	if m.maxWidth > 0 {
		w = min(w, m.maxWidth)
	}

	return max(1, w)
}

// hideCursor hides the virtual cursor while keeping the input active.
func (m *Model) hideCursor() {
	m.filter.SetVirtualCursor(false)
	m.filter.Blur()
}

// showCursor shows the virtual cursor and returns the blink command.
func (m *Model) showCursor() tea.Cmd {
	m.filter.SetVirtualCursor(true)
	return m.filter.Focus()
}

func (m *Model) filterValue() string {
	return strings.TrimSpace(m.filter.Value())
}

func (m *Model) renderLoading() string {
	return fmt.Sprintf("%s %s", m.loadingMessage, m.spinner.View())
}
