package app

import (
	"catalyst/internal/app/components/statusbar"
	"catalyst/internal/app/styles"
	"catalyst/internal/config"
	"catalyst/internal/local"
	"catalyst/internal/ssh"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/v2/textinput"
	tea "github.com/charmbracelet/bubbletea/v2"
)

// Define application states.
type state int

const (
	checkingSpellbook state = iota
	creatingSpellbook
	ready
	showingRunes
	creatingRune
	executingRune
	showingLoegs
	creatingLoeg
	editingRune
	errState
)

// Define messages for async operations.
type (
	gotSpellbookMsg struct{ spellbook Spellbook }
	runeCreatedMsg  struct{}
	runeExecutedMsg struct{ output string }
	gotLoegsMsg     struct{ loegs map[string]string }
	loegSetMsg      struct{}
	loegRemovedMsg  struct{}
	runeUpdatedMsg  struct{}
	runeDeletedMsg  struct{}
	errMsg          struct{ err error }
)

// Model is the main application model.
type Model struct {
	sshClient   *ssh.Client
	localRunner *local.Runner
	state       state
	pwd         string // Current working directory (spellbook path)
	menuItems   []string
	cursor      int
	spellbook   *Spellbook        // Our in-memory cache
	loegKeys    []string          // For ordered display and selection
	inputs      []textinput.Model // For the "Create Rune" form
	focusIndex  int
	output      string // To store output from executed runes
	err         error
	StatusBar   statusbar.StatusBar
	Theme       *styles.Theme
}

// NewModel creates a new application model.
func NewModel(cfg *config.Config) Model {
	pwd, _ := os.Getwd() // Get PWD once at the start

	theme := styles.NewCharmtoneTheme()
	statusbar := statusbar.New(
		"initializing Catalyst",
		statusbar.LevelInfo,
		50,
		theme,
	)

	m := Model{
		sshClient:   ssh.NewClient(cfg.RuneCraftHost),
		localRunner: local.NewRunner(),
		state:       checkingSpellbook,
		pwd:         pwd,
		menuItems:   []string{"Get Runes", "Create Rune", "Manage Loegs"},
		inputs:      make([]textinput.Model, 3), // name, desc, cmds
		focusIndex:  0,
		Theme:       theme,
		StatusBar:   statusbar,
	}

	// Initialize text inputs for the create rune form
	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		t.CharLimit = 128

		switch i {
		case 0:
			t.Placeholder = "Rune Name"
			t.Focus()
		case 1:
			t.Placeholder = "Description"
		case 2:
			t.Placeholder = "Commands (semicolon-separated)"
			t.CharLimit = 256
		}
		m.inputs[i] = t
	}

	return m
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
	return m.getSpellbookContentCmd() // Refresh cache
}

// executeRuneCmd runs the commands for a selected rune.
func (m *Model) executeRuneCmd() tea.Msg {
	if m.cursor < 0 || m.cursor >= len(m.spellbook.Runes) {
		return errMsg{fmt.Errorf("invalid rune selection")}
	}
	selectedRune := m.spellbook.Runes[m.cursor]

	output, err := m.localRunner.ExecuteCommands(selectedRune.Commands)
	if err != nil {
		return runeExecutedMsg{output: output}
	}
	return runeExecutedMsg{output: output}
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
	return m.getSpellbookContentCmd() // Refresh cache
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
	return m.getSpellbookContentCmd() // Refresh cache
}

// updateRuneCmd sends the command to update an existing rune.
func (m *Model) updateRuneCmd() tea.Msg {
	// ... (build command as before)
	if m.cursor < 0 || m.cursor >= len(m.spellbook.Runes) {
		return errMsg{fmt.Errorf("invalid rune selection for update")}
	}
	selectedRune := m.spellbook.Runes[m.cursor]
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
		return runeUpdatedMsg{} // No-op, just go back
	}

	cmd := strings.Join(parts, " ")
	_, err := m.sshClient.Command(cmd)
	if err != nil {
		return errMsg{err}
	}
	return m.getSpellbookContentCmd() // Refresh cache
}

// deleteRuneCmd sends the command to delete a rune.
func (m *Model) deleteRuneCmd() tea.Msg {
	if m.cursor < 0 || m.cursor >= len(m.spellbook.Runes) {
		return errMsg{fmt.Errorf("invalid rune selection for delete")}
	}
	runeName := m.spellbook.Runes[m.cursor].Name
	// ... (build command as before)
	cmd := fmt.Sprintf("delete-rune %q %q", m.pwd, runeName)
	_, err := m.sshClient.Command(cmd)
	if err != nil {
		return errMsg{err}
	}
	return m.getSpellbookContentCmd() // Refresh cache
}

// Init is called once when the application starts.
func (m Model) Init() tea.Cmd {
	return m.getSpellbookContentCmd
}
