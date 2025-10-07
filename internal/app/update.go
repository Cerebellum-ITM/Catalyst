package app

import (
	"catalyst/internal/types"
	"sort"
	"time"

	"github.com/charmbracelet/glamour"


	"catalyst/internal/app/components/core"
	"catalyst/internal/app/components/statusbar"

	"github.com/charmbracelet/bubbles/v2/key"
	"github.com/charmbracelet/bubbles/v2/list"
	"github.com/charmbracelet/bubbles/v2/textinput"
	tea "github.com/charmbracelet/bubbletea/v2"

)

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd
	m.StatusBar, cmd = m.StatusBar.Update(msg)
	cmds = append(cmds, cmd)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.StatusBar.AppWith = m.width
		m.recalculateSizes()
		return m, nil
	case tea.KeyMsg:
		if m.state == spellbookLoaded {
			m.state = ready
			return m, noDelayClearStatusCmd()
		}
		switch {
		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
		case key.Matches(msg, m.keys.GlobalQuit):
			return m, tea.Quit
		}
	case clearStatusMsg:
		m.StatusBar.Content = m.getDefaultStatusBarContent()
		m.StatusBar.Level = statusbar.LevelInfo
		return m, nil
	case continueToReadyMsg:
		if m.state == spellbookLoaded {
			m.state = ready
		}
		return m, noDelayClearStatusCmd()
	}

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
	case editingRune:
		return updateEditingRune(msg, m)
	case showingHistory:
		return updateShowingHistory(msg, m)
	default:
		return updateInitial(msg, m)
	}
}

func (m *Model) changeFocusedElement() {
	if m.focusedElement == listElement {
		m.focusedElement = viewportElement
	} else {
		m.focusedElement = listElement
	}

	switch m.state {
	case ready:
		switch m.focusedElement {
		case listElement:
			m.keys = mainListKeys()
		case viewportElement:
			m.keys = viewPortKeys()
		}
	case showingRunes:
		switch m.focusedElement {
		case listElement:
			m.keys = viewingRunesKeys()
		case viewportElement:
			m.keys = viewPortKeys()
		}
	}
}

func (m *Model) getDefaultStatusBarContent() string {
	switch m.state {
	case ready:
		return m.SpellbookString
	case showingRunes:
		return "Viewing Runes"
	case creatingRune:
		return "Creating a new Rune"
	case editingRune:
		return "Editing Rune"
	case executingRune:
		return "Executing Rune"
	case showingLoegs:
		return "Viewing Loegs"
	case creatingLoeg:
		return "Creating a new Loeg"
	default:
		return "Ready"
	}
}

type continueToReadyMsg struct{}

func continueToReadyCmd() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return continueToReadyMsg{}
	})
}

// updateInitial handles updates during the initial spellbook check/creation.
func updateInitial(msg tea.Msg, m *Model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		}
	case gotSpellbookMsg:
		m.spellbook = &msg.spellbook
		m.state = spellbookLoaded
		m.StatusBar.Level = statusbar.LevelSuccess
		m.StatusBar.Content = "Ready to start press any key ...."
		m.StatusBar.StopSpinner()
		m.focusedElement = listElement
		m.menuItems.ResetFilter()
		m.menuItems.SetFilterText("")
		m.menuItems.SetFilterState(list.FilterState(list.Filtering))
		return m, continueToReadyCmd()
	case errMsg:
		if msg.err == nil { // This means the spellbook doesn't exist, time to create it
			m.state = creatingSpellbook
			m.StatusBar.Content = "Spellbook not found creating a new one...."
			m.StatusBar.Level = statusbar.LevelFatal
			return m, tea.Batch(m.StatusBar.StartSpinner(), m.createSpellbookCmd)
		}
		m.err = msg.err
		m.state = errState
		m.StatusBar.StopSpinner()
		m.StatusBar.Content = "Error: " + msg.err.Error()
		m.StatusBar.Level = statusbar.LevelError
		return m, clearStatusCmd()
	}
	return m, nil
}

