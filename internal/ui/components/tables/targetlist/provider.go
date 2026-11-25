package targetlist

import (
	"slices"

	"github.com/MohsenBg/bgscan/internal/core/fileutil"
	"github.com/MohsenBg/bgscan/internal/core/target"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/crud"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/notice"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/table"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
	"github.com/dustin/go-humanize"
)

// source adapts the target package to the crud.Source interface.
type source struct {
	title   string
	session *ui.Session
}

func newSource(session *ui.Session, title string) crud.Source[target.FileInfo] {
	return &source{
		title:   title,
		session: session,
	}
}

func (s *source) Title() string { return s.title }

func (s *source) Name() string {
	return "target file"
}

func (s *source) Columns() []table.Column {
	return []table.Column{
		{Title: "Name", Width: 36},
		{Title: "Type", Width: 10},
		{Title: "Created Time", Width: 36},
		{Title: "Size", Width: 18},
	}
}

func (s *source) Load() ([]target.FileInfo, error) {
	files, err := target.ListFiles()
	if err != nil {
		s.session.Env().Log.Error("Failed to load target files: %v", err)
		return nil, err
	}
	s.session.Env().Log.Info("Loaded %d target files", len(files))

	slices.SortFunc(files, func(i, j target.FileInfo) int {
		return j.CreatedAt.Compare(i.CreatedAt)
	})
	return files, nil
}

func (s *source) RenderRow(item target.FileInfo) table.Row {
	return table.Row{
		item.Name,
		item.Kind.String(),
		item.CreatedAt.Format("2006-01-02 15:04:05"),
		humanize.Bytes(uint64(item.Size)),
	}
}

// Identity prefixes the name with the kind because an IP list and a domain
// list can share the same name.
func (s *source) Identity(item target.FileInfo) string {
	return item.Kind.String() + ":" + item.Name
}

// newHooks wires the mutation behaviors. OnAdd only registers the "add"
// keybinding: the real file-picker workflow is intercepted in Update via
// crud.MsgActionTrigger, so its body never runs.
func newHooks(session *ui.Session, onSelect func(*target.FileInfo) tea.Cmd, add func() tea.Cmd) crud.Hooks[target.FileInfo] {
	hooks := crud.Hooks[target.FileInfo]{
		OnDelete: func(item target.FileInfo) tea.Cmd {
			return func() tea.Msg {
				if err := fileutil.RemoveFile(item.Path); err != nil {
					session.Env().Log.Error("Failed to delete target file: %v", err)
					return notice.NewNoticeCmd(session.Env(), "Delete Failed", err.Error(), notice.NOTICE_ERROR)()
				}
				return nil
			}
		},
		OnRename: func(item target.FileInfo, newName string) tea.Cmd {
			return func() tea.Msg {
				dstPath, err := target.ResolveTargetFilePath(newName, item.Kind)
				if err != nil {
					session.Env().Log.Error("Failed to resolve destination path: %v", err)
					return notice.NewNoticeCmd(session.Env(), "Rename Failed", err.Error(), notice.NOTICE_ERROR)()
				}

				if err := fileutil.ReplaceFile(item.Path, dstPath); err != nil {
					session.Env().Log.Error("Rename failed: %v", err)
					return notice.NewNoticeCmd(session.Env(), "Rename Failed", err.Error(), notice.NOTICE_ERROR)()
				}
				return nil
			}
		},
		OnAdd: add,
	}

	if onSelect != nil {
		hooks.OnSelect = func(item target.FileInfo) tea.Cmd {
			return onSelect(&item)
		}
	}

	return hooks
}
