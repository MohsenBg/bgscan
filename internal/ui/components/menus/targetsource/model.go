// Package targetsource lets the user choose a target list source before starting a scan.
package targetsource

import (
	"github.com/MohsenBg/bgscan/internal/core/result"
	"github.com/MohsenBg/bgscan/internal/core/target"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/menu"
	"github.com/MohsenBg/bgscan/internal/ui/components/menus/scantype"
	resultlistTable "github.com/MohsenBg/bgscan/internal/ui/components/tables/resultlist"
	targetlistTable "github.com/MohsenBg/bgscan/internal/ui/components/tables/targetlist"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

// Model is the target source selection component.
type Model struct {
	id      ui.ComponentID
	name    string
	session *ui.Session
	menu    ui.Component
}

// New creates the target source menu.
func New(session *ui.Session) *Model {
	m := &Model{
		session: session,
		id:      ui.NewComponentID(),
		name:    "Target Source",
	}
	items := []menu.MenuItem{
		menu.NewMenuItem("▣", "Target List", "i", func() tea.Msg {
			return m.OpenIPList(func(i *target.FileInfo) tea.Cmd {
				return ui.OpenComponentCmd(scantype.New(session, i.Path))
			})
		}).WithDescription("Pick a list of IPs or domains to scan"),
		menu.NewMenuItem("▤", "Result List", "r", func() tea.Msg {
			return m.OpenResultIPList(func(r *result.ResultFile) tea.Cmd {
				return ui.OpenComponentCmd(scantype.New(session, r.Path))
			})
		}).WithDescription("Reuse IPs from a previous scan result"),
	}

	m.menu = menu.New(items, "Select Target Source", session.Env())
	return m
}

// OpenIPList opens the target file picker overlay.
// onSelect is called by the iplist component once the user picks a file.
func (m *Model) OpenIPList(onSelect func(*target.FileInfo) tea.Cmd) tea.Msg {
	return ui.OpenComponentMsg{Component: targetlistTable.New(m.session, "Select Target File", onSelect)}
}

// OpenResultIPList opens the ResultIP file picker overlay.
// onSelect is called by the resultlist component once the user picks a file.
func (m *Model) OpenResultIPList(onSelect func(*result.ResultFile) tea.Cmd) tea.Msg {
	var maxRenderIP uint32 = 10_000
	return ui.OpenComponentMsg{Component: resultlistTable.New(m.session, "Select IP Result File", maxRenderIP, onSelect)}
}

func (m *Model) Init() tea.Cmd {
	return m.menu.Init()
}

func (m *Model) ID() ui.ComponentID {
	return m.id
}

func (m *Model) Name() string {
	return m.name
}

func (m *Model) OnClose() tea.Cmd {
	return nil
}

func (m *Model) Mode() ui.KeyMode {
	return ui.KeyModeNormal
}
