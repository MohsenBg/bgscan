package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/MohsenBg/bgscan/internal/core/config"
	"github.com/MohsenBg/bgscan/internal/logger"
	"github.com/MohsenBg/bgscan/internal/ui/main/app"
)

var Version = "dev"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Println(Version)
		return
	}

	coreLogger, err := logger.New("core.log")
	if err != nil || coreLogger == nil {
		fmt.Fprintf(os.Stderr, "warning: log files unavailable: %v\n", err)
		coreLogger = logger.Discard()
	}

	uiLogger, err := logger.New("ui.log")
	if err != nil || uiLogger == nil {
		fmt.Fprintf(os.Stderr, "warning: log files unavailable: %v\n", err)
		uiLogger = logger.Discard()
	}

	defer coreLogger.Close()
	defer uiLogger.Close()

	config.AppVersion = Version

	app := app.New(*uiLogger, *coreLogger)
	p := tea.NewProgram(app)
	app.SetProgram(p)

	if _, err := p.Run(); err != nil {
		fmt.Printf("BubbleTea runtime error:%s", err.Error())
		os.Exit(1)
	}
}
