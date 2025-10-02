package app

import (
	"sort"

	"github.com/charmbracelet/bubbles/v2/key"
	"github.com/charmbracelet/bubbles/v2/textinput"
	tea "github.com/charmbracelet/bubbletea/v2"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd
	m.StatusBar, cmd = m.StatusBar.Update(msg)
	cmds = append(cmds, cmd)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		var cmd tea.Cmd
		m.width = msg.Width
		m.height = msg.Height
		m.StatusBar.AppWith = m.width
		return m, cmd
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
		}
	}

	var subCmd tea.Cmd
	var subModel tea.Model

	switch m.state {
	case ready:
		subModel, subCmd = updateReady(msg, m)
	case showingRunes:
		subModel, subCmd = updateShowingRunes(msg, m)
	case creatingRune:
		subModel, subCmd = updateCreatingRune(msg, m)
	case executingRune:
		subModel, subCmd = updateExecutingRune(msg, m)
	case showingLoegs:
		subModel, subCmd = updateShowingLoegs(msg, m)
	case creatingLoeg:
		subModel, subCmd = updateCreatingLoeg(msg, m)
	case editingRune:
		subModel, subCmd = updateEditingRune(msg, m)
	default: // Covers checking, creating, error states
		subModel, subCmd = updateInitial(msg, m)
	}
	cmds = append(cmds, subCmd)
	return subModel, tea.Batch(cmds...)
}

// updateInitial handles updates during the initial spellbook check/creation.
func updateInitial(msg tea.Msg, m Model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			return m, tea.Quit
		}
	case gotSpellbookMsg:
		m.spellbook = &msg.spellbook
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
		switch {
		case key.Matches(msg, m.keys.GlobalQuit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, m.keys.Down):
			if m.cursor < len(m.menuItems)-1 {
				m.cursor++
			}
		case key.Matches(msg, m.keys.Enter):
			switch m.cursor {
			case 0: // Get Runes
				m.state = showingRunes
				m.keys = viewingRunesKeys()
				m.cursor = 0
				return m, nil
			case 1: // Create Rune
				m.state = creatingRune
				m.keys = formKeys()
				m.focusIndex = 0

				// Initialize form for creating a rune
				m.inputs = make([]textinput.Model, 3) // name, desc, one command
				var t textinput.Model
				for i := range m.inputs {
					t = textinput.New()
					t.CharLimit = 256
					switch i {
					case 0:
						t.Placeholder = "Rune Name"
						t.Focus()
					case 1:
						t.Placeholder = "Description"
					case 2:
						t.Placeholder = "Command"
					}
					m.inputs[i] = t
				}
				return m, textinput.Blink
			case 2: // Manage Loegs
				m.state = showingLoegs
				m.keys = viewingLoegsKeys()
				m.cursor = 0
				m.loegKeys = make([]string, 0, len(m.spellbook.Loegs))
				for k := range m.spellbook.Loegs {
					m.loegKeys = append(m.loegKeys, k)
				}
				sort.Strings(m.loegKeys)
				return m, nil
			}
		}
	case gotSpellbookMsg: // This is the new centralized update path
		m.spellbook = &msg.spellbook
		m.state = ready // Go back to ready screen after any CRUD
		m.keys = mainListKeys()
		return m, nil
	}
	return m, nil
}

