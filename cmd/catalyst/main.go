package main

import (
	"catalyst/internal/app"
	"catalyst/internal/config"
	"catalyst/internal/db"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/log"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("could not load config: %v", err)
	}

	db, err := db.InitDB()
	if err != nil {
		log.Fatalf("could not initialize database: %v", err)
	}
	defer db.Close()

	m := app.NewModel(cfg, db)
	p := tea.NewProgram(m)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
