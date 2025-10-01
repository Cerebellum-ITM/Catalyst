package app

import "fmt"

func (m Model) View() string {
	switch m.state {
	case checkingSpellbook:
		return "Checking for Spellbook..."
	case creatingSpellbook:
		return "Creating Spellbook..."
	case ready:
		return "Spellbook is ready! Press 'q' to quit."
	case errState:
		return fmt.Sprintf("An error occurred: %v\n\nPress 'q' to quit.", m.err)
	}
	return "Unknown state."
}
