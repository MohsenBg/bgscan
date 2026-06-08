package iplist

import (
	"os"
	"slices"

	"bgscan/internal/core/iplist"
	"bgscan/internal/logger"
	"bgscan/internal/ui/components/basic/crud"
	"bgscan/internal/ui/components/basic/notice"
	"bgscan/internal/ui/components/basic/table"
	"bgscan/internal/ui/shared/layout"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dustin/go-humanize"
)

type provider struct {
	layout   *layout.Layout
	onSelect func(*iplist.IPFileInfo) tea.Cmd
}

func newProvider(layout *layout.Layout, onSelect func(*iplist.IPFileInfo) tea.Cmd) crud.Provider[iplist.IPFileInfo] {
	return &provider{
		layout:   layout,
		onSelect: onSelect,
	}
}

func (p *provider) Title() string { return "IP Files" }

func (p *provider) Columns() []table.Column {
	return []table.Column{
		{Title: "Name", Width: 30},
		{Title: "Created Time", Width: 35},
		{Title: "Size", Width: 30},
	}
}

func (p *provider) Load() ([]iplist.IPFileInfo, error) {
	files, err := iplist.ListIPFiles()
	if err != nil {
		logger.UIError("Failed to load IP files: %v", err)
		return nil, err
	}
	logger.UIInfo("Loaded %d IP files", len(files))

	// Explicit sorting: Newest files first using built-in time comparison patterns
	slices.SortFunc(files, func(i, j iplist.IPFileInfo) int {
		return j.CreatedAt.Compare(i.CreatedAt)
	})
	return files, nil
}

func (p *provider) RenderRow(item iplist.IPFileInfo) table.Row {
	return table.Row{
		item.Name,
		item.CreatedAt.Format("2006-01-02 15:04:05"),
		humanize.Bytes(uint64(item.Size)),
	}
}

func (p *provider) Identity(item iplist.IPFileInfo) string {
	return item.Name
}

func (p *provider) OnSelect(item iplist.IPFileInfo) (tea.Cmd, bool) {
	if p.onSelect != nil {
		return p.onSelect(&item), true
	}
	return nil, false
}

// OnDelete runs asynchronously inside an isolated thread to preserve user interface frame rates
func (p *provider) OnDelete(item iplist.IPFileInfo) (tea.Cmd, bool) {
	cmd := func() tea.Msg {
		if err := os.Remove(item.Path); err != nil && !os.IsNotExist(err) {
			logger.UIError("Failed to delete IP file: %v", err)
			return notice.NewNoticeCmd(p.layout, "Delete Failed", err.Error(), notice.NOTICE_ERROR)()
		}
		return nil
	}
	return cmd, true
}

// OnRename runs asynchronously inside an isolated thread to protect the main run-loop
func (p *provider) OnRename(item iplist.IPFileInfo, newName string) (tea.Cmd, bool) {
	cmd := func() tea.Msg {
		dstPath, err := iplist.GetIPFilePath(newName)
		if err != nil {
			logger.UIError("Failed to resolve destination path: %v", err)
			return notice.NewNoticeCmd(p.layout, "Rename Failed", err.Error(), notice.NOTICE_ERROR)()
		}

		if err := os.Rename(item.Path, dstPath); err != nil {
			logger.UIError("Rename failed: %v", err)
			return notice.NewNoticeCmd(p.layout, "Rename Failed", err.Error(), notice.NOTICE_ERROR)()
		}
		return nil
	}
	return cmd, true
}

func (p *provider) OnAdd(item iplist.IPFileInfo) (tea.Cmd, bool) {
	return nil, true
}

