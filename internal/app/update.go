package app

import (
	"github.com/charmbracelet/bubbles/v2/textinput"
	tea "github.com/charmbracelet/bubbletea/v2"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.state {
	case ready:
		return updateReady(msg, m)
	case showingRunes:
		return updateShowingRunes(msg, m)
	case creatingRune:
		return updateCreatingRune(msg, m)
	case executingRune:
		return updateExecutingRune(msg, m)
	case showingLoegs:
		return updateShowingLoegs(msg, m)
	case creatingLoeg:
		return updateCreatingLoeg(msg, m)
	default: // Covers checking, creating, error states
		return updateInitial(msg, m)
	}
}

// updateInitial handles updates during the initial spellbook check/creation.
func updateInitial(msg tea.Msg, m Model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			return m, tea.Quit
		}
	case spellbookCheckMsg, spellbookCreateMsg:
		m.state = ready
		return m, nil
	case errMsg:
		if msg.err == nil {
			m.state = creatingSpellbook
			return m, m.createSpellbookCmd
		}
		m.err = msg.err
		m.state = errState
		return m, nil
	}
	return m, nil
}

// updateReady handles updates when the main menu is active.
func updateReady(msg tea.Msg, m Model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.menuItems)-1 {
				m.cursor++
			}
		case "enter":
			switch m.cursor {
			case 0: // Get Runes
				m.state = showingRunes
				return m, m.getRunesCmd
			case 1: // Create Rune
				m.state = creatingRune
				return m, nil
			case 2: // Manage Loegs
				m.state = showingLoegs
				return m, m.getLoegsCmd
			}
		}
	case gotRunesMsg: // This can be received if we come back to 'ready'
		m.runes = msg.runes
		m.state = showingRunes
		return m, nil
	}
	return m, nil
}

// updateShowingRunes handles updates when displaying the list of runes.
func updateShowingRunes(msg tea.Msg, m Model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.state = ready // Go back to the main menu
			m.cursor = 0    // Reset cursor for main menu
			return m, nil
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.runes)-1 {
				m.cursor++
			}
		case "enter":
			if len(m.runes) > 0 {
				m.state = executingRune
				return m, m.executeRuneCmd
			}
		}
	case gotRunesMsg:
		m.runes = msg.runes
		m.cursor = 0 // Reset cursor for rune list
		return m, nil
	case errMsg:
		m.err = msg.err
		m.state = errState
		return m, nil
	}
	return m, nil
}

// updateCreatingRune handles updates for the rune creation form.
func updateCreatingRune(msg tea.Msg, m Model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.state = ready // Go back to the main menu
			return m, nil
		case "tab", "shift+tab", "up", "down":
			s := msg.String()
			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex > len(m.inputs) {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs)
			}

			cmds := make([]tea.Cmd, len(m.inputs))
			for i := 0; i <= len(m.inputs)-1; i++ {
				if i == m.focusIndex {
					cmds[i] = m.inputs[i].Focus()
					continue
				}
				m.inputs[i].Blur()
			}
			return m, tea.Batch(cmds...)

		case "enter":
			if m.focusIndex == len(m.inputs) { // "Submit" is focused
				return m, m.createRuneCmd
			}
		}
	case runeCreatedMsg:
		m.state = ready // Go back to main menu after creation
		return m, nil
	case errMsg:
		m.err = msg.err
		m.state = errState
		return m, nil
	}

	// Update the focused text input
	cmd := m.updateInputs(msg)
	return m, cmd
}

// updateExecutingRune handles updates while a rune is running.
func updateExecutingRune(msg tea.Msg, m Model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc", "enter":
			m.state = showingRunes // Go back to the rune list
			return m, nil
		}
	case runeExecutedMsg:
		m.output = msg.output
		return m, nil
	}
	return m, nil
}

// updateInputs passes messages to the textinput components.
func (m *Model) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}
	return tea.Batch(cmds...)
}

// updateShowingLoegs handles updates when displaying the list of loegs.
func updateShowingLoegs(msg tea.Msg, m Model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.state = ready
			m.cursor = 0
			return m, nil
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.loegKeys)-1 {
				m.cursor++
			}
		case "d": // Delete
			if len(m.loegKeys) > 0 {
				return m, m.removeLoegCmd
			}
		case "n": // New
			m.state = creatingLoeg
			m.inputs = make([]textinput.Model, 2) // KEY and VALUE
			m.focusIndex = 0

			var t textinput.Model
			for i := range m.inputs {
				t = textinput.New()
				t.CharLimit = 128
				switch i {
				case 0:
					t.Placeholder = "KEY"
					t.Focus()
				case 1:
					t.Placeholder = "VALUE"
				}
				m.inputs[i] = t
			}
			return m, textinput.Blink
		}
	case gotLoegsMsg:
		m.loegs = msg.loegs
		m.loegKeys = make([]string, 0, len(m.loegs))
		for k := range m.loegs {
			m.loegKeys = append(m.loegKeys, k)
		}
		m.cursor = 0
		return m, nil
	case loegRemovedMsg:
		// Refresh the list after deleting
		return m, m.getLoegsCmd
	case errMsg:
		m.err = msg.err
		m.state = errState
		return m, nil
	}
	return m, nil
}

// updateCreatingLoeg handles the form for creating a new loeg.
func updateCreatingLoeg(msg tea.Msg, m Model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.state = showingLoegs
			return m, m.getLoegsCmd // Refresh list
		case "tab", "shift+tab", "up", "down":
			s := msg.String()
			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex > len(m.inputs) {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs)
			}

			cmds := make([]tea.Cmd, len(m.inputs))
			for i := 0; i <= len(m.inputs)-1; i++ {
				if i == m.focusIndex {
					cmds[i] = m.inputs[i].Focus()
					continue
				}
				m.inputs[i].Blur()
			}
			return m, tea.Batch(cmds...)

		case "enter":
			// If the user presses enter on the last input field or the submit button,
			// submit the form.
			if m.focusIndex == len(m.inputs)-1 || m.focusIndex == len(m.inputs) {
				return m, m.setLoegCmd
			}
		}
	case loegSetMsg:
		m.state = showingLoegs
		return m, m.getLoegsCmd // Refresh list
	case errMsg:
		m.err = msg.err
		m.state = errState
		return m, nil
	}

	// Handle text input updates
	cmd := m.updateInputs(msg)
	return m, cmd
}