// updateReady handles updates when the main menu is active.
func updateReady(msg tea.Msg, m *Model) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	// Handle global keys and other messages first
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.SwitchFocus):
			m.changeFocusedElement()
			return m, nil
		case key.Matches(msg, m.keys.GlobalQuit):
			return m, tea.Quit
		}
	case gotSpellbookMsg: // Centralized update path
		m.spellbook = &msg.spellbook
		m.state = ready
		m.keys = mainListKeys()
		m.StatusBar.StopSpinner()
		m.StatusBar.Content = "Ready"
		m.StatusBar.Level = statusbar.LevelSuccess
		return m, clearStatusCmd()
	}

	// Then, route messages to the correct component based on focus
	switch m.focusedElement {
	case listElement:
		// Handle list filtering and selection
		if m.menuItems.FilterState() == list.Filtering {
			// Restore the original filter logic here
			if keyMsg, ok := msg.(tea.KeyMsg); ok {
				switch {
				case key.Matches(keyMsg, m.keys.Up):
					m.menuItems.CursorUp()
				case key.Matches(keyMsg, m.keys.Down):
					m.menuItems.CursorDown()
				}
			}
		}

		if keyMsg, ok := msg.(tea.KeyMsg); ok && key.Matches(keyMsg, m.keys.Enter) {
			selectedItem := m.menuItems.SelectedItem()
			if item, ok := selectedItem.(core.MenuItem); ok {
				switch item.Value() {
				case 0: // Get Runes
					// Populate the runes list with items from the spellbook
					items := make([]list.Item, len(m.spellbook.Runes))
					for i, r := range m.spellbook.Runes {
						items[i] = core.RuneItem{Rune: r}
					}
					m.runesList.SetItems(items)

					m.state = showingRunes
					m.keys = viewingRunesKeys()
					m.StatusBar.Content = "Viewing Runes"

					// Initialize viewport with the first rune's details
					if len(m.spellbook.Runes) > 0 {
						md := formatRuneDetail(m.spellbook.Runes[0])
						rendered, _ := glamour.Render(md, "dark")
						m.viewportSpellBook.SetContent(rendered)
					}
					
					// Recalculate sizes for the new layout
					m.recalculateSizes()
					
					return m, nil
				case 1: // Create Rune
					m.state = creatingRune
					m.keys = formKeys()
					m.StatusBar.Content = "Creating a new Rune"
					// Form initialization logic...
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
					m.StatusBar.Content = "Viewing Loegs"
					m.loegKeys = make([]string, 0, len(m.spellbook.Loegs))
					for k := range m.spellbook.Loegs {
						m.loegKeys = append(m.loegKeys, k)
					}
					sort.Strings(m.loegKeys)
					return m, nil
				case 3: // View History
					m.state = showingHistory
					m.StatusBar.Content = "Viewing History"
					return m, m.getHistoryCmd
				}
			}
		}
		m.menuItems, cmd = m.menuItems.Update(msg)
		cmds = append(cmds, cmd)

	case viewportElement:
		m.viewportSpellBook, cmd = m.viewportSpellBook.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// updateShowingRunes handles updates when displaying the list of runes.
func updateShowingRunes(msg tea.Msg, m *Model) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, m.keys.SwitchFocus) {
			m.changeFocusedElement()
			return m, nil
		}

		// When the viewport is focused, handle its scrolling.
		if m.focusedElement == viewportElement {
			if key.Matches(msg, m.keys.Esc) || key.Matches(msg, m.keys.SwitchFocus) {
				m.focusedElement = listElement
				m.keys = viewingRunesKeys()
				return m, nil
			}
			m.viewportSpellBook, cmd = m.viewportSpellBook.Update(msg)
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)
		}

		// Otherwise, handle list navigation and actions.
		switch {
		case key.Matches(msg, m.keys.Esc):
			m.state = ready
			m.keys = mainListKeys()
			m.focusedElement = listElement // Reset focus
			m.StatusBar.Content = m.SpellbookString
			return m, nil

		case key.Matches(msg, m.keys.Enter):
			selectedItem, ok := m.runesList.SelectedItem().(core.RuneItem)
			if !ok {
				return m, nil
			}
			m.previousState = showingRunes
			m.state = executingRune
			m.keys = executingRuneKeys()
			m.StatusBar.Content = "Executing rune..."
			return m, tea.Batch(m.StatusBar.StartSpinner(), m.executeSpecificRuneCmd(selectedItem.Rune))

		case key.Matches(msg, m.keys.Delete):
			if len(m.spellbook.Runes) > 0 {
				m.StatusBar.Content = "Deleting rune..."
				m.StatusBar.Level = statusbar.LevelWarning
				return m, tea.Batch(m.StatusBar.StartSpinner(), m.deleteRuneCmd)
			}

		case key.Matches(msg, m.keys.Edit):
			if len(m.spellbook.Runes) > 0 {
				m.state = editingRune
				m.keys = formKeys()
				m.StatusBar.Content = "Editing rune"
				m.StatusBar.Level = statusbar.LevelInfo
				m.focusIndex = 0

				selectedRuneItem, ok := m.runesList.SelectedItem().(core.RuneItem)
				if !ok {
					return m, nil
				}
				selectedRune := selectedRuneItem.Rune
				
				m.inputs = make([]textinput.Model, 2+len(selectedRune.Commands))

				t := textinput.New()
				t.Placeholder = "Rune Name"
				t.SetValue(selectedRune.Name)
				t.Focus()
				m.inputs[0] = t

				t = textinput.New()
				t.Placeholder = "Description"
				t.SetValue(selectedRune.Description)
				m.inputs[1] = t

				for i, cmd := range selectedRune.Commands {
					t = textinput.New()
					t.Placeholder = "Command"
					t.SetValue(cmd)
					t.CharLimit = 256
					m.inputs[2+i] = t
				}
				return m, textinput.Blink
			}

		// ... other key matches like delete, edit ...
		}
	
	case gotSpellbookMsg: // This case is now primarily for refreshing the data
		m.spellbook = &msg.spellbook
		
		items := make([]list.Item, len(m.spellbook.Runes))
		for i, r := range m.spellbook.Runes {
			items[i] = core.RuneItem{Rune: r}
		}
		m.runesList.SetItems(items)

		if len(m.spellbook.Runes) > 0 {
			md := formatRuneDetail(m.spellbook.Runes[0])
			rendered, _ := glamour.Render(md, "dark")
			m.viewportSpellBook.SetContent(rendered)
		}
		return m, nil // No need for other state changes here
	}

	// Update the list and get commands
	m.runesList, cmd = m.runesList.Update(msg)
	cmds = append(cmds, cmd)

	// When the selected item changes, update the viewport
	if selectedItem, ok := m.runesList.SelectedItem().(core.RuneItem); ok {
		md := formatRuneDetail(selectedItem.Rune)
		rendered, _ := glamour.Render(md, "dark")
		m.viewportSpellBook.SetContent(rendered)
	}

	return m, tea.Batch(cmds...)
}

