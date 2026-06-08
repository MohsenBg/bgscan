package ipviewer

import (
	"bgscan/internal/core/result"
	"bgscan/internal/ui/components/basic/notice"
	"bgscan/internal/ui/components/basic/table"
	"bgscan/internal/ui/shared/ui"
	"fmt"
	"time"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok && key.String() == tea.KeyEnter.String() {
		if cmd := m.copySelectedIP(); cmd != nil {
			return m, cmd
		}
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m *Model) copySelectedIP() tea.Cmd {
	t, ok := m.table.(*table.Model)
	if !ok || t == nil {
		return nil
	}

	row := t.BubbleTable.SelectedRow()
	if len(row) == 0 {
		return nil
	}

	if err := clipboard.WriteAll(row[0]); err != nil {
		return m.errorCmd("Error Copying IP", err.Error())
	}
	return m.infoCmd("IP Copied", fmt.Sprintf("IP copied to clipboard:%s", row[0]))
}

func (m *Model) updateRows(rows []result.IPScanResult) {
	limit := min(len(rows), int(m.maxRow))
	list := make([]table.Row, 0, limit)

	for _, row := range rows[:limit] {
		if m.viewMode == ShortView {
			list = append(list, table.Row{
				row.IP,
				row.Latency.Truncate(time.Millisecond).String(),
			})
		} else {
			list = append(list, table.Row{
				row.IP,
				row.Latency.Truncate(time.Millisecond).String(),
				row.Download.Truncate(time.Millisecond).String(),
				row.Upload.Truncate(time.Millisecond).String(),
			})
		}
	}

	m.rows = list
	if t, ok := m.table.(*table.Model); ok && t != nil {
		t.BubbleTable.SetRows(m.rows)
	}
}

func (m *Model) errorCmd(title, message string) tea.Cmd {
	return notice.NewNoticeCmd(m.layout, title, message, notice.NOTICE_ERROR)
}

func (m *Model) infoCmd(title, message string) tea.Cmd {
	return notice.NewNoticeCmd(m.layout, title, message, notice.NOTICE_INFO)
}
