package main

import (
	"fmt"
	"os"

	"github.com/MohsenBg/bgscan/internal/core/config"
	"github.com/MohsenBg/bgscan/internal/logger"
	"github.com/MohsenBg/bgscan/internal/ui/main/app"
	"github.com/MohsenBg/bgscan/internal/ui/theme"

	tea "charm.land/bubbletea/v2"
)

var Version = "dev"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Println(Version)
		return
	}

	theme.Init()

	defer logger.CloseAll()
	config.AppVersion = Version

	app := app.New()
	p := tea.NewProgram(app)
	app.SetProgram(p)

	if _, err := p.Run(); err != nil {
		fmt.Printf("BubbleTea runtime error:%s", err.Error())
		os.Exit(1)
	}
}
