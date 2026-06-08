package resultlist

import (
	"bgscan/internal/core/result"
	"bgscan/internal/logger"
	"bgscan/internal/ui/components/basic/crud"
	"bgscan/internal/ui/components/basic/notice"
	"bgscan/internal/ui/components/basic/table"
	"bgscan/internal/ui/shared/layout"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dustin/go-humanize"
)

type provider struct {
	layout   *layout.Layout
	onSelect func(*result.ResultFile) tea.Cmd
}

func newProvider(l *layout.Layout, onSelect func(*result.ResultFile) tea.Cmd) crud.Provider[result.ResultFile] {
	return &provider{
		layout:   l,
		onSelect: onSelect,
	}
}

func (p *provider) OnAdd(item result.ResultFile) (tea.Cmd, bool) {
	return nil, true
}

func (p *provider) Title() string {
	return "Scan Results Log"
}

func (p *provider) Columns() []table.Column {
	return []table.Column{
		{Title: "File Name", Width: 40},
		{Title: "Created Time", Width: 25},
		{Title: "Type", Width: 15},
		{Title: "Size", Width: 15},
	}
}

func (p *provider) Load() ([]result.ResultFile, error) {
	files, err := result.GetResultFiles()
	if err != nil {
		logger.UIError("Failed to load result logs: %v", err)
		return nil, err
	}
	logger.DebugDump("rs", files)
	logger.UIInfo("Loaded %d result files from disk", len(files))
	return files, nil
}

func (p *provider) RenderRow(item result.ResultFile) table.Row {
	return table.Row{
		item.Name,
		item.CreatedTime.Format("2006-01-02 15:04:05"),
		item.Type.String(),
		humanize.Bytes(uint64(item.SizeBytes)),
	}
}

func (p *provider) Identity(item result.ResultFile) string {
	return item.Name
}

func (p *provider) OnSelect(item result.ResultFile) (tea.Cmd, bool) {
	if p.onSelect != nil {
		return p.onSelect(&item), true
	}
	return nil, true
}

func (p *provider) OnDelete(item result.ResultFile) (tea.Cmd, bool) {
	if err := os.Remove(item.Path); err != nil && !os.IsNotExist(err) {
		logger.UIError("Failed to delete result log file: %v", err)
		return notice.NewNoticeCmd(p.layout, "Delete Failed", err.Error(), notice.NOTICE_ERROR), true
	}
	return nil, true
}

func (p *provider) OnRename(item result.ResultFile, newName string) (tea.Cmd, bool) {
	newName = result.NormalizeResultFileName(newName)
	dstPath := filepath.Join(filepath.Dir(item.Path), newName)
	if err := os.Rename(item.Path, dstPath); err != nil {
		logger.UIError("Failed to rename file on disk: %v", err)
		return notice.NewNoticeCmd(p.layout, "Rename Failed", err.Error(), notice.NOTICE_ERROR), true
	}
	return nil, true
}
