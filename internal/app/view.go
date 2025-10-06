package app

import (
	"fmt"
	"strings"

	"catalyst/internal/ascii"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/styles"
	"github.com/charmbracelet/lipgloss/v2"
)

var (
	highlight       = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	VerticalSpace   = lipgloss.NewStyle().Height(1).Render("")
	HorizontalSpace = lipgloss.NewStyle().Width(10).Render("")
)

func (m Model) showProntMessage(availableHeightForMainContent int) string {
	var prontMessage string
	asciiLogo := ascii.PrintLogo()
	specsText := ascii.PrintSpecs()

	finalPrompt := lipgloss.JoinHorizontal(
		lipgloss.Left,
		HorizontalSpace,
		asciiLogo,
		HorizontalSpace,
		specsText,
	)
	prontMessage = lipgloss.Place(
		m.width,
		availableHeightForMainContent,
		lipgloss.Left,
		lipgloss.Center,
		finalPrompt,
	)

	return prontMessage
}

func (m Model) View() string {
	var s strings.Builder
	var stateView string

	statusBarContent := m.StatusBar.Render()
	helpView := lipgloss.NewStyle().Padding(0, 2).SetString(m.help.View(m.keys)).String()
	contentHeight := m.height
	statusBarH := lipgloss.Height(statusBarContent)
	VerticalSpaceH := 2 * lipgloss.Height(VerticalSpace)
	helpViewH := lipgloss.Height(helpView)
	availableHeightForMainContent := contentHeight - statusBarH - VerticalSpaceH - helpViewH
	printMessage := m.showProntMessage(availableHeightForMainContent)

	switch m.state {
	case checkingSpellbook:
		m.StatusBar.Content = "Checking for Spellbook..."
		s.WriteString(printMessage)
	case creatingSpellbook:
		m.StatusBar.Content = "Creating Spellbook..."
		s.WriteString(printMessage)
	case spellbookLoaded:
		s.WriteString(printMessage)
	case ready:
		m.viewportSpellBook.SetWidth(m.width * 3 / 4)
		m.viewportSpellBook.SetHeight(availableHeightForMainContent)
		glamourContent := "Test text"
		glamourStyle := styles.DarkStyleConfig
		rendererGlamour, _ := glamour.NewTermRenderer(
			glamour.WithStyles(glamourStyle),
			glamour.WithWordWrap(m.viewportSpellBook.Width()),
		)
		glamourContentStr, _ := rendererGlamour.Render(glamourContent)
		m.viewportSpellBook.SetContent(glamourContentStr)
		m.menuItems.SetWidth(m.width / 4)
		m.menuItems.SetHeight(availableHeightForMainContent)
		stateView = lipgloss.JoinHorizontal(
			lipgloss.Left,
			m.menuItems.View(),
			m.viewportSpellBook.View(),
		)
		s.WriteString(stateView)

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

	case showingHistory:
		s.WriteString("Execution History:\n\n")
		if len(m.history) == 0 {
			s.WriteString("No history found.\n")
		} else {
			for i, entry := range m.history {
				cursor := " "
				if m.cursor == i {
					cursor = ">"
				}
				s.WriteString(fmt.Sprintf("%s %s on %s at %s\n",
					highlight.Render(cursor),
					entry.RuneID,
					entry.SpellbookID,
					entry.ExecutedAt.Format("2006-01-02 15:04:05")),
				)
			}
		}

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
