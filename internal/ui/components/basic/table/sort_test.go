package table

import (
	"testing"

	"github.com/MohsenBg/bgscan/internal/logger"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
	"github.com/MohsenBg/bgscan/internal/ui/theme"

	"charm.land/bubbles/v2/table"
)

func testTheme(t *testing.T) *theme.Theme {
	t.Helper()
	theme.Init()
	th, err := theme.Get("bgscan-dark")
	if err != nil {
		t.Fatalf("theme.Get: %v", err)
	}
	return th
}

func TestCompareCells(t *testing.T) {
	cases := []struct {
		a, b string
		want int // sign only
	}{
		{"9", "10", -1},       // numeric, not lexicographic
		{"10", "9", 1},        // numeric, not lexicographic
		{"9 MB", "10 MB", -1}, // humanized sizes by bytes
		{"1.2 GB", "800 MB", 1},
		// ISO timestamps sort chronologically as plain strings.
		{"2024-01-02 15:04:05", "2024-01-03 15:04:05", -1},
		{"b", "a", 1},
		{"", "a", -1},
		{"same", "same", 0},
		// Numbers sort before non-numbers.
		{"10", "abc", -1},
		{"abc", "10", 1},
	}

	for _, c := range cases {
		if got := compareCells(c.a, c.b); sign(got) != c.want {
			t.Errorf("compareCells(%q, %q) = %d, want sign %d", c.a, c.b, got, c.want)
		}
	}
}

func sign(n int) int {
	switch {
	case n < 0:
		return -1
	case n > 0:
		return 1
	default:
		return 0
	}
}

func newSortTestModel(t *testing.T) *Model {
	return New(ui.Deps{Theme: testTheme(t), Log: logger.DiscardSet()},
		WithColumns([]table.Column{
			{Title: "Name", Width: 20},
			{Title: "Size", Width: 10},
		}),
		WithRows([]table.Row{
			{"b", "10 MB"},
			{"a", "9 MB"},
			{"c", "1.2 GB"},
		}),
	)
}

func rowNames(m *Model) []string {
	rows := m.BubbleTable.Rows()
	out := make([]string, len(rows))
	for i, r := range rows {
		out[i] = r[0]
	}
	return out
}

func equalStr(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestSetSortOrdersRows(t *testing.T) {
	m := newSortTestModel(t)

	m.SetSort(0, true)
	if got, want := rowNames(m), []string{"a", "b", "c"}; !equalStr(got, want) {
		t.Fatalf("sort name asc = %v, want %v", got, want)
	}

	m.SetSort(1, false)
	if got, want := rowNames(m), []string{"c", "b", "a"}; !equalStr(got, want) {
		t.Fatalf("sort size desc = %v, want %v", got, want)
	}

	if col, asc := m.SortState(); col != 1 || asc {
		t.Fatalf("SortState = (%d, %v), want (1, false)", col, asc)
	}
}

func TestSetRowsKeepsActiveSort(t *testing.T) {
	m := newSortTestModel(t)
	m.SetSort(0, true)

	m.SetRows([]table.Row{{"z", "1 MB"}, {"m", "2 MB"}})
	if got, want := rowNames(m), []string{"m", "z"}; !equalStr(got, want) {
		t.Fatalf("SetRows with active sort = %v, want %v", got, want)
	}
}

func TestClearSortRestoresLoadOrder(t *testing.T) {
	m := newSortTestModel(t)
	m.SetSort(0, true)
	m.ClearSort()

	if got, want := rowNames(m), []string{"b", "a", "c"}; !equalStr(got, want) {
		t.Fatalf("after ClearSort = %v, want load order %v", got, want)
	}
	if col, _ := m.SortState(); col != -1 {
		t.Fatalf("SortState col = %d, want -1", col)
	}
}

func TestCycleSortColumnSkipsUnsortable(t *testing.T) {
	m := New(ui.Deps{Theme: testTheme(t), Log: logger.DiscardSet()},
		WithColumns([]table.Column{
			{Title: "A", Width: 10},
			{Title: "B", Width: 10},
		}),
		WithRows([]table.Row{{"2", "x"}, {"1", "y"}}),
		WithUnsortableColumns(0),
	)

	m.mu.Lock()
	m.cycleSortColumnLocked()
	m.mu.Unlock()
	if col, _ := m.SortState(); col != 1 {
		t.Fatalf("cycle landed on col %d, want 1 (0 is unsortable)", col)
	}

	// Cycling past the last column returns to load order.
	m.mu.Lock()
	m.cycleSortColumnLocked()
	m.mu.Unlock()
	if col, _ := m.SortState(); col != -1 {
		t.Fatalf("cycle past last col = %d, want -1 (load order)", col)
	}
	if got, want := rowNames(m), []string{"2", "1"}; !equalStr(got, want) {
		t.Fatalf("after cycle to none = %v, want load order %v", got, want)
	}
}

func TestSetSortInvalidColumnClears(t *testing.T) {
	m := newSortTestModel(t)
	m.SetSort(0, true)
	m.SetSort(99, true)
	if col, _ := m.SortState(); col != -1 {
		t.Fatalf("invalid SetSort col = %d, want -1", col)
	}
}
