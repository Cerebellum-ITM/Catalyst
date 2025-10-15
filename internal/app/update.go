package app

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"catalyst/internal/types"

	"github.com/charmbracelet/log/v2"

	"github.com/charmbracelet/glamour"

	"catalyst/internal/app/components/core"
	"catalyst/internal/app/components/statusbar"
	"catalyst/internal/utils"

	"github.com/charmbracelet/bubbles/v2/key"
	"github.com/charmbracelet/bubbles/v2/list"
	"github.com/charmbracelet/bubbles/v2/textinput"
	tea "github.com/charmbracelet/bubbletea/v2"
	"golang.design/x/clipboard"
)

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	// Always update the status bar.
	m.StatusBar, cmd = m.StatusBar.Update(msg)
	cmds = append(cmds, cmd)

	// Handle component lifecycle messages first. These are high priority.
	switch msg := msg.(type) {
	case core.PopupConfirmedMsg:
		m.popup = nil
		return m, msg.ConfirmCmd
	case HideLockScreenMsg:
		m.lockScreen = nil
		return m, m.getSpellbookContentCmd // This is the new centralized refresh point
	}

	// If a popup is active, it captures all input and blocks other components.
	if m.popup != nil {
		var popupModel tea.Model
		popupModel, cmd = m.popup.Update(msg)
		if newPopupModel, ok := popupModel.(*core.PopupModel); ok {
			m.popup = newPopupModel
		} else {
			m.popup = nil // Should not happen, but good to be safe.
		}
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)
	}

	// If a lock screen is active, it also captures messages for animation,
	// but we must allow completion messages to fall through to the main logic.
	if m.lockScreen != nil {
		// If the lock screen was just created, we need to initialize it.
		// This is where the animation starts.
		if m.lockScreenJustCreated {
			cmds = append(cmds, m.lockScreen.Init())
			m.lockScreenJustCreated = false
		}

		// For messages that are not completion signals, pass them to the lock
		// screen, but don't return. This allows multiple animation messages
		// (e.g., for spinner and progress bar) to be processed in one cycle.
		var lockScreenModel tea.Model
		lockScreenModel, cmd = m.lockScreen.Update(msg)
		if newLockScreenModel, ok := lockScreenModel.(*core.LockScreenModel); ok {
			m.lockScreen = newLockScreenModel
		} else {
			m.lockScreen = nil // Should not happen.
		}
		cmds = append(cmds, cmd)

		// Only return early if the message was NOT a completion signal.
		// Completion signals need to fall through to the main state logic.
		switch msg.(type) {
		case gotSpellbookMsg, errMsg, tea.WindowSizeMsg:
		// Fall through
		default:
			return m, tea.Batch(cmds...)
		}
	}

	// Main application logic.
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.StatusBar.AppWith = m.width
		m.recalculateSizes()
		if m.lockScreen != nil {
			m.lockScreen.Resize(m.width, m.availableHeight)
		}
		// No need to return here, let other components process the size change.
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

	// Route updates to the correct state handler.
	var stateCmd tea.Cmd
	switch m.state {
	case ready:
		_, stateCmd = updateReady(msg, m)
	case showingRunes:
		_, stateCmd = updateShowingRunes(msg, m)
	case executingRune:
		_, stateCmd = updateExecutingRune(msg, m)
	case showingLoegs:
		_, stateCmd = updateShowingLoegs(msg, m)
	case creatingLoeg:
		_, stateCmd = updateCreatingLoeg(msg, m)
	case editingRune:
		_, stateCmd = updateEditingRune(msg, m)
	case showingHistory:
		_, stateCmd = updateShowingHistory(msg, m)
	default:
		_, stateCmd = updateInitial(msg, m)
	}
	cmds = append(cmds, stateCmd)

	return m, tea.Batch(cmds...)
}

