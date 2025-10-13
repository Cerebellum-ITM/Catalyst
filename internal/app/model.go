package app

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"catalyst/internal/app/components/core"
	"catalyst/internal/app/components/statusbar"
	"catalyst/internal/app/styles"
	"catalyst/internal/config"
	"catalyst/internal/db"
	"catalyst/internal/local"
	"catalyst/internal/ssh"
	"catalyst/internal/types"
	"catalyst/internal/utils"

	"github.com/charmbracelet/bubbles/v2/help"
	"github.com/charmbracelet/bubbles/v2/list"
	"github.com/charmbracelet/bubbles/v2/viewport"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
)

// Define application states.
type state int

const (
	checkingSpellbook state = iota
	creatingSpellbook
	spellbookLoaded
	ready
	showingRunes
	creatingRune
	executingRune
	showingLoegs
	creatingLoeg
	editingRune
	showingHistory
	errState
	demo
)

type focusableElement int

const (
	formElement focusableElement = iota
	listElement
	viewportElement
	logsViewportElement
	outputViewportElement
)

// Define messages for async operations.
type (
	gotSpellbookMsg struct{ spellbook Spellbook }
	runeCreatedMsg  struct{}
	executeNextCommandMsg struct{}
	runeExecutedMsg struct{ output string }
	gotLoegsMsg     struct{ loegs map[string]string }
	loegSetMsg      struct{}
	loegRemovedMsg  struct{}
	runeUpdatedMsg  struct{}
	runeDeletedMsg  struct{}
	gotHistoryMsg   struct{ history []db.HistoryEntry }
	noChangesMsg    struct{} // Message to indicate no changes were made
	clearStatusMsg struct{}
	ClosePopupMsg  core.ClosePopupMsg
	demoFinishedMsg   struct{}
	runDemoStepMsg    struct{}
	confirmedDeleteRuneMsg struct{}
	errMsg                 struct{ err error }
)


type HideLockScreenMsg struct{}

// Command to clear the status bar after a delay
func clearStatusCmd() tea.Cmd {
	return tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
		return clearStatusMsg{}
	})
}

func noDelayClearStatusCmd() tea.Cmd {
	return tea.Tick(time.Second*0, func(t time.Time) tea.Msg {
		return clearStatusMsg{}
	})
}

// Model is the main application model.
type Model struct {
	availableHeight   int
	keys              KeyMap
	help              help.Model
	sshClient         *ssh.Client
	localRunner       *local.Runner
	db                *db.Database
	state             state
	previousState     state
	focusedElement    focusableElement
	pwd               string // Current working directory (spellbook path)
	menuItems         list.Model
	runesList         list.Model
	cursor            int
	spellbook         *Spellbook // Our in-memory cache
	viewportSpellBook viewport.Model
	formViewport      viewport.Model
	executingViewport viewport.Model
	loegKeys          []string               // For ordered display and selection
	history           []db.HistoryEntry      // For the history view
	inputs            []core.CustomTextInput // For the "Create Rune" form
	focusIndex        int
	output                  string // To store output from executed runes
	err                     error
	lockScreen              *core.LockScreenModel
	logsView                *core.LogsViewModel
	lockScreenJustCreated   bool
	popup                   *core.PopupModel
	demoStep                int
	demoCounter             int
	width                   int
	height                  int
	StatusBar               statusbar.StatusBar
	Theme                   *styles.Theme
	Version                 string
	SpellbookString         string

	// For sequential command execution
	commandsToExecute   []string
	currentCommandIndex int
	aggregatedOutput    string
	executingRuneName   string
}

