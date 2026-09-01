package app

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/MohsenBg/bgscan/internal/logger"
	"github.com/MohsenBg/bgscan/internal/ui/main/splash"
	"github.com/MohsenBg/bgscan/internal/ui/main/startup"
	"github.com/MohsenBg/bgscan/internal/ui/shared/env"

	tea "charm.land/bubbletea/v2"
)

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.state.Layout.Update(msg.Width, msg.Height)

	case tea.KeyPressMsg:
		if msg.String() == env.KeyCtrlC {
			return m, tea.Quit
		}

		if msg.String() == env.KeyCtrlT {
			logger.DebugInfo("%s", dumpGoroutines())
		}

	case splash.SplashDoneMsg:
		cmds = append(cmds, m.splash.OnClose(), m.startup.Init())
		m.stage = StageStartUP

	case startup.StartupChecksDoneMsg:
		cmds = append(cmds, m.startup.OnClose(), m.workspace.Init())
		m.stage = StageWorkspace
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
