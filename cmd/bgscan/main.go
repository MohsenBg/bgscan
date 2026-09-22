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

	logs, err := logger.NewSet()
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: log files unavailable: %v\n", err)
		logs = logger.DiscardSet()
	}

	defer logs.Close()

	config.AppVersion = Version

	app := app.New(logs)
	p := tea.NewProgram(app)
	app.SetProgram(p)

	if _, err := p.Run(); err != nil {
		fmt.Printf("BubbleTea runtime error:%s", err.Error())
		os.Exit(1)
	}
}
