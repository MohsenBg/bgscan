package resultlist

import (
	"os"
	"path/filepath"
	"slices"

	"github.com/MohsenBg/bgscan/internal/core/result"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/crud"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/notice"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/table"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
	"github.com/dustin/go-humanize"
)

type source struct {
	title string
	state *ui.AppState
}

func newSource(state *ui.AppState, title string) crud.Source[result.ResultFile] {
	return &source{
		title: title,
		state: state,
	}
}

func (s *source) Title() string {
	return s.title
}

func (s *source) Columns() []table.Column {
	return []table.Column{
		{Title: "File Name", Width: 40},
		{Title: "Created Time", Width: 35},
		{Title: "Type", Width: 10},
		{Title: "Size", Width: 15},
	}
}

func (s *source) Load() ([]result.ResultFile, error) {
	files, err := result.GetResultFiles(s.state.Config.Writer)
	if err != nil {
		s.state.Log.UI.Error("Failed to load result logs: %v", err)
		return nil, err
	}

	slices.SortFunc(files, func(i, j result.ResultFile) int {
		return j.CreatedTime.Compare(i.CreatedTime)
	})

	s.state.Log.UI.Info("Loaded %d result files from disk", len(files))
	return files, nil
}

func (s *source) RenderRow(item result.ResultFile) table.Row {
	return table.Row{
		item.Name,
		item.CreatedTime.Format("2006-01-02 15:04:05"),
		item.Schema.Name,
		humanize.Bytes(uint64(item.SizeBytes)),
	}
}

func (s *source) Identity(item result.ResultFile) string {
	return item.Name
}

// newHooks wires the mutation behaviors. There is no add flow for result
// files, so OnAdd stays nil and no "add" keybinding is registered.
func newHooks(state *ui.AppState, onSelect func(*result.ResultFile) tea.Cmd) crud.Hooks[result.ResultFile] {
	hooks := crud.Hooks[result.ResultFile]{
		OnDelete: func(item result.ResultFile) tea.Cmd {
			if err := os.Remove(item.Path); err != nil && !os.IsNotExist(err) {
				state.Log.UI.Error("Failed to delete result log file: %v", err)
				return notice.NewNoticeCmd(state.Deps(), "Delete Failed", err.Error(), notice.NOTICE_ERROR)
			}
			return nil
		},
		OnRename: func(item result.ResultFile, newName string) tea.Cmd {
			newName = result.NormalizeResultFileName(newName)
			dstPath := filepath.Join(filepath.Dir(item.Path), newName)
			if err := os.Rename(item.Path, dstPath); err != nil {
				state.Log.UI.Error("Failed to rename file on disk: %v", err)
				return notice.NewNoticeCmd(state.Deps(), "Rename Failed", err.Error(), notice.NOTICE_ERROR)
			}
			return nil
		},
	}

	if onSelect != nil {
		hooks.OnSelect = func(item result.ResultFile) tea.Cmd {
			return onSelect(&item)
		}
	}

	return hooks
}
