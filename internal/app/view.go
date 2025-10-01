package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	subtle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	highlight = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
)

func (m Model) View() string {
	var s strings.Builder

	switch m.state {
	case checkingSpellbook:
		s.WriteString("Checking for Spellbook...")
	case creatingSpellbook:
		s.WriteString("Creating Spellbook...")
	case ready:
		s.WriteString(fmt.Sprintf("🔮 Spellbook ready at: %s\n\n", m.pwd))
		s.WriteString("What would you like to do?\n\n")

		for i, item := range m.menuItems {
			cursor := " "
			if m.cursor == i {
				cursor = ">"
			}
			s.WriteString(fmt.Sprintf("%s %s\n", highlight.Render(cursor), item))
		}
		s.WriteString(subtle.Render("\n(Use arrows, enter to select, q to quit)\n"))

	case showingRunes:
		s.WriteString(fmt.Sprintf("📜 Runes in %s:\n\n", m.pwd))
		if len(m.runes) == 0 {
			s.WriteString("No runes found.\n")
		} else {
			for i, r := range m.runes {
				cursor := " "
				if m.cursor == i {
					cursor = ">"
				}
				s.WriteString(fmt.Sprintf("%s %s: %s\n", highlight.Render(cursor), r.Name, r.Description))
			}
		}
		s.WriteString(subtle.Render("\n(enter: run, e: edit, d: delete, esc: back)\n"))

	case creatingRune:
		s.WriteString("✨ Create a New Rune\n\n")
		for i := range m.inputs {
			s.WriteString(m.inputs[i].View() + "\n")
		}

		submitButton := "Submit"
		if m.focusIndex == len(m.inputs) {
			submitButton = highlight.Render(submitButton)
		}
		s.WriteString(fmt.Sprintf("\n%s\n", submitButton))
		s.WriteString(subtle.Render("\n(Use tab to navigate, enter to submit, esc to cancel)\n"))

	case editingRune:
		s.WriteString(fmt.Sprintf("✍️ Editing Rune: %s\n\n", highlight.Render(m.runes[m.cursor].Name)))
		for i := range m.inputs {
			s.WriteString(m.inputs[i].View() + "\n")
		}

		submitButton := "Submit"
		if m.focusIndex == len(m.inputs) {
			submitButton = highlight.Render(submitButton)
		}
		s.WriteString(fmt.Sprintf("\n%s\n", submitButton))
		s.WriteString(subtle.Render("\n(Use tab to navigate, enter to save, esc to cancel)\n"))

	case executingRune:
		selectedRune := m.runes[m.cursor]
		s.WriteString(fmt.Sprintf("🏃 Executing Rune: %s\n\n", highlight.Render(selectedRune.Name)))
		s.WriteString(m.output)
		// Add a blinking cursor or spinner here in a real app to show activity.
		if m.output == "" {
			s.WriteString("Running commands...")
		}
		s.WriteString(subtle.Render("\n(Press enter or esc to return to the rune list)\n"))

	case showingLoegs:
		s.WriteString("🌿 Loegs (Environment Variables):\n\n")
		if len(m.loegKeys) == 0 {
			s.WriteString("No loegs found.\n")
		} else {
			for i, k := range m.loegKeys {
				cursor := " "
				if m.cursor == i {
					cursor = ">"
				}
				s.WriteString(fmt.Sprintf("%s %s = %s\n", highlight.Render(cursor), k, m.loegs[k]))
			}
		}
		s.WriteString(subtle.Render("\n(n: new, d: delete, esc: back)\n"))

	case creatingLoeg:
		s.WriteString("🌿 Create a New Loeg\n\n")
		for i := range m.inputs {
			s.WriteString(m.inputs[i].View() + "\n")
		}

		submitButton := "Submit"
		if m.focusIndex == len(m.inputs) {
			submitButton = highlight.Render(submitButton)
		}
		s.WriteString(fmt.Sprintf("\n%s\n", submitButton))
		s.WriteString(subtle.Render("\n(Use tab to navigate, enter to submit, esc to cancel)\n"))

	case errState:
		s.WriteString(fmt.Sprintf("An error occurred: %v\n\n", m.err))
		s.WriteString(subtle.Render("(Press 'q' to quit)\n"))
	}

	return s.String()
}
