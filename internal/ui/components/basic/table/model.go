package table

import (
	"cmp"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/table"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/MohsenBg/bgscan/internal/ui/components/basic/help"
	"github.com/MohsenBg/bgscan/internal/ui/shared/keys"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	"github.com/dustin/go-humanize"
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

	title string
	env   *ui.Env
	help  ui.Component

	bubble table.Model

	additionalShortHints []keys.Hint
	additionalLongHints  []keys.Hint

	originalCols []table.Column
	originalRows []Row

	width     int // 0 means use layout
	height    int // 0 means use layout
	maxWidth  int // 0 means unlimited
	maxHeight int // 0 means unlimited

	sortCol    int // -1 means load order
	sortAsc    bool
	unsortable map[int]bool

	selPrefix string
	marker    string

	emptyMessage string
	filtering    bool
	filter       textinput.Model

	spinner        spinner.Model
	loading        bool
	loadingMessage string
}

type Option func(*Model)

func WithTitle(title string) Option {
	return func(m *Model) { m.title = title }
}

func WithColumns(cols []table.Column) Option {
	return func(m *Model) { m.originalCols = slices.Clone(cols) }
}

func WithRows(rows []table.Row) Option {
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
	return func(m *Model) { m.additionalShortHints = hints }
}

func WithKeyLongHints(hints ...keys.Hint) Option {
	return func(m *Model) { m.additionalLongHints = hints }
}

// WithEmptyMessage sets the message displayed when the table has no rows.
func WithEmptyMessage(msg string) Option {
	return func(m *Model) { m.emptyMessage = msg }
}

// WithLoading sets the initial loading state.
// The default loading state is false.
func WithLoading(loading bool) Option {
	return func(m *Model) { m.loading = loading }
}

// WithLoadingMessage sets the message displayed while loading.
func WithLoadingMessage(message string) Option {
	return func(m *Model) { m.loadingMessage = message }
}

// WithUnsortableColumns excludes columns from sort cycling and SetSort.
func WithUnsortableColumns(cols ...int) Option {
	return func(m *Model) {
		if m.unsortable == nil {
			m.unsortable = make(map[int]bool)
		}

		for _, c := range cols {
			m.unsortable[c] = true
		}
	}
}

func New(env *ui.Env, opts ...Option) *Model {
	m := &Model{
		id:             ui.NewComponentID(),
		name:           "table",
		env:            env,
		sortCol:        -1,
		sortAsc:        true,
		emptyMessage:   "No data to display.",
		loadingMessage: "Loading records please wait!",
		loading:        false,
		spinner:        spinner.New(spinner.WithSpinner(spinner.Points)),
	}

	m.filter = textinput.New()
	m.filter.Prompt = ""
	m.filter.CharLimit = maxFilterChar

	for _, opt := range opts {
		opt(m)
	}

	if m.originalCols == nil {
		m.originalCols = []table.Column{}
	}

	if m.originalRows == nil {
		m.originalRows = []table.Row{}
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

func (m *Model) ID() ui.ComponentID {
	return m.id
}

func (m *Model) Name() string {
	return m.name
}

func (m *Model) OnClose() tea.Cmd {
	return nil
}

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

// SelectedIndex returns the selected visible row index.
// It returns -1 when there are no visible rows or the table is loading.
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

// EnableLoading enables the loading state.
func (m *Model) EnableLoading() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.setLoadingLocked(true)
}

// DisableLoading disables the loading state.
func (m *Model) DisableLoading() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.setLoadingLocked(false)
}

// SetLoading sets the loading state explicitly.
func (m *Model) SetLoading(loading bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.setLoadingLocked(loading)
}

// setLoadingLocked changes the loading state.
// It must be called with m.mu held.
func (m *Model) setLoadingLocked(loading bool) {
	if m.loading == loading {
		return
	}

	m.loading = loading
	m.updateTableSizeLocked()
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

func (m *Model) AppendRow(row table.Row) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.originalRows = append(m.originalRows, slices.Clone(row))
	m.refreshDisplayLocked(false)
}

// SetRows replaces all rows, keeping the active sort, filter, and cursor
// position. This is important for live scan refreshes.
func (m *Model) SetRows(rows []table.Row) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.originalRows = cloneRows(rows)
	m.refreshDisplayLocked(false)
}

// SortState reports the active sort column and direction.
// A column of -1 means load order.
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

// cycleSortColumnLocked advances to the next sortable column, wrapping to
// load order after the last one. It must be called with m.mu held.
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

// toggleSortDirectionLocked flips the sort direction. When no sort is active,
// it starts at the first sortable column in descending order.
// It must be called with m.mu held.
func (m *Model) toggleSortDirectionLocked() {
	if m.sortCol < 0 {
		col := 0
		for col < len(m.originalCols) && m.unsortable[col] {
			col++
		}

		if col >= len(m.originalCols) {
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
// resetCursor jumps to the first row for explicit user sorts; data refreshes
// pass false to preserve the cursor. It must be called with m.mu held.
func (m *Model) refreshDisplayLocked(resetCursor bool) {
	m.bubble.SetRows(m.displayRowsLocked())

	if resetCursor {
		m.bubble.GotoTop()
	}

	m.updateTableSizeLocked()
}

// displayRowsLocked returns the load-order snapshot with active sort and
// filter applied. Sort order is stable. It must be called with m.mu held.
func (m *Model) displayRowsLocked() []Row {
	needle := strings.ToLower(m.filterValue())

	if m.sortCol < 0 && needle == "" {
		return m.originalRows
	}

	rows := slices.Clone(m.originalRows)

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

	if needle != "" {
		rows = slices.DeleteFunc(rows, func(r Row) bool {
			return !rowMatches(r, needle)
		})
	}

	return rows
}

// rowMatches reports whether any cell contains needle case-insensitively.
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

// compareCells orders rendered cells numerically, then by byte size
// (for example, "10 MB" > "9 MB"), then lexicographically.
func compareCells(a, b string) int {
	a = strings.TrimSpace(a)
	b = strings.TrimSpace(b)

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
		return -1 // Byte sizes sort before other strings.
	case errB == nil:
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

// scaledColumns distributes content width across columns proportionally.
// It must be called with m.mu held.
func (m *Model) scaledColumns() []table.Column {
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

	overhead := (cellPadLeft+1)*len(out) + m.frameH()
	target := max(len(out), m.tableWidthLocked()-overhead)

	if m.sortCol >= 0 && m.sortCol < len(out) {
		if m.sortAsc {
			out[m.sortCol].Title += " ▲"
		} else {
			out[m.sortCol].Title += " ▼"
		}
	}

	assigned := 0
	for i, c := range out {
		w := int(float64(target) * float64(c.Width) / float64(total))
		out[i].Width = max(1, w)
		assigned += out[i].Width
	}

	if diff := target - assigned; diff != 0 {
		out[len(out)-1].Width = max(1, out[len(out)-1].Width+diff)
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

	var total int

	switch {
	case m.height > 0:
		total = m.height

	case m.env.Layout.Body().Height > 0:
		total = max(
			tableMinHeight,
			rows+2+helpHeight+titleHeight+paddingVertical*2,
		)
		total = min(total, m.env.Layout.Body().Height)

	default:
		total = max(
			tableMinHeight,
			rows+helpHeight+titleHeight+paddingVertical*2,
		)
	}

	if m.maxHeight > 0 {
		total = min(total, m.maxHeight)
	}

	height := max(1, total-helpHeight-titleHeight-paddingVertical*2-m.frameV())
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
