package main

import (
	"catalyst/internal/app"
	"catalyst/internal/config"
	"catalyst/internal/db"
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/log"
)

const version = "0.1.0"

func main() {
	versionFlag := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("Catalyst version %s\n", version)
		os.Exit(0)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("could not load config: %v", err)
	}

	db, err := db.InitDB()
	if err != nil {
		log.Fatalf("could not initialize database: %v", err)
	}
	defer db.Close()

	m := app.NewModel(cfg, db, version)
	p := tea.NewProgram(m)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
