package app

import tea "github.com/charmbracelet/bubbletea/v2"

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	// Handle key presses
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}

	// Handle successful spellbook check
	case spellbookCheckMsg:
		m.state = ready
		// In a real app, we would parse msg.spellbookJSON here.
		return m, nil

	// Handle successful spellbook creation
	case spellbookCreateMsg:
		m.state = ready
		// In a real app, we would parse msg.spellbookJSON here.
		return m, nil

	// Handle errors
	case errMsg:
		if msg.err == nil { // This is our signal to create the spellbook
			m.state = creatingSpellbook
			return m, m.createSpellbookCmd
		}
		m.err = msg.err
		m.state = errState
		return m, nil
	}

	return m, nil
}