// updateCreatingRune handles updates for the rune creation form.
func updateCreatingRune(msg tea.Msg, m *Model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Esc):
			m.state = ready
			m.keys = mainListKeys()
			m.StatusBar.Content = m.SpellbookString
			m.StatusBar.Level = statusbar.LevelInfo
			return m, nil
		case key.Matches(msg, m.keys.Enter):
			if m.focusIndex == len(m.inputs) {
				m.StatusBar.Content = "Creating rune..."
				m.StatusBar.Level = statusbar.LevelInfo
				return m, tea.Batch(m.StatusBar.StartSpinner(), m.createRuneCmd)
			}
		}
	case gotSpellbookMsg: // Success message
		m.spellbook = &msg.spellbook
		m.state = showingRunes
		m.keys = viewingRunesKeys()
		m.StatusBar.StopSpinner()
		m.StatusBar.Content = "Successfully created rune"
		m.StatusBar.Level = statusbar.LevelSuccess
		return m, clearStatusCmd()
	case errMsg:
		m.err = msg.err
		m.state = errState
		m.StatusBar.StopSpinner()
		m.StatusBar.Content = "Error: " + msg.err.Error()
		m.StatusBar.Level = statusbar.LevelError
		return m, clearStatusCmd()
	}
	cmd := updateRuneForm(msg, m)
	return m, cmd
}

