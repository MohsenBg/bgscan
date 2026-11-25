// Package resultlist shows the saved scan result files in a CRUD table.
// The user can rename, delete, and pick a file. Picking a file opens its
// results in the IP viewer, capped at the maxIPs limit given to New.
package resultlist

import (
	"fmt"

	"github.com/MohsenBg/bgscan/internal/core/result"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/crud"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/notice"
	ipviewer "github.com/MohsenBg/bgscan/internal/ui/components/tables/ipviewer"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

// Model is the result file table component. It wraps a CRUD table of
// result files.
type Model struct {
	id        ui.ComponentID
	name      string
	session   *ui.Session
	maxIPs    uint32
	crudTable *crud.Model[result.ResultFile]
}

// New creates the result file table bound to session with the given
// table title. maxIPs caps how many results OpenResultIP loads from a
// picked file. A nil onSelect means picking a file opens it in the IP
// viewer. A non-nil onSelect is called with the picked file instead.
func New(session *ui.Session, title string, maxIPs uint32, onSelect func(*result.ResultFile) tea.Cmd) *Model {
	m := &Model{
		id:      ui.NewComponentID(),
		name:    "Result Files",
		session: session,
		maxIPs:  maxIPs,
	}

	if onSelect == nil {
		onSelect = m.defaultSelectHandler
	}

	m.crudTable = crud.New(session.Env(), newSource(session, title), newHooks(session, onSelect), 100)

	return m
}

// Init starts the underlying CRUD table and returns its first command.
func (m *Model) Init() tea.Cmd { return m.crudTable.Init() }

// ID returns the component ID assigned at creation.
func (m *Model) ID() ui.ComponentID { return m.id }

// Name returns the component display name.
func (m *Model) Name() string { return m.name }

// OnClose forwards the close to the underlying CRUD table.
func (m *Model) OnClose() tea.Cmd { return m.crudTable.OnClose() }

// Mode returns the key mode of the underlying CRUD table.
func (m *Model) Mode() ui.KeyMode { return m.crudTable.Mode() }

// defaultSelectHandler reads the picked file and opens its results in
// the IP viewer. It is the select handler New uses when onSelect is nil.
func (m *Model) defaultSelectHandler(file *result.ResultFile) tea.Cmd {
	ips, err := result.ReadResultFile(file.Path, file.Schema)
	if err != nil {
		return notice.NewNoticeCmd(m.session.Env(), "Selection", err.Error(), notice.NOTICE_ERROR)
	}
	return m.OpenResultIP(ips)
}

// OpenResultIP loads IP results from a result file and opens the IP viewer.
func (m *Model) OpenResultIP(file result.ResultFile) tea.Cmd {
	coreLog := m.session.CoreLog()
	ips, err := result.LoadAll(&coreLog, file.Path, file.Schema, m.maxIPs)
	if err != nil {
		return notice.NewNoticeCmd(m.session.Env(),
			"Result File Error",
			fmt.Sprintf("Error while reading result file: %v", err),
			notice.NOTICE_ERROR,
		)
	}

	return func() tea.Msg {
		return ui.OpenComponentMsg{
			Component: ipviewer.New(
				m.session.Env(),
				fmt.Sprintf("IP Scan [%s]", file.Schema.Name),
				ips,
				file.Schema,
				ipviewer.WithMaxRow(m.session.Env().Settings.MaxRow),
			),
		}
	}
}
