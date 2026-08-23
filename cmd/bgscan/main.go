package main

import (
	"fmt"
	"os"

	"bgscan/internal/logger"
	"bgscan/internal/ui/main/app"
	"bgscan/internal/ui/theme"

	tea "charm.land/bubbletea/v2"
)

func main() {
	theme.Init()
	// startup.RunHealthChecks(&cfg, &store)

	defer logger.CloseAll()

	app := app.New()
	p := tea.NewProgram(app)
	app.SetProgram(p)

	if _, err := p.Run(); err != nil {
		fmt.Printf("BubbleTea runtime error:%s", err.Error())
		os.Exit(1)
	}
}
