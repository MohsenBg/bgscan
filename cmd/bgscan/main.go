package main

import (
	"fmt"
	"log"
	"os"

	"bgscan/internal/core"
	"bgscan/internal/core/config"
	"bgscan/internal/logger"
	"bgscan/internal/startup"
	"bgscan/internal/ui/main/app"

	tea "charm.land/bubbletea/v2"
)

func main() {
	if err := core.Init(); err != nil {
		log.Fatalf("failed to initialize core: %v", err)
	}

	store := config.NewStore()
	cfg, err := store.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	startup.RunHealthChecks(&cfg, &store)

	defer logger.CloseAll()

	p := tea.NewProgram(app.New(&cfg, &store))

	if _, err := p.Run(); err != nil {
		fmt.Printf("BubbleTea runtime error:%s", err.Error())
		os.Exit(1)
	}
}