// NewModel creates a new application model.
func NewModel(cfg *config.Config, db *db.Database, version string) Model {
	pwd, _ := os.Getwd() // Get PWD once at the start

	theme := styles.NewCharmtoneTheme()
	help := help.New()
	help.Styles = theme.AppStyles().Help
	initialsKeys := mainListKeys()
	statusbar := statusbar.New(
		"initializing Catalyst",
		statusbar.LevelWarning,
		50,
		theme,
		version,
	)
	statusbar.ShowSpinner = true

	m := Model{
		help:              help,
		keys:              initialsKeys,
		sshClient:         ssh.NewClient(cfg.RuneCraftHost),
		localRunner:       local.NewRunner(),
		db:                db,
		state:             checkingSpellbook,
		pwd:               pwd,
		menuItems:         core.NewMainMenu(*theme),
		runesList:         core.NewRunesList(*theme, []types.Rune{}),
		inputs:            make([]core.CustomTextInput, 3), // name, desc, cmds
		focusIndex:        0,
		Theme:             theme,
		StatusBar:         statusbar,
		Version:           version,
		SpellbookString:   fmt.Sprintf("Main Menu - %s", utils.TruncatePath(pwd, 2)),
		viewportSpellBook: viewport.New(),
		formViewport:      viewport.New(),
		executingViewport: viewport.New(),
	}

	// Initialize text inputs for the create rune form
	var t core.CustomTextInput
	for i := range m.inputs {
		t = core.NewTextInput("", *theme)

		switch i {
		case 0:
			t.Name = "Rune Name"
			t.Model.Placeholder = "Rune Name"
			t.Model.Focus()
		case 1:
			t.Name = "Description"
			t.Model.Placeholder = "Description"
		case 2:
			t.Name = "Cmd"
			t.Model.Placeholder = "Commands (semicolon-separated)"
		}
		m.inputs[i] = t
	}

	return m
}

type recalcOptions struct {
	extraContent int
}

type RecalcOption func(*recalcOptions)

func WithExtraContent(height int) RecalcOption {
	return func(opts *recalcOptions) {
		opts.extraContent = height
	}
}

func (m *Model) recalculateSizes(options ...RecalcOption) {
	if m.width == 0 || m.height == 0 {
		return
	}

	config := &recalcOptions{
		extraContent: 0,
	}

	for _, option := range options {
		option(config)
	}

	// Calculate available height for main content
	statusBarContent := m.StatusBar.Render()
	helpView := lipgloss.NewStyle().Padding(0, 2).SetString(m.help.View(m.keys)).String()
	statusBarH := lipgloss.Height(statusBarContent)
	helpViewH := lipgloss.Height(helpView)
	// 2 for vertical spaces between status bar, content, and help
	availableHeightForMainContent := m.height - statusBarH - helpViewH - 2 - config.extraContent

	// Set sizes for components
	switch m.state {
	case showingRunes:
		m.runesList.SetWidth(m.width / 3)
		m.runesList.SetHeight(availableHeightForMainContent)
		m.viewportSpellBook.SetWidth(m.width * 2 / 3)
		m.viewportSpellBook.SetHeight(availableHeightForMainContent)
	case editingRune:
		for i := range m.inputs {
			m.inputs[i].SetWidth(m.width / 2)
		}
		m.formViewport.SetWidth(m.width / 2)
		m.formViewport.SetHeight(availableHeightForMainContent)
	case executingRune:
		if m.logsView != nil {
			m.logsView.Resize(m.width/3, availableHeightForMainContent)
		}
		m.executingViewport.SetWidth(m.width * 2 / 3)
		m.executingViewport.SetHeight(availableHeightForMainContent)
	default:
		m.viewportSpellBook.SetWidth(m.width * 3 / 4)
		m.viewportSpellBook.SetHeight(availableHeightForMainContent)
		m.menuItems.SetWidth(m.width / 4)
		m.menuItems.SetHeight(availableHeightForMainContent)
	}
	m.availableHeight = availableHeightForMainContent
}

// getSpellbookContentCmd fetches the entire spellbook content.
func (m *Model) getSpellbookContentCmd() tea.Msg {
	cmd := fmt.Sprintf("get-spellbook-content %q", m.pwd)
	jsonStr, err := m.sshClient.Command(cmd)
	if err != nil {
		// A failure here might mean the spellbook doesn't exist.
		return errMsg{err: nil} // Signal to create it.
	}

	var sb Spellbook
	if err := json.Unmarshal([]byte(jsonStr), &sb); err != nil {
		return errMsg{err}
	}
	return gotSpellbookMsg{spellbook: sb}
}

