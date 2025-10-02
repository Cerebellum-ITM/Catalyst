package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss/v2"
)

var (
	subtle        = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	highlight     = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	VerticalSpace = lipgloss.NewStyle().Height(1).Render("")
)

func (m Model) View() string {
	var s strings.Builder

	statusBarContent := m.StatusBar.Render()
	helpView := lipgloss.NewStyle().Padding(0, 2).SetString(m.help.View(m.keys)).String()

	switch m.state {
	case checkingSpellbook:
		m.StatusBar.Content = "Checking for Spellbook..."
	case creatingSpellbook:
		s.WriteString("Creating Spellbook...")
	case ready:
		s.WriteString(fmt.Sprintf("Spellbook ready at: %s\n\n", m.pwd))
		s.WriteString("What would you like to do?\n\n")

		for i, item := range m.menuItems {
			cursor := " "
			if m.cursor == i {
				cursor = ">"
			}
			s.WriteString(fmt.Sprintf("%s %s\n", highlight.Render(cursor), item))
		}


	case showingRunes:
		s.WriteString(fmt.Sprintf("Runes in %s:\n\n", m.pwd))
		if len(m.spellbook.Runes) == 0 {
			s.WriteString("No runes found.\n")
		} else {
			for i, r := range m.spellbook.Runes {
				cursor := " "
				if m.cursor == i {
					cursor = ">"
				}
				s.WriteString(fmt.Sprintf("%s %s: %s\n", highlight.Render(cursor), r.Name, r.Description))
			}
		}


	case creatingRune:
		s.WriteString("Create a New Rune\n\n")
		for i := range m.inputs {
			s.WriteString(m.inputs[i].View() + "\n")
		}

		submitButton := "Submit"
		if m.focusIndex == len(m.inputs) {
			submitButton = highlight.Render(submitButton)
		}
		s.WriteString(fmt.Sprintf("\n%s\n", submitButton))


	case editingRune:
		s.WriteString(
			fmt.Sprintf(
				"Editing Rune: %s\n\n",
				highlight.Render(m.spellbook.Runes[m.cursor].Name),
			),
		)
		for i := range m.inputs {
			s.WriteString(m.inputs[i].View() + "\n")
		}

		submitButton := "Submit"
		if m.focusIndex == len(m.inputs) {
			submitButton = highlight.Render(submitButton)
		}
		s.WriteString(fmt.Sprintf("\n%s\n", submitButton))


	case executingRune:
		selectedRune := m.spellbook.Runes[m.cursor]
		s.WriteString(fmt.Sprintf("Executing Rune: %s\n\n", highlight.Render(selectedRune.Name)))
		s.WriteString(m.output)
		// Add a blinking cursor or spinner here in a real app to show activity.
		if m.output == "" {
			s.WriteString("Running commands...")
		}


	case showingLoegs:
		s.WriteString("Loegs (Environment Variables):\n\n")
		if len(m.loegKeys) == 0 {
			s.WriteString("No loegs found.\n")
		} else {
			for i, k := range m.loegKeys {
				cursor := " "
				if m.cursor == i {
					cursor = ">"
				}
				s.WriteString(fmt.Sprintf("%s %s = %s\n", highlight.Render(cursor), k, m.spellbook.Loegs[k]))
			}
		}


	case creatingLoeg:
		s.WriteString("Create a New Loeg\n\n")
		for i := range m.inputs {
			s.WriteString(m.inputs[i].View() + "\n")
		}

		submitButton := "Submit"
		if m.focusIndex == len(m.inputs) {
			submitButton = highlight.Render(submitButton)
		}
		s.WriteString(fmt.Sprintf("\n%s\n", submitButton))


	case errState:
		s.WriteString(fmt.Sprintf("An error occurred: %v\n\n", m.err))

	}

	uiElements := s.String()
	mainContent := lipgloss.JoinVertical(lipgloss.Left,
		statusBarContent,
		VerticalSpace,
		uiElements,
		helpView,
	)

	mainLayer := lipgloss.NewLayer(mainContent)
	canvas := lipgloss.NewCanvas(mainLayer)

	return canvas.Render()
}
