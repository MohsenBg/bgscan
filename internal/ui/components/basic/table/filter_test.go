package table

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func pressKey(m *Model, code rune, text string) *Model {
	updated, _ := m.Update(tea.KeyPressMsg{Code: code, Text: text})
	out, ok := updated.(*Model)
	if !ok {
		panic("table.Update returned non-*Model")
	}
	return out
}

func TestSetFilterMatchesAnyCell(t *testing.T) {
	m := newSortTestModel(t)

	m.SetFilter("9 mb")
	if got, want := rowNames(m), []string{"a"}; !equalStr(got, want) {
		t.Fatalf("filter '9 mb' = %v, want %v", got, want)
	}
	if m.Filter() != "9 mb" {
		t.Fatalf("Filter() = %q, want %q", m.Filter(), "9 mb")
	}
}

func TestSetFilterCaseInsensitive(t *testing.T) {
	m := newSortTestModel(t)

	// "C" matches only row c: every size unit (MB/GB) contains "b",
	// so "B" would match all rows.
	m.SetFilter("C")
	if got, want := rowNames(m), []string{"c"}; !equalStr(got, want) {
		t.Fatalf("filter 'C' = %v, want %v", got, want)
	}
}

func TestSetRowsKeepsFilter(t *testing.T) {
	m := newSortTestModel(t)
	m.SetFilter("mb")

	m.SetRows([]Row{{"z", "1 MB"}, {"q", "plain"}})
	if got, want := rowNames(m), []string{"z"}; !equalStr(got, want) {
		t.Fatalf("SetRows with filter = %v, want %v", got, want)
	}
}

func TestClearFilterRestoresRows(t *testing.T) {
	m := newSortTestModel(t)
	m.SetFilter("zzz")
	if n := len(m.BubbleTable.Rows()); n != 0 {
		t.Fatalf("filter 'zzz' shows %d rows, want 0", n)
	}

	m.ClearFilter()
	if got, want := rowNames(m), []string{"b", "a", "c"}; !equalStr(got, want) {
		t.Fatalf("after ClearFilter = %v, want load order %v", got, want)
	}
}

func TestSortAndFilterCompose(t *testing.T) {
	m := newSortTestModel(t)
	m.SetSort(0, true)
	// "mb" matches the two MB rows but not "1.2 GB".
	m.SetFilter("mb")

	if got, want := rowNames(m), []string{"a", "b"}; !equalStr(got, want) {
		t.Fatalf("sorted + filtered = %v, want %v", got, want)
	}
}

func TestFilterKeysEndToEnd(t *testing.T) {
	m := newSortTestModel(t)

	// "/" enters filter mode.
	m = pressKey(m, '/', "/")
	if !m.Filtering() {
		t.Fatal("expected filtering mode after '/'")
	}

	// Typing filters live; table nav must not move the cursor.
	m = pressKey(m, 'a', "a")
	if got, want := rowNames(m), []string{"a"}; !equalStr(got, want) {
		t.Fatalf("after typing 'a' = %v, want %v", got, want)
	}

	// Backspace removes the char, restoring all rows.
	m = pressKey(m, tea.KeyBackspace, "")
	if got, want := rowNames(m), []string{"b", "a", "c"}; !equalStr(got, want) {
		t.Fatalf("after backspace = %v, want %v", got, want)
	}

	// Esc clears the filter and exits edit mode.
	m = pressKey(m, 'a', "a")
	m = pressKey(m, tea.KeyEsc, "")
	if m.Filtering() || m.Filter() != "" {
		t.Fatalf("after esc: filtering=%v filter=%q, want false/\"\"", m.Filtering(), m.Filter())
	}
	if got, want := rowNames(m), []string{"b", "a", "c"}; !equalStr(got, want) {
		t.Fatalf("after esc rows = %v, want %v", got, want)
	}

	// Enter keeps the filter and exits edit mode. ("9" matches only
	// row a; "b" would also match the MB/GB units.)
	m = pressKey(m, '/', "/")
	m = pressKey(m, '9', "9")
	m = pressKey(m, tea.KeyEnter, "")
	if m.Filtering() || m.Filter() != "9" {
		t.Fatalf("after enter: filtering=%v filter=%q, want false/\"9\"", m.Filtering(), m.Filter())
	}
	if got, want := rowNames(m), []string{"a"}; !equalStr(got, want) {
		t.Fatalf("after enter rows = %v, want %v", got, want)
	}
}