// createSpellbookCmd now also fetches the content after creation.
func (m *Model) createSpellbookCmd() tea.Msg {
	cmd := fmt.Sprintf("create-spellbook %q", m.pwd)
	jsonStr, err := m.sshClient.Command(cmd)
	if err != nil {
		return errMsg{err}
	}

	var sb Spellbook
	if err := json.Unmarshal([]byte(jsonStr), &sb); err != nil {
		// Even if creation is successful, we might fail to parse.
		// Let's return the raw JSON in this edge case.
		return errMsg{fmt.Errorf("failed to parse created spellbook: %w", err)}
	}
	return gotSpellbookMsg{spellbook: sb}
}

// All CRUD operations will now just trigger a full refresh of the spellbook.
// This simplifies state management significantly.

// createRuneCmd sends the command to create a new rune.
func (m *Model) createRuneCmd() tea.Msg {
	// ... (build cmdsStr as before)
	name := m.inputs[0].Value()
	desc := m.inputs[1].Value()

	var cmds []string
	for i := 2; i < len(m.inputs); i++ {
		if val := m.inputs[i].Value(); val != "" {
			cmds = append(cmds, val)
		}
	}
	cmdsStr := strings.Join(cmds, ";")

	if name == "" || desc == "" || cmdsStr == "" {
		return errMsg{fmt.Errorf("name, description, and at least one command are required")}
	}

	cmd := fmt.Sprintf("create-rune %q -name %q -desc %q -cmds %q", m.pwd, name, desc, cmdsStr)
	_, err := m.sshClient.Command(cmd)
	if err != nil {
		return errMsg{err}
	}
	return tea.Sequence(
		func() tea.Msg {
			return core.ProgressUpdateMsg{Percent: 0.6, LogLine: "Refreshing spellbook..."}
		},
		m.getSpellbookContentCmd,
	)()
}

// executeRuneCmd runs the commands for a selected rune from the main list.
func (m *Model) executeRuneCmd() tea.Msg {
	selectedItem, ok := m.runesList.SelectedItem().(core.RuneItem)
	if !ok {
		return errMsg{fmt.Errorf("invalid rune selection")}
	}
	return m.executeSpecificRuneCmd(selectedItem.Rune)()
}

// executeSpecificRuneCmd sets up the model for sequential command execution.
func (m *Model) executeSpecificRuneCmd(r types.Rune) tea.Cmd {
	return func() tea.Msg {
		// Save to history first
		if err := m.db.AddHistoryEntry(r.Name, m.spellbook.Name); err != nil {
			return errMsg{err}
		}

		// Set up the state for sequential execution
		m.commandsToExecute = r.Commands
		m.currentCommandIndex = 0
		m.aggregatedOutput = ""

		// Start the execution of the first command
		return executeNextCommandMsg{}
	}
}

// executeNextCommandCmd executes the current command in the sequence.
func (m *Model) executeNextCommandCmd() tea.Msg {
	if m.currentCommandIndex >= len(m.commandsToExecute) {
		// No more commands to run, signal completion
		return runeExecutedMsg{output: m.aggregatedOutput}
	}

	command := m.commandsToExecute[m.currentCommandIndex]
	output, err := m.localRunner.ExecuteCommand(command)
	m.aggregatedOutput += output + "\n"
	m.currentCommandIndex++

	if err != nil {
		// If a command fails, stop the sequence and report the error
		return runeExecutedMsg{output: m.aggregatedOutput}
	}

	// Continue to the next command
	return executeNextCommandMsg{}
}

// getHistoryCmd retrieves the execution history from the database.
func (m *Model) getHistoryCmd() tea.Msg {
	history, err := m.db.GetHistory()
	if err != nil {
		return errMsg{err}
	}
	return gotHistoryMsg{history: history}
}

// No need for getLoegsCmd anymore, it's part of getSpellbookContentCmd

// setLoegCmd creates or updates a loeg.
func (m *Model) setLoegCmd() tea.Msg {
	// ... (build command as before)
	key := m.inputs[0].Value()
	val := m.inputs[1].Value()
	if key == "" || val == "" {
		return errMsg{fmt.Errorf("key and value are required")}
	}

	arg := fmt.Sprintf("%s=\"%s\"", key, val)
	cmd := fmt.Sprintf("loeg set %q %s", m.pwd, arg)
	_, err := m.sshClient.Command(cmd)
	if err != nil {
		return errMsg{err}
	}
	return tea.Sequence(
		func() tea.Msg {
			return core.ProgressUpdateMsg{Percent: 0.6, LogLine: "Refreshing spellbook..."}
		},
		m.getSpellbookContentCmd,
	)()
}