func (m *Model) changeFocusedElement() {
	switch m.focusedElement {
	case listElement:
		m.focusedElement = viewportElement
	case formElement:
		m.focusedElement = viewportElement
	case viewportElement:
		switch m.state {
		case editingRune:
			m.focusedElement = formElement
		default:
			m.focusedElement = listElement
		}
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
	case editingRune:
		switch m.focusedElement {
		case formElement:
			m.keys = formKeys()
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
		utils.ResetListFilterState(&m.menuItems)
		return m, continueToReadyCmd()
	case errMsg:
		finalMsg := "An error occurred"
		if msg.err != nil {
			finalMsg = msg.err.Error()
		}

		if m.lockScreen != nil {
			return m, tea.Sequence(
				func() tea.Msg {
					return core.ProgressUpdateMsg{Percent: 1.0, LogLine: finalMsg}
				},
				tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
					return HideLockScreenMsg{}
				}),
			)
		}

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
		finalMsg := "Spellbook updated successfully"
		// If the lock screen is active, it means this is the result of an async operation.
		if m.lockScreen != nil {
			return m, tea.Sequence(
				func() tea.Msg {
					return core.ProgressUpdateMsg{Percent: 1.0, LogLine: finalMsg}
				},
				tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
					return HideLockScreenMsg{}
				}),
			)
		}
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
				case key.Matches(keyMsg, m.keys.ClearFilter):
					utils.ResetListFilterState(&m.menuItems)
					return m, nil
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
					utils.ResetListFilterState(&m.runesList)

					m.state = showingRunes
					m.focusedElement = listElement
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
					m.previousState = m.state
					m.state = editingRune
					m.focusedElement = formElement
					m.keys = formKeys()
					m.keys.AddCommand.SetEnabled(false)
					m.keys.RemoveCommand.SetEnabled(false)
					m.StatusBar.Content = "Creating a new Rune"

					// Clear inputs for new rune entry
					m.inputs = make([]core.CustomTextInput, 3) // name, desc, one command
					var t core.CustomTextInput
					for i := range m.inputs {
						t = core.NewTextInput("", *m.Theme)
						switch i {
						case 0:
							t.Name = "Rune Name"
							t.Model.Placeholder = "Init project"
							t.Model.Focus()
						case 1:
							t.Name = "Description"
							t.Model.Placeholder = "Start the project docker"
						case 2:
							t.Name = "Cmd 1"
							t.Model.Placeholder = "docker compose up -d"
						}
						m.inputs[i] = t
					}
					m.recalculateSizes()
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
		case key.Matches(msg, m.keys.Up):
			m.runesList.CursorUp()
		case key.Matches(msg, m.keys.Down):
			m.runesList.CursorDown()
		case key.Matches(msg, m.keys.ClearFilter):
			utils.ResetListFilterState(&m.runesList)
			return m, nil
		case key.Matches(msg, m.keys.Esc):
			m.state = ready
			m.keys = mainListKeys()
			m.focusedElement = listElement // Reset focus
			m.StatusBar.Content = m.SpellbookString
			return m, nil

		case key.Matches(msg, m.keys.Enter):
			m.previousState = showingRunes
			m.state = executingRune
			m.keys = executingRuneKeys()
			m.executingViewport.SetContent("")
			m.output = ""

			if len(m.executionQueue) > 0 {
				m.StatusBar.Content = "Executing rune queue..."
				m.executionQueueIndex = 0
				m.logsView = core.NewLogsView(m.width/3, m.availableHeight, m.Theme)
				m.focusedElement = logsViewportElement // Set initial focus
				m.recalculateSizes()

				// Save the entire queue to history
				var runeIDs []string
				for _, r := range m.executionQueue {
					runeIDs = append(runeIDs, r.Name)
				}
				if err := m.db.AddHistoryEntry(runeIDs, m.spellbook.Name); err != nil {
					return m, func() tea.Msg { return errMsg{err} }
				}

				return m, tea.Batch(m.StatusBar.StartSpinner(), m.executeQueuedRuneCmd())
			} else {
				selectedItem, ok := m.runesList.SelectedItem().(core.RuneItem)
				if !ok {
					return m, nil
				}

				m.StatusBar.Content = "Executing rune..."
				m.executingRuneName = selectedItem.Rune.Name // Store the name
				m.logsView = core.NewLogsView(m.width/3, m.availableHeight, m.Theme)
				m.focusedElement = logsViewportElement // Set initial focus
				m.recalculateSizes()
				return m, tea.Batch(m.StatusBar.StartSpinner(), m.executeSpecificRuneCmd(selectedItem.Rune, true))
			}

		case key.Matches(msg, m.keys.Delete):
			if len(m.spellbook.Runes) > 0 {
				selectedItem, ok := m.runesList.SelectedItem().(core.RuneItem)
				if !ok {
					return m, nil
				}
				title := "Confirm Deletion"
				message := fmt.Sprintf("Are you sure you want to delete the rune '%s'?", selectedItem.Rune.Name)
				confirmCmd := func() tea.Msg {
					return confirmedDeleteRuneMsg{}
				}
				popup := core.NewPopup(title, message, confirmCmd, m.Theme, m.width, m.height)
				m.popup = &popup
			}

		case key.Matches(msg, m.keys.Edit):
			if len(m.spellbook.Runes) > 0 {
				m.previousState = m.state
				m.state = editingRune
				m.keys = formKeys()
				m.keys.AddCommand.SetEnabled(false)
				m.keys.RemoveCommand.SetEnabled(false)
				m.StatusBar.Content = "Editing rune"
				m.StatusBar.Level = statusbar.LevelInfo
				m.focusIndex = 0

				selectedRuneItem, ok := m.runesList.SelectedItem().(core.RuneItem)
				if !ok {
					return m, nil
				}
				selectedRune := selectedRuneItem.Rune

				m.inputs = make([]core.CustomTextInput, 2+len(selectedRune.Commands))

				var t core.CustomTextInput
				t = core.NewTextInput("", *m.Theme)
				t.Name = "Rune Name"
				t.Model.Placeholder = "Rune Name"
				t.Model.SetValue(selectedRune.Name)
				t.Model.Focus()
				m.inputs[0] = t

				t = core.NewTextInput("", *m.Theme)
				t.Name = "Description"

				t.Model.Placeholder = "Description"
				t.Model.SetValue(selectedRune.Description)
				m.inputs[1] = t

				for i, cmd := range selectedRune.Commands {
					textinputCmdName := fmt.Sprintf("Cmd %d", i+1)
					t = core.NewTextInput(textinputCmdName, *m.Theme)
					t.Model.Placeholder = "Command"
					t.Model.SetValue(cmd)
					m.inputs[2+i] = t
				}
				return m, textinput.Blink
			}

		case key.Matches(msg, m.keys.QueueRune):
			selectedItem, ok := m.runesList.SelectedItem().(core.RuneItem)
			if !ok {
				return m, nil
			}
			selectedRune := selectedItem.Rune

			// Check if the rune is already in the queue
			foundIndex := -1
			for i, r := range m.executionQueue {
				if r.Name == selectedRune.Name {
					foundIndex = i
					break
				}
			}

			if foundIndex != -1 {
				// Remove from the queue
				m.executionQueue = append(m.executionQueue[:foundIndex], m.executionQueue[foundIndex+1:]...)
			} else {
				// Add to the queue
				m.executionQueue = append(m.executionQueue, selectedRune)
			}

			// Update QueuePosition for all items
			queueMap := make(map[string]int)
			for i, r := range m.executionQueue {
				queueMap[r.Name] = i + 1
			}

			listItems := m.runesList.Items()
			for i, item := range listItems {
				runeItem := item.(core.RuneItem)
				if pos, ok := queueMap[runeItem.Rune.Name]; ok {
					runeItem.QueuePosition = pos
				} else {
					runeItem.QueuePosition = 0
				}
				listItems[i] = runeItem
			}
			return m, m.runesList.SetItems(listItems)
		}

	case confirmedDeleteRuneMsg:
		m.lockScreen = core.NewLockScreen(m.width, m.availableHeight, "Deleting Rune...", m.Theme)
		m.lockScreenJustCreated = true
		return m, tea.Sequence(
			func() tea.Msg {
				return core.ProgressUpdateMsg{Percent: 0.3, LogLine: "Deleting rune..."}
			},
			m.deleteRuneCmd,
		)

	case gotSpellbookMsg: // This case is now primarily for refreshing the data
		m.spellbook = &msg.spellbook
		items := make([]list.Item, len(m.spellbook.Runes))
		for i, r := range m.spellbook.Runes {
			items[i] = core.RuneItem{Rune: r}
		}
		m.runesList.SetItems(items)

		finalMsg := "Runes list updated"
		if m.lockScreen != nil {
			return m, tea.Sequence(
				func() tea.Msg {
					return core.ProgressUpdateMsg{Percent: 1.0, LogLine: finalMsg}
				},
				tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
					return HideLockScreenMsg{}
				}),
			)
		}

		m.StatusBar.StopSpinner()
		m.StatusBar.Content = "Updated runes list"
		m.StatusBar.Level = statusbar.LevelSuccess
		utils.ResetListFilterState(&m.runesList)

		if len(m.spellbook.Runes) > 0 {
			md := formatRuneDetail(m.spellbook.Runes[0])
			rendered, _ := glamour.Render(md, "dark")
			m.viewportSpellBook.SetContent(rendered)
		}
		return m, clearStatusCmd()
	case runeDeletedMsg:
		return m, m.getSpellbookContentCmd
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