// updateExecutingRune handles updates while a rune is running.
func updateExecutingRune(msg tea.Msg, m *Model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Esc), key.Matches(msg, m.keys.Enter):
			// Return to the previous state (either showingRunes or showingHistory)
			if m.previousState == showingHistory {
				m.state = showingHistory
				// m.keys = viewingHistoryKeys() // TODO
				m.StatusBar.Content = "Viewing History"
				return m, m.getHistoryCmd // Refresh history view
			}
			m.state = showingRunes
			m.keys = viewingRunesKeys()
			m.StatusBar.Content = "Viewing Runes"
			return m, nil
		}
	case runeExecutedMsg:
		m.output = msg.output
		m.StatusBar.StopSpinner()
		m.StatusBar.Content = "Execution finished"
		m.StatusBar.Level = statusbar.LevelSuccess
		return m, clearStatusCmd()
	}
	return m, nil
}

// updateEditingRune handles the form for updating an existing rune.
func updateEditingRune(msg tea.Msg, m *Model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Esc):
			m.state = showingRunes
			m.keys = viewingRunesKeys()
			m.StatusBar.Content = "Viewing Runes"
			m.StatusBar.Level = statusbar.LevelInfo
			return m, nil
		case key.Matches(msg, m.keys.Enter):
			if m.focusIndex == len(m.inputs) {
				m.StatusBar.Content = "Updating rune..."
				m.StatusBar.Level = statusbar.LevelInfo
				return m, tea.Batch(m.StatusBar.StartSpinner(), m.updateRuneCmd)
			}
		}
	case gotSpellbookMsg: // Success message
		m.spellbook = &msg.spellbook
		m.state = showingRunes
		m.keys = viewingRunesKeys()
		m.StatusBar.StopSpinner()
		m.StatusBar.Content = "Successfully updated rune"
		m.StatusBar.Level = statusbar.LevelSuccess
		return m, clearStatusCmd()
	case noChangesMsg:
		m.state = showingRunes
		m.keys = viewingRunesKeys()
		m.StatusBar.StopSpinner()
		m.StatusBar.Content = "No changes detected"
		m.StatusBar.Level = statusbar.LevelWarning
		return m, clearStatusCmd()
	case errMsg:
		m.err = msg.err
		m.state = errState
		m.StatusBar.StopSpinner()
		m.StatusBar.Content = "Error: " + msg.err.Error()
		m.StatusBar.Level = statusbar.LevelError
		return m, clearStatusCmd()
	}
	cmd := updateRuneForm(msg, m)
	return m, cmd
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
func updateShowingLoegs(msg tea.Msg, m *Model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Esc):
			m.state = ready
			m.keys = mainListKeys()
			m.StatusBar.Content = m.SpellbookString
			m.StatusBar.Level = statusbar.LevelInfo
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
				m.StatusBar.Content = "Deleting loeg..."
				m.StatusBar.Level = statusbar.LevelWarning
				return m, tea.Batch(m.StatusBar.StartSpinner(), m.removeLoegCmd)
			}
		case key.Matches(msg, m.keys.New):
			m.state = creatingLoeg
			m.keys = formKeys()
			m.StatusBar.Content = "Creating a new loeg"
			m.StatusBar.Level = statusbar.LevelInfo
			m.inputs = make([]textinput.Model, 2)
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
		m.StatusBar.StopSpinner()
		m.StatusBar.Content = "Updated loegs list"
		m.StatusBar.Level = statusbar.LevelSuccess
		m.loegKeys = make([]string, 0, len(m.spellbook.Loegs))
		for k := range m.spellbook.Loegs {
			m.loegKeys = append(m.loegKeys, k)
		}
		sort.Strings(m.loegKeys)
		return m, clearStatusCmd()
	case loegRemovedMsg:
		return m, m.getSpellbookContentCmd
	case errMsg:
		m.err = msg.err
		m.state = errState
		m.StatusBar.StopSpinner()
		m.StatusBar.Content = "Error: " + msg.err.Error()
		m.StatusBar.Level = statusbar.LevelError
		return m, clearStatusCmd()
	}
	return m, nil
}