// removeLoegCmd removes a loeg.
func (m *Model) removeLoegCmd() tea.Msg {
	if m.cursor < 0 || m.cursor >= len(m.loegKeys) {
		return errMsg{fmt.Errorf("invalid loeg selection")}
	}
	key := m.loegKeys[m.cursor]
	// ... (build command as before)
	cmd := fmt.Sprintf("loeg rm %q %s", m.pwd, key)
	_, err := m.sshClient.Command(cmd)
	if err != nil {
		return errMsg{err}
	}
	return tea.Sequence(
		func() tea.Msg {
			return core.ProgressUpdateMsg{Percent: 0.6, LogLine: "Refreshing spellbook..."}
		},
		m.getSpellbookContentCmd,
	)()
}

// updateRuneCmd sends the command to update an existing rune.
func (m *Model) updateRuneCmd() tea.Msg {
	selectedItem, ok := m.runesList.SelectedItem().(core.RuneItem)
	if !ok {
		return errMsg{fmt.Errorf("invalid rune selection for update")}
	}
	selectedRune := selectedItem.Rune
	originalName := selectedRune.Name

	newName := m.inputs[0].Value()
	newDesc := m.inputs[1].Value()

	var newCmds []string
	for i := 2; i < len(m.inputs); i++ {
		if val := m.inputs[i].Value(); val != "" {
			newCmds = append(newCmds, val)
		}
	}
	newCmdsStr := strings.Join(newCmds, ";")

	// Build the command dynamically
	var parts []string
	parts = append(parts, "update-rune", fmt.Sprintf("%q", m.pwd), fmt.Sprintf("%q", originalName))

	if newName != "" && newName != originalName {
		parts = append(parts, "-name", fmt.Sprintf("%q", newName))
	}
	if newDesc != "" && newDesc != selectedRune.Description {
		parts = append(parts, "-desc", fmt.Sprintf("%q", newDesc))
	}
	if newCmdsStr != "" && newCmdsStr != strings.Join(selectedRune.Commands, ";") {
		parts = append(parts, "-cmds", fmt.Sprintf("%q", newCmdsStr))
	}

	// If no changes were made, don't run the command
	if len(parts) <= 3 {
		return noChangesMsg{}
	}

	cmd := strings.Join(parts, " ")
	_, err := m.sshClient.Command(cmd)
	if err != nil {
		return errMsg{err}
	}
	return tea.Sequence(
		func() tea.Msg {
			return core.ProgressUpdateMsg{Percent: 0.6, LogLine: "Refreshing spellbook..."}
		},
		m.getSpellbookContentCmd,
	)()
}

// deleteRuneCmd sends the command to delete a rune.
func (m *Model) deleteRuneCmd() tea.Msg {
	selectedItem, ok := m.runesList.SelectedItem().(core.RuneItem)
	if !ok {
		return errMsg{fmt.Errorf("invalid rune selection for delete")}
	}
	runeName := selectedItem.Rune.Name
	cmd := fmt.Sprintf("delete-rune %q %q", m.pwd, runeName)
	_, err := m.sshClient.Command(cmd)
	if err != nil {
		return errMsg{err}
	}
	return tea.Sequence(
		func() tea.Msg {
			return core.ProgressUpdateMsg{Percent: 0.6, LogLine: "Refreshing spellbook..."}
		},
		m.getSpellbookContentCmd,
	)()
}

func (m *Model) SetProgram(p *tea.Program) {
	// This function is no longer needed and will be removed.
}

// Init is called once when the application starts.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		m.getSpellbookContentCmd,
		m.StatusBar.Spinner.Tick, // Start the spinner
	)
}

func (m *Model) runDemoCmd() tea.Cmd {
	m.demoStep = 0
	return func() tea.Msg {
		return runDemoStepMsg{}
	}
}
