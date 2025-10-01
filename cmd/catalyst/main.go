package main

import (
	"catalyst/internal/app"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea/v2"
)

func main() {
	m := app.NewModel()
	p := tea.NewProgram(m)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
