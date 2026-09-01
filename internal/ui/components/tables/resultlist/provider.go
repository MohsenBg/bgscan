package resultlist

import (
	"os"
	"path/filepath"
	"slices"

	"github.com/MohsenBg/bgscan/internal/core/result"
	"github.com/MohsenBg/bgscan/internal/logger"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/crud"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/notice"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/table"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
	"github.com/dustin/go-humanize"
)

type provider struct {
	title    string
	state    *ui.AppState
	onSelect func(*result.ResultFile) tea.Cmd
}

func newProvider(state *ui.AppState, title string, onSelect func(*result.ResultFile) tea.Cmd) crud.Provider[result.ResultFile] {
	return &provider{
		title:    title,
		state:    state,
		onSelect: onSelect,
	}
}

func (p *provider) OnAdd(item result.ResultFile) (tea.Cmd, bool) {
	return nil, true
}

func (p *provider) Title() string {
	return p.title
}

func (p *provider) Columns() []table.Column {
	return []table.Column{
		{Title: "File Name", Width: 40},
		{Title: "Created Time", Width: 35},
		{Title: "Type", Width: 10},
		{Title: "Size", Width: 15},
	}
}

func (p *provider) Load() ([]result.ResultFile, error) {
	files, err := result.GetResultFiles(p.state.Config.Writer)
	if err != nil {
		logger.UIError("Failed to load result logs: %v", err)
		return nil, err
	}

	slices.SortFunc(files, func(i, j result.ResultFile) int {
		return j.CreatedTime.Compare(i.CreatedTime)
	})

	logger.UIInfo("Loaded %d result files from disk", len(files))
	return files, nil
}

func (p *provider) RenderRow(item result.ResultFile) table.Row {
	return table.Row{
		item.Name,
		item.CreatedTime.Format("2006-01-02 15:04:05"),
		item.Schema.Name,
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
		return notice.NewNoticeCmd(p.state.Layout, "Delete Failed", err.Error(), notice.NOTICE_ERROR), true
	}
	return nil, true
}

func (p *provider) OnRename(item result.ResultFile, newName string) (tea.Cmd, bool) {
	newName = result.NormalizeResultFileName(newName)
	dstPath := filepath.Join(filepath.Dir(item.Path), newName)
	if err := os.Rename(item.Path, dstPath); err != nil {
		logger.UIError("Failed to rename file on disk: %v", err)
		return notice.NewNoticeCmd(p.state.Layout, "Rename Failed", err.Error(), notice.NOTICE_ERROR), true
	}
	return nil, true
}