// updateExecutingRune handles updates while a rune is running.
func updateExecutingRune(msg tea.Msg, m *Model) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.SwitchFocus):
			if m.focusedElement == logsViewportElement {
				m.focusedElement = outputViewportElement
			} else {
				m.focusedElement = logsViewportElement
			}
			return m, nil
		case key.Matches(msg, m.keys.Cancel):
			if m.currentCancelFunc != nil {
				m.currentCancelFunc()
			}
			return m, nil
		case key.Matches(msg, m.keys.Yank):
			if m.focusedElement == logsViewportElement && m.logsView != nil {
				clipboard.Write(clipboard.FmtText, []byte(m.logsView.GetContent()))
				m.StatusBar.Content = "Logs copied to clipboard!"
				m.StatusBar.Level = statusbar.LevelSuccess
				return m, clearStatusCmd()
			}
			if m.focusedElement == outputViewportElement {
				clipboard.Write(clipboard.FmtText, []byte(m.output))
				m.StatusBar.Content = "Output copied to clipboard!"
				m.StatusBar.Level = statusbar.LevelSuccess
				return m, clearStatusCmd()
			}
		case key.Matches(msg, m.keys.Esc), key.Matches(msg, m.keys.Enter):
			if m.currentCancelFunc != nil {
				m.currentCancelFunc()
			}
			if m.previousState == showingHistory {
				m.state = showingHistory
				m.keys = viewingHistoryKeys()
				m.StatusBar.Content = "Viewing History"
				m.output = ""
				m.executingViewport.SetContent("")
				return m, m.getHistoryCmd
			}
			m.state = showingRunes
			m.keys = viewingRunesKeys()
			m.StatusBar.Content = "Viewing Runes"
			return m, nil
		}

	case runNextCommandMsg:
		if m.logsView != nil {
			if m.currentCommandIndex == 0 {
				m.logsView.AddLog(log.DebugLevel, "Execution started", "rune", m.executingRuneName)
			}
			if m.currentCommandIndex < len(m.commandsToExecute) {
				command := m.commandsToExecute[m.currentCommandIndex]
				m.logsView.AddLog(log.InfoLevel, "Executing command", "cmd", command)
			}
		}
		return m, m.executeNextCommandCmd()

	case types.RuneCommandOutputMsg:
		m.output += msg.Output
		m.executingViewport.SetContent(m.output)
		m.executingViewport.GotoBottom()
		return m, waitForOutput(m.msgChan)

	case types.RuneCommandFinished:
		m.currentCancelFunc = nil // Command is done.
		if msg.Err != nil {
			m.logsView.AddLog(log.ErrorLevel, "Command failed, stopping execution", "error", msg.Err)
			if len(m.executionQueue) > 0 {
				m.logsView.AddLog(log.ErrorLevel, "Execution queue stopped due to error")
				m.executionQueue = nil
				m.executionQueueIndex = 0
			}
			m.StatusBar.StopSpinner()
			m.StatusBar.Content = "Execution failed!"
			m.StatusBar.Level = statusbar.LevelError
			return m, clearStatusCmd()
		}

		m.currentCommandIndex++
		if m.currentCommandIndex < len(m.commandsToExecute) {
			return m, func() tea.Msg { return runNextCommandMsg{} }
		}

		m.logsView.AddLog(log.DebugLevel, "Rune finished", "rune", m.executingRuneName)
		if len(m.executionQueue) > 0 {
			m.executionQueueIndex++
			if m.executionQueueIndex < len(m.executionQueue) {
				m.logsView.AddSeparator()
				m.executingViewport.SetContent("")
				m.output = m.output + "\n\n"
				return m, m.executeQueuedRuneCmd()
			}
			m.executionQueue = nil
			m.executionQueueIndex = 0
			m.StatusBar.StopSpinner()
			m.StatusBar.Content = "Execution queue finished"
			m.StatusBar.Level = statusbar.LevelSuccess
			m.logsView.AddLog(log.DebugLevel, "All runes in queue executed successfully")
			return m, clearStatusCmd()
		}

		m.StatusBar.StopSpinner()
		m.StatusBar.Content = "Execution finished"
		m.StatusBar.Level = statusbar.LevelSuccess
		return m, clearStatusCmd()
	}

	var cmd tea.Cmd
	var logsModel tea.Model
	if m.focusedElement == logsViewportElement && m.logsView != nil {
		logsModel, cmd = m.logsView.Update(msg)
		if newLogsModel, ok := logsModel.(*core.LogsViewModel); ok {
			m.logsView = newLogsModel
		}
		cmds = append(cmds, cmd)
	} else {
		m.executingViewport, cmd = m.executingViewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// updateEditingRune handles the form for updating an existing rune.
func updateEditingRune(msg tea.Msg, m *Model) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	handleSubmit := func() {
		isUpdating := m.previousState == showingRunes
		if isUpdating {
			m.lockScreen = core.NewLockScreen(
				m.width,
				m.availableHeight,
				"Updating Rune...",
				m.Theme,
			)
			m.lockScreenJustCreated = true
			cmds = append(cmds, tea.Sequence(
				func() tea.Msg {
					return core.ProgressUpdateMsg{Percent: 0.3, LogLine: "Updating rune..."}
				},
				m.updateRuneCmd,
			))
		} else {
			m.lockScreen = core.NewLockScreen(m.width, m.availableHeight, "Creating Rune...", m.Theme)
			m.lockScreenJustCreated = true
			cmds = append(cmds, tea.Sequence(
				func() tea.Msg {
					return core.ProgressUpdateMsg{Percent: 0.3, LogLine: "Creating rune..."}
				},
				m.createRuneCmd,
			))
		}
	}

	switch msg := msg.(type) {
	case gotSpellbookMsg:
		// The rune has been updated and the cache has been refreshed.
		// Now we update the model with the new data.
		m.spellbook = &msg.spellbook
		items := make([]list.Item, len(m.spellbook.Runes))
		for i, r := range m.spellbook.Runes {
			items[i] = core.RuneItem{Rune: r}
		}
		m.runesList.SetItems(items)

		// Transition the app state back to the rune list.
		m.state = showingRunes
		m.keys = viewingRunesKeys()
		m.cursor = 0
		utils.ResetListFilterState(&m.runesList)

		// The lock screen is still active. Update it to show the final success
		// message, and then schedule it to close after a short delay.
		// This is the command that will finally unlock the UI.
		return m, tea.Sequence(
			func() tea.Msg {
				return core.ProgressUpdateMsg{Percent: 1.0, LogLine: "Rune operation successful"}
			},
			tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
				return HideLockScreenMsg{}
			}),
		)
	case noChangesMsg:
		m.state = showingRunes
		m.keys = viewingRunesKeys()
		m.StatusBar.Content = "No changes were made"
		m.StatusBar.Level = statusbar.LevelInfo
		return m, clearStatusCmd()
	}
	// First, allow the focused input to process the message.
	// This is crucial for typing and cursor blinking.
	cmd = m.updateInputs(msg)
	cmds = append(cmds, cmd)

	// Then, handle navigation and actions based on key presses.
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(keyMsg, m.keys.Esc):
			m.state = showingRunes
			m.keys = viewingRunesKeys()
			return m, nil

		case key.Matches(keyMsg, m.keys.Up):
			m.focusIndex--
			if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs)
			}
			m.keys.AddCommand.SetEnabled(m.focusIndex >= 2)
			m.keys.RemoveCommand.SetEnabled(m.focusIndex >= 2)
			m.keys.MoveCmdUp.SetEnabled(m.focusIndex > 2)
			m.keys.MoveCmdDown.SetEnabled(m.focusIndex >= 2 && m.focusIndex < len(m.inputs)-1)
			navCmds := make([]tea.Cmd, len(m.inputs))
			for i := range m.inputs {
				if i == m.focusIndex {
					navCmds[i] = m.inputs[i].Focus()
				} else {
					m.inputs[i].Blur()
				}
			}
			cmds = append(cmds, tea.Batch(navCmds...))

		case key.Matches(keyMsg, m.keys.Down):
			m.focusIndex++
			if m.focusIndex > len(m.inputs) {
				m.focusIndex = 0
			}
			m.keys.AddCommand.SetEnabled(m.focusIndex >= 2)
			m.keys.RemoveCommand.SetEnabled(m.focusIndex >= 2)
			m.keys.MoveCmdUp.SetEnabled(m.focusIndex >= 2 && m.focusIndex > 2)
			m.keys.MoveCmdDown.SetEnabled(m.focusIndex >= 2 && m.focusIndex < len(m.inputs)-1)
			navCmds := make([]tea.Cmd, len(m.inputs))
			for i := range m.inputs {
				if i == m.focusIndex {
					navCmds[i] = m.inputs[i].Focus()
				} else {
					m.inputs[i].Blur()
				}
			}
			cmds = append(cmds, tea.Batch(navCmds...))

		case key.Matches(keyMsg, m.keys.MoveCmdUp):
			if m.focusIndex > 2 {
				m.inputs[m.focusIndex], m.inputs[m.focusIndex-1] = m.inputs[m.focusIndex-1], m.inputs[m.focusIndex]
				m.focusIndex--
				for i := 2; i < len(m.inputs); i++ {
					m.inputs[i].Name = fmt.Sprintf("Cmd %d", i-1)
				}
				m.keys.MoveCmdUp.SetEnabled(m.focusIndex > 2)
				m.keys.MoveCmdDown.SetEnabled(m.focusIndex < len(m.inputs)-1)
			}

		case key.Matches(keyMsg, m.keys.MoveCmdDown):
			if m.focusIndex >= 2 && m.focusIndex < len(m.inputs)-1 {
				m.inputs[m.focusIndex], m.inputs[m.focusIndex+1] = m.inputs[m.focusIndex+1], m.inputs[m.focusIndex]
				m.focusIndex++
				for i := 2; i < len(m.inputs); i++ {
					m.inputs[i].Name = fmt.Sprintf("Cmd %d", i-1)
				}
				m.keys.MoveCmdUp.SetEnabled(m.focusIndex > 2)
				m.keys.MoveCmdDown.SetEnabled(m.focusIndex < len(m.inputs)-1)
			}

		case key.Matches(keyMsg, m.keys.AddCommand):
			if m.focusIndex >= 2 {
				textinputCmdName := fmt.Sprintf("Cmd %d", (len(m.inputs) - 1))
				newInput := core.NewTextInput(textinputCmdName, *m.Theme)
				newInput.Model.Placeholder = "Command"
				newInput.Model.Focus()
				newIndex := m.focusIndex + 1
				m.inputs = append(
					m.inputs[:newIndex],
					append([]core.CustomTextInput{newInput}, m.inputs[newIndex:]...)...)
				m.inputs[m.focusIndex].Model.Blur()
				m.focusIndex = newIndex
				cmds = append(cmds, textinput.Blink)
			}
		case key.Matches(keyMsg, m.keys.RemoveCommand):
			if m.focusIndex >= 2 && len(m.inputs) > 3 {
				m.inputs = append(m.inputs[:m.focusIndex], m.inputs[m.focusIndex+1:]...)
				if m.focusIndex >= len(m.inputs) {
					m.focusIndex = len(m.inputs) - 1
				}
				for i, input := range m.inputs {
					if i >= 2 {
						textinputCmdName := fmt.Sprintf("Cmd %d", i-1)
						input.Name = textinputCmdName
						m.inputs[i] = input
					}
				}
				m.inputs[m.focusIndex].Focus()
				cmds = append(cmds, textinput.Blink)
			}

		case key.Matches(keyMsg, m.keys.Enter):
			if m.focusIndex == len(m.inputs) {
				handleSubmit()
			}
		case key.Matches(keyMsg, m.keys.submit):
			m.focusIndex = len(m.inputs)
			handleSubmit()
		}
	}

	// Update the preview viewport regardless of the message type
	tempRune := types.Rune{
		Name:        m.inputs[0].Value(),
		Description: m.inputs[1].Value(),
	}
	var runeCmds []string
	for i := 2; i < len(m.inputs); i++ {
		if val := m.inputs[i].Value(); val != "" {
			runeCmds = append(runeCmds, val)
		}
	}
	tempRune.Commands = runeCmds

	md := formatRuneDetail(tempRune)
	rendered, _ := glamour.Render(md, "dark")
	m.formViewport.SetContent(rendered)

	return m, tea.Batch(cmds...)
}