// updateCreatingLoeg handles the form for creating a new loeg.
func updateCreatingLoeg(msg tea.Msg, m *Model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Esc):
			m.state = showingLoegs
			m.keys = viewingLoegsKeys()
			return m, m.getSpellbookContentCmd
		case key.Matches(msg, m.keys.Up), key.Matches(msg, m.keys.Down):
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
			if m.focusIndex == len(m.inputs)-1 || m.focusIndex == len(m.inputs) {
				m.StatusBar.Content = "Setting loeg..."
				m.StatusBar.Level = statusbar.LevelInfo
				return m, tea.Batch(m.StatusBar.StartSpinner(), m.setLoegCmd)
			}
		}
	case gotSpellbookMsg: // Success message
		m.spellbook = &msg.spellbook
		m.state = showingLoegs
		m.keys = viewingLoegsKeys()
		m.StatusBar.StopSpinner()
		m.StatusBar.Content = "Successfully set loeg"
		m.StatusBar.Level = statusbar.LevelSuccess
		return m, clearStatusCmd()
	case errMsg:
		m.err = msg.err
		m.state = errState
		m.StatusBar.StopSpinner()
		m.StatusBar.Content = "Error: " + msg.err.Error()
		m.StatusBar.Level = statusbar.LevelError
		return m, clearStatusCmd()
	}

	cmd := m.updateInputs(msg)
	return m, cmd
}

// updateShowingHistory handles updates when viewing the execution history.
func updateShowingHistory(msg tea.Msg, m *Model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Esc):
			m.state = ready
			m.keys = mainListKeys()
			m.StatusBar.Content = m.SpellbookString
			m.StatusBar.Level = statusbar.LevelInfo
			m.cursor = 0
			return m, nil
		case key.Matches(msg, m.keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, m.keys.Down):
			if m.cursor < len(m.history)-1 {
				m.cursor++
			}
		case key.Matches(msg, m.keys.Enter):
			if m.cursor >= 0 && m.cursor < len(m.history) {
				selectedEntry := m.history[m.cursor]
				var selectedRune *types.Rune
				for i := range m.spellbook.Runes {
					if m.spellbook.Runes[i].Name == selectedEntry.RuneID {
						selectedRune = &m.spellbook.Runes[i]
						break
					}
				}

				if selectedRune != nil {
					m.previousState = showingHistory
					m.state = executingRune
					m.keys = executingRuneKeys()
					m.StatusBar.Content = "Executing rune from history..."
					m.StatusBar.Level = statusbar.LevelInfo
					return m, tea.Batch(m.StatusBar.StartSpinner(), m.executeSpecificRuneCmd(*selectedRune))
				}
				// Optional: Handle case where rune is not found anymore
				m.StatusBar.Content = "Rune from history not found in current spellbook"
				m.StatusBar.Level = statusbar.LevelWarning
				return m, clearStatusCmd()
			}
		}
	case gotHistoryMsg:
		m.history = msg.history
		m.StatusBar.Content = "Viewing history"
		m.StatusBar.Level = statusbar.LevelSuccess
		return m, clearStatusCmd()
	case errMsg:
		m.err = msg.err
		m.state = errState
		m.StatusBar.Content = "Error: " + msg.err.Error()
		m.StatusBar.Level = statusbar.LevelError
		return m, clearStatusCmd()
	}
	return m, nil
}
