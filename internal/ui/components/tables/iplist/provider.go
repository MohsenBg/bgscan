package iplist

import (
	"os"
	"slices"

	"github.com/MohsenBg/bgscan/internal/core/iplist"
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

func newSource(state *ui.AppState, title string) crud.Source[iplist.IPFileInfo] {
	return &source{
		title: title,
		state: state,
	}
}

func (s *source) Title() string { return s.title }

func (s *source) Columns() []table.Column {
	return []table.Column{
		{Title: "Name", Width: 40},
		{Title: "Created Time", Width: 40},
		{Title: "Size", Width: 20},
	}
}

func (s *source) Load() ([]iplist.IPFileInfo, error) {
	files, err := iplist.ListIPFiles()
	if err != nil {
		s.state.Log.UI.Error("Failed to load IP files: %v", err)
		return nil, err
	}
	s.state.Log.UI.Info("Loaded %d IP files", len(files))

	slices.SortFunc(files, func(i, j iplist.IPFileInfo) int {
		return j.CreatedAt.Compare(i.CreatedAt)
	})
	return files, nil
}

func (s *source) RenderRow(item iplist.IPFileInfo) table.Row {
	return table.Row{
		item.Name,
		item.CreatedAt.Format("2006-01-02 15:04:05"),
		humanize.Bytes(uint64(item.Size)),
	}
}

func (s *source) Identity(item iplist.IPFileInfo) string {
	return item.Name
}

// newHooks wires the mutation behaviors. OnAdd only registers the "add"
// keybinding: the real file-picker workflow is intercepted in Update via
// crud.MsgActionTrigger, so its body never runs.
func newHooks(state *ui.AppState, onSelect func(*iplist.IPFileInfo) tea.Cmd) crud.Hooks[iplist.IPFileInfo] {
	hooks := crud.Hooks[iplist.IPFileInfo]{
		OnDelete: func(item iplist.IPFileInfo) tea.Cmd {
			return func() tea.Msg {
				if err := os.Remove(item.Path); err != nil && !os.IsNotExist(err) {
					state.Log.UI.Error("Failed to delete IP file: %v", err)
					return notice.NewNoticeCmd(state.Deps(), "Delete Failed", err.Error(), notice.NOTICE_ERROR)()
				}
				return nil
			}
		},
		OnRename: func(item iplist.IPFileInfo, newName string) tea.Cmd {
			return func() tea.Msg {
				dstPath, err := iplist.GetIPFilePath(newName)
				if err != nil {
					state.Log.UI.Error("Failed to resolve destination path: %v", err)
					return notice.NewNoticeCmd(state.Deps(), "Rename Failed", err.Error(), notice.NOTICE_ERROR)()
				}

				if err := os.Rename(item.Path, dstPath); err != nil {
					state.Log.UI.Error("Rename failed: %v", err)
					return notice.NewNoticeCmd(state.Deps(), "Rename Failed", err.Error(), notice.NOTICE_ERROR)()
				}
				return nil
			}
		},
		OnAdd: func() tea.Cmd { return nil },
	}

	if onSelect != nil {
		hooks.OnSelect = func(item iplist.IPFileInfo) tea.Cmd {
			return onSelect(&item)
		}
	}

	return hooks
}