// updateInputs passes messages to the textinput components.
func (m *Model) updateInputs(msg tea.Msg) tea.Cmd {
	if m.focusIndex >= len(m.inputs) {
		return nil
	}
	var cmd tea.Cmd
	m.inputs[m.focusIndex], cmd = m.inputs[m.focusIndex].Update(msg)
	return cmd
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
				key := m.loegKeys[m.cursor]
				title := "Confirm Deletion"
				message := fmt.Sprintf("Are you sure you want to delete the loeg '%s'?", key)
				confirmCmd := func() tea.Msg {
					m.lockScreen = core.NewLockScreen(m.width, m.availableHeight, "Deleting Loeg...", m.Theme)
					m.lockScreenJustCreated = true
					return tea.Sequence(
						func() tea.Msg {
							return core.ProgressUpdateMsg{Percent: 0.3, LogLine: "Deleting loeg..."}
						},
						m.removeLoegCmd,
					)()
				}
				popup := core.NewPopup(title, message, confirmCmd, m.Theme, m.width, m.height)
				m.popup = &popup
			}
		case key.Matches(msg, m.keys.New):
			m.state = creatingLoeg
			m.keys = formKeys()
			m.StatusBar.Content = "Creating a new loeg"
			m.StatusBar.Level = statusbar.LevelInfo
			m.inputs = make([]core.CustomTextInput, 2)
			m.focusIndex = 0

			var t core.CustomTextInput
			for i := range m.inputs {
				textinputLoegName := fmt.Sprintf("Loeg %d", i+1)
				t = core.NewTextInput(textinputLoegName, *m.Theme)
				switch i {
				case 0:
					t.Model.Placeholder = "KEY"
					t.Model.Focus()
				case 1:
					t.Model.Placeholder = "VALUE"
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
		finalMsg := "Loegs list updated"
		if m.lockScreen != nil {
			return m, tea.Sequence(
				func() tea.Msg {
					return core.ProgressUpdateMsg{Percent: 1.0, LogLine: finalMsg}
				},
				tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
					return HideLockScreenMsg{}
				}),
			)
		}

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
		finalMsg := "An error occurred"
		if msg.err != nil {
			finalMsg = msg.err.Error()
		}
		if m.lockScreen != nil {
			return m, tea.Sequence(
				func() tea.Msg {
					return core.ProgressUpdateMsg{Percent: 1.0, LogLine: finalMsg}
				},
				tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
					return HideLockScreenMsg{}
				}),
			)
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

