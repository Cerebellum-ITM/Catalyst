package app

import (
	"catalyst/internal/config"

	tea "github.com/charmbracelet/bubbletea/v2"
)

type Model struct {
	config *config.Config
}

func NewModel(cfg *config.Config) Model {
	return Model{config: cfg}
}

func (m Model) Init() tea.Cmd {
	return nil
}