// updateShowingRunes handles updates when displaying the list of runes.
func updateShowingRunes(msg tea.Msg, m Model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Esc):
			m.state = ready
			m.keys = mainListKeys()
			m.cursor = 0
			return m, nil
		case key.Matches(msg, m.keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, m.keys.Down):
			if m.cursor < len(m.spellbook.Runes)-1 {
				m.cursor++
			}
		case key.Matches(msg, m.keys.Enter):
			if len(m.spellbook.Runes) > 0 {
				m.state = executingRune
				m.keys = executingRuneKeys()
				return m, m.executeRuneCmd
			}
		case key.Matches(msg, m.keys.Delete):
			if len(m.spellbook.Runes) > 0 {
				return m, m.deleteRuneCmd
			}
		case key.Matches(msg, m.keys.Edit):
			if len(m.spellbook.Runes) > 0 {
				m.state = editingRune
				m.keys = formKeys()
				m.focusIndex = 0

				selectedRune := m.spellbook.Runes[m.cursor]

				// Start with 3 inputs: name, desc, first command
				m.inputs = make([]textinput.Model, 2+len(selectedRune.Commands))

				// Setup Name
				t := textinput.New()
				t.Placeholder = "Rune Name"
				t.SetValue(selectedRune.Name)
				t.Focus()
				m.inputs[0] = t

				// Setup Description
				t = textinput.New()
				t.Placeholder = "Description"
				t.SetValue(selectedRune.Description)
				m.inputs[1] = t

				// Setup Command fields
				for i, cmd := range selectedRune.Commands {
					t = textinput.New()
					t.Placeholder = "Command"
					t.SetValue(cmd)
					t.CharLimit = 256
					m.inputs[2+i] = t
				}
				return m, textinput.Blink
			}
		}
	case gotSpellbookMsg:
		m.spellbook = &msg.spellbook
		m.state = showingRunes
		m.keys = viewingRunesKeys()
		m.cursor = 0
		return m, nil
	case runeDeletedMsg:
		// After a rune is deleted, refresh the list.
		return m, m.getSpellbookContentCmd
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
		switch {
		case key.Matches(msg, m.keys.Esc):
			m.state = ready
			m.keys = mainListKeys()
			return m, nil
		case key.Matches(msg, m.keys.Enter):
			// Only submit if the "Submit" button is focused.
			if m.focusIndex == len(m.inputs) {
				return m, m.createRuneCmd
			}
		}
	case runeCreatedMsg:
		m.state = ready
		m.keys = mainListKeys()
		return m, m.getSpellbookContentCmd // Refresh runes list
	case errMsg:
		m.err = msg.err
		m.state = errState
		return m, nil
	}

	// Handle form logic
	return updateRuneForm(msg, m)
}

// updateExecutingRune handles updates while a rune is running.
func updateExecutingRune(msg tea.Msg, m Model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Esc), key.Matches(msg, m.keys.Enter):
			m.state = showingRunes
			m.keys = viewingRunesKeys()
			return m, nil
		}
	case runeExecutedMsg:
		m.output = msg.output
		return m, nil
	}
	return m, nil
}

// updateEditingRune handles the form for updating an existing rune.
func updateEditingRune(msg tea.Msg, m Model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Esc):
			m.state = showingRunes
			m.keys = viewingRunesKeys()
			return m, m.getSpellbookContentCmd
		case key.Matches(msg, m.keys.Enter):
			// Only submit if the "Submit" button is focused.
			if m.focusIndex == len(m.inputs) {
				return m, m.updateRuneCmd
			}
		}
	case runeUpdatedMsg:
		m.state = showingRunes
		m.keys = viewingRunesKeys()
		return m, m.getSpellbookContentCmd
	case errMsg:
		m.err = msg.err
		m.state = errState
		return m, nil
	}

	// Handle form logic
	return updateRuneForm(msg, m)
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
		switch {
		case key.Matches(msg, m.keys.Esc):
			m.state = ready
			m.keys = mainListKeys()
			m.cursor = 0
			return m, nil
		case key.Matches(msg, m.keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, m.keys.Down):
			if m.cursor < len(m.loegKeys)-1 {
				m.cursor++
			}
		case key.Matches(msg, m.keys.Delete):
			if len(m.loegKeys) > 0 {
				return m, m.removeLoegCmd
			}
		case key.Matches(msg, m.keys.New):
			m.state = creatingLoeg
			m.keys = formKeys()
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
	case gotSpellbookMsg:
		m.spellbook = &msg.spellbook
		m.state = showingLoegs
		m.keys = viewingLoegsKeys()
		m.cursor = 0
		m.loegKeys = make([]string, 0, len(m.spellbook.Loegs))
		for k := range m.spellbook.Loegs {
			m.loegKeys = append(m.loegKeys, k)
		}
		sort.Strings(m.loegKeys)
		return m, nil
	case loegRemovedMsg:
		return m, m.getSpellbookContentCmd
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
		switch {
		case key.Matches(msg, m.keys.Esc):
			m.state = showingLoegs
			m.keys = viewingLoegsKeys()
			return m, m.getSpellbookContentCmd // Refresh list
		case key.Matches(msg, m.keys.Up), key.Matches(msg, m.keys.Down):
			// Handle tab focus logic
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

		case key.Matches(msg, m.keys.Enter):
			// If the user presses enter on the last input field or the submit button,
			// submit the form.
			if m.focusIndex == len(m.inputs)-1 || m.focusIndex == len(m.inputs) {
				return m, m.setLoegCmd
			}
		}
	case loegSetMsg:
		m.state = showingLoegs
		m.keys = viewingLoegsKeys()
		return m, m.getSpellbookContentCmd // Refresh list
	case errMsg:
		m.err = msg.err
		m.state = errState
		return m, nil
	}

	// Handle text input updates
	cmd := m.updateInputs(msg)
	return m, cmd
}
