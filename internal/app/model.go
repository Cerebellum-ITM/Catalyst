package app

import tea "github.com/charmbracelet/bubbletea/v2"

type Model struct{}

func NewModel() Model {
	return Model{}
}

func (m Model) Init() tea.Cmd {
	return nil
}