// executeQueuedRuneCmd prepares and executes the current rune in the queue.
func (m *Model) executeQueuedRuneCmd() tea.Cmd {
	if m.executionQueueIndex < len(m.executionQueue) {
		runeToExecute := m.executionQueue[m.executionQueueIndex]
		if m.logsView != nil {
			m.logsView.AddLog(
				log.InfoLevel,
				"--- PREPARING NEXT RUNE ---",
				"index",
				m.executionQueueIndex,
				"rune",
				runeToExecute.Name,
			)
		}
		m.executingRuneName = runeToExecute.Name
		return m.executeSpecificRuneCmd(runeToExecute, false)
	}
	return nil
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
				m.lockScreen = core.NewLockScreen(m.width, m.availableHeight, "Setting Loeg...", m.Theme)
				m.lockScreenJustCreated = true
				return m, tea.Sequence(
					func() tea.Msg {
						return core.ProgressUpdateMsg{Percent: 0.3, LogLine: "Setting loeg..."}
					},
					m.setLoegCmd,
				)
			}
		}
	case gotSpellbookMsg: // Success message
		m.spellbook = &msg.spellbook
		m.state = showingLoegs
		m.keys = viewingLoegsKeys()
		finalMsg := "Successfully set loeg"
		if m.lockScreen != nil {
			return m, tea.Sequence(
				func() tea.Msg {
					return core.ProgressUpdateMsg{Percent: 1.0, LogLine: finalMsg}
				},
				tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
					return HideLockScreenMsg{}
				}),
			)
		}
		m.StatusBar.StopSpinner()
		m.StatusBar.Content = "Successfully set loeg"
		m.StatusBar.Level = statusbar.LevelSuccess
		return m, clearStatusCmd()
	case errMsg:
		finalMsg := "An error occurred"
		if msg.err != nil {
			finalMsg = msg.err.Error()
		}
		if m.lockScreen != nil {
			return m, tea.Sequence(
				func() tea.Msg {
					return core.ProgressUpdateMsg{Percent: 1.0, LogLine: finalMsg}
				},
				tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
					return HideLockScreenMsg{}
				}),
			)
		}
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
				runeIDs := strings.Split(selectedEntry.RuneID, ",")

				// Find all runes from the spellbook
				var runesToExecute []types.Rune
				for _, runeID := range runeIDs {
					found := false
					for _, r := range m.spellbook.Runes {
						if r.Name == runeID {
							runesToExecute = append(runesToExecute, r)
							found = true
							break
						}
					}
					if !found {
						m.StatusBar.Content = fmt.Sprintf("Rune '%s' from history not found", runeID)
						m.StatusBar.Level = statusbar.LevelWarning
						return m, clearStatusCmd()
					}
				}

				if len(runesToExecute) > 0 {
					m.previousState = showingHistory
					m.state = executingRune
					m.keys = executingRuneKeys()
					m.executingViewport.SetContent("")
					m.output = ""
					m.logsView = core.NewLogsView(m.width/3, m.availableHeight, m.Theme)
					m.focusedElement = logsViewportElement
					m.recalculateSizes()

					if len(runesToExecute) > 1 {
						// It's a queue
						m.StatusBar.Content = "Executing rune queue from history..."
						m.executionQueue = runesToExecute
						m.executionQueueIndex = 0
						var runeIDs []string
						for _, r := range m.executionQueue {
							runeIDs = append(runeIDs, r.Name)
						}
						if err := m.db.AddHistoryEntry(runeIDs, m.spellbook.Name); err != nil {
							return m, func() tea.Msg { return errMsg{err} }
						}

						return m, tea.Batch(m.StatusBar.StartSpinner(), m.executeQueuedRuneCmd())
					} else {
						// It's a single rune
						m.StatusBar.Content = "Executing rune from history..."
						m.executingRuneName = runesToExecute[0].Name
						return m, tea.Batch(m.StatusBar.StartSpinner(), m.executeSpecificRuneCmd(runesToExecute[0], true))
					}
				}
			}
		}
	case gotHistoryMsg:
		m.history = msg.history
		m.StatusBar.Content = "Viewing history"
		m.StatusBar.Level = statusbar.LevelSuccess
		return m, clearStatusCmd()
	case errMsg:
		m.lockScreen = nil
		m.err = msg.err
		m.state = errState
		m.StatusBar.Content = "Error: " + msg.err.Error()
		m.StatusBar.Level = statusbar.LevelError
		return m, clearStatusCmd()
	}
	return m, nil
}
