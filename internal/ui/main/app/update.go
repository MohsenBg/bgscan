package app

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/MohsenBg/bgscan/internal/ui/main/splash"
	"github.com/MohsenBg/bgscan/internal/ui/main/startup"
	"github.com/MohsenBg/bgscan/internal/ui/shared/dialog"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
	"github.com/MohsenBg/bgscan/internal/ui/theme"

	tea "charm.land/bubbletea/v2"
)

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.session.Env().Layout.Update(msg.Width, msg.Height)

	case tea.KeyPressMsg:
		if m.session.Env().Keys.Global.ForceQuit.Matches(msg) {
			return m, tea.Quit
		}

		if m.session.Env().Keys.Global.DumpGoroutines.Matches(msg) {
			m.session.Env().Log.Debug("%s", dumpGoroutines())
		}

		if cmd, closed := m.closeTopDialog(msg); closed {
			return m, cmd
		}

	case splash.SplashDoneMsg:
		cmds = append(cmds, m.splash.OnClose(), m.startup.Init())
		m.stage = StageStartUP

	case startup.StartupChecksDoneMsg:
		cmds = append(cmds, m.startup.OnClose(), m.workspace.Init())
		m.stage = StageWorkspace

	case dialog.OpenDialogMsg:
		m.push(msg)

		cmds = append(cmds, msg.Component.Init(), tickDialogFrames())

		return m, tea.Sequence(cmds...)

	case dialogFrameMsg:
		now := time.Now()

		// Closing layers are removed here rather than by whoever asked for
		// the close, so they stay on the stack long enough to be drawn
		// while they collapse away.
		cmds = append(cmds, m.expire(now)...)

		// Only reappear while something is still on screen. An idle
		// stack schedules no timers at all.
		if m.animating(now) {
			cmds = append(cmds, tickDialogFrames())
		}

		return m, tea.Sequence(cmds...)

	case ui.CloseComponentMsg:
		cmds = append(cmds, m.beginCloseByID(msg.ID))
	}

	if _, themed := msg.(theme.ThemeChanged); themed {
		cmds = append(cmds, m.updateAll(msg)...)
	} else if cmd, active := m.updateTop(msg); active {
		if _, isKey := msg.(tea.KeyMsg); isKey {
			return m, cmd
		}
		cmds = append(cmds, cmd)
	}

	var cmd tea.Cmd
	switch m.stage {
	case StageSplash:
		m.splash, cmd = m.splash.Update(msg)
	case StageStartUP:
		m.startup, cmd = m.startup.Update(msg)
	case StageWorkspace:
		m.workspace, cmd = m.workspace.Update(msg)
	}

	cmds = append(cmds, cmd)
	return m, tea.Sequence(cmds...)
}

// closeTopDialog closes the active dialog if a back or quit key is pressed.
// It returns the close command and a boolean indicating if a close was triggered.
func (m *model) closeTopDialog(msg tea.KeyPressMsg) (tea.Cmd, bool) {
	topLayer := m.top()

	//  nothing to close, or already in the process of closing
	if topLayer == nil || topLayer.closing() {
		return nil, false
	}

	keys := m.session.Env().Keys.Global
	mode := topLayer.component.Mode()

	// Check if the current keypress should trigger a close based on the mode
	switch mode {
	case ui.KeyModeNormal:
		if keys.Back.Matches(msg) || keys.Quit.Matches(msg) {
			return m.beginCloseTop(), true
		}
	case ui.KeyModeInput:
		if keys.BackFromInput.Matches(msg) {
			return m.beginCloseTop(), true
		}
	}

	return nil, false
}

func dumpGoroutines() string {
	buf := make([]byte, 1<<20)
	for {
		n := runtime.Stack(buf, true)
		if n < len(buf) {
			buf = buf[:n]
			break
		}
		buf = make([]byte, len(buf)*2)
	}
	suspiciousStates := []string{
		"[chan receive",
		"[chan send",
		"[select",
		"[IO wait",
		"[sleep",
		"[semacquire",
		"[sync.Mutex.Lock",
	}

	blocks := strings.Split(string(buf), "\n\n")
	var matched []string
	for _, block := range blocks {
		block = strings.TrimSpace(block)
		if block == "" {
			continue
		}
		for _, state := range suspiciousStates {
			if strings.Contains(block, state) {
				matched = append(matched, block)
				break
			}
		}
	}

	total := runtime.NumGoroutine()
	header := fmt.Sprintf("=== Goroutine Dump: %d total, %d suspicious ===\n", total, len(matched))

	if len(matched) == 0 {
		return header + "(none)\n"
	}

	return header + strings.Join(matched, "\n\n") + "\n"
}
