package app

import (
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
type spellbookCheckMsg struct{ spellbookJSON string }
type spellbookCreateMsg struct{ spellbookJSON string }
type gotRunesMsg struct{ runes []Rune }
type runeCreatedMsg struct{}
type runeExecutedMsg struct{ output string }
type gotLoegsMsg struct{ loegs map[string]string }
type loegSetMsg struct{}
type loegRemovedMsg struct{}
type runeUpdatedMsg struct{}
type runeDeletedMsg struct{}
type errMsg struct{ err error }

// Model is the main application model.
type Model struct {
	sshClient   *ssh.Client
	localRunner *local.Runner
	state       state
	pwd         string // Current working directory (spellbook path)
	menuItems   []string
	cursor      int
	runes       []Rune
	loegs       map[string]string
	loegKeys    []string // For ordered display and selection
	inputs      []textinput.Model // For the "Create Rune" form
	focusIndex  int
	output      string // To store output from executed runes
	err         error
}

// NewModel creates a new application model.
func NewModel(cfg *config.Config) Model {
	pwd, _ := os.Getwd() // Get PWD once at the start

	m := Model{
		sshClient:   ssh.NewClient(cfg.RuneCraftHost),
		localRunner: local.NewRunner(),
		state:       checkingSpellbook,
		pwd:         pwd,
		menuItems:   []string{"Get Runes", "Create Rune", "Manage Loegs"},
		inputs:      make([]textinput.Model, 3), // name, desc, cmds
		focusIndex:  0,
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

// checkSpellbookCmd is a command that checks for the spellbook's existence.
func (m *Model) checkSpellbookCmd() tea.Msg {
	cmd := fmt.Sprintf("get-runes %q", m.pwd)
	jsonStr, err := m.sshClient.Command(cmd)
	if err != nil {
		return errMsg{err: nil} // Signal to create the spellbook
	}
	return spellbookCheckMsg{spellbookJSON: jsonStr}
}

// createSpellbookCmd is a command that creates a new spellbook.
func (m *Model) createSpellbookCmd() tea.Msg {
	cmd := fmt.Sprintf("create-spellbook %q", m.pwd)
	jsonStr, err := m.sshClient.Command(cmd)
	if err != nil {
		return errMsg{err}
	}
	return spellbookCreateMsg{spellbookJSON: jsonStr}
}

// getRunesCmd fetches the list of runes from the backend.
func (m *Model) getRunesCmd() tea.Msg {
	cmd := fmt.Sprintf("get-runes %q", m.pwd)
	jsonStr, err := m.sshClient.Command(cmd)
	if err != nil {
		return errMsg{err}
	}

	var runes []Rune
	if err := json.Unmarshal([]byte(jsonStr), &runes); err != nil {
		return errMsg{err}
	}
	return gotRunesMsg{runes: runes}
}

// createRuneCmd sends the command to create a new rune.
func (m *Model) createRuneCmd() tea.Msg {
	name := m.inputs[0].Value()
	desc := m.inputs[1].Value()
	cmds := m.inputs[2].Value()

	if name == "" || desc == "" || cmds == "" {
		return errMsg{fmt.Errorf("all fields are required")}
	}

	cmd := fmt.Sprintf("create-rune %q -name %q -desc %q -cmds %q", m.pwd, name, desc, cmds)
	_, err := m.sshClient.Command(cmd)
	if err != nil {
		return errMsg{err}
	}
	return runeCreatedMsg{}
}

// executeRuneCmd runs the commands for a selected rune.
func (m *Model) executeRuneCmd() tea.Msg {
	if m.cursor < 0 || m.cursor >= len(m.runes) {
		return errMsg{fmt.Errorf("invalid rune selection")}
	}
	selectedRune := m.runes[m.cursor]

	output, err := m.localRunner.ExecuteCommands(selectedRune.Commands)
	if err != nil {
		// The error is already included in the output by the runner.
		return runeExecutedMsg{output: output}
	}

	return runeExecutedMsg{output: output}
}

// getLoegsCmd fetches the list of loegs.
func (m *Model) getLoegsCmd() tea.Msg {
	cmd := fmt.Sprintf("loeg list %q", m.pwd)
	jsonStr, err := m.sshClient.Command(cmd)
	if err != nil {
		return errMsg{err}
	}

	var loegs map[string]string
	if err := json.Unmarshal([]byte(jsonStr), &loegs); err != nil {
		return errMsg{err}
	}
	return gotLoegsMsg{loegs: loegs}
}

// setLoegCmd creates or updates a loeg.
func (m *Model) setLoegCmd() tea.Msg {
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
	return loegSetMsg{}
}

// removeLoegCmd removes a loeg.
func (m *Model) removeLoegCmd() tea.Msg {
	if m.cursor < 0 || m.cursor >= len(m.loegKeys) {
		return errMsg{fmt.Errorf("invalid loeg selection")}
	}
	key := m.loegKeys[m.cursor]

	cmd := fmt.Sprintf("loeg rm %q %s", m.pwd, key)
	_, err := m.sshClient.Command(cmd)
	if err != nil {
		return errMsg{err}
	}
	return loegRemovedMsg{}
}

// updateRuneCmd sends the command to update an existing rune.
func (m *Model) updateRuneCmd() tea.Msg {
	if m.cursor < 0 || m.cursor >= len(m.runes) {
		return errMsg{fmt.Errorf("invalid rune selection for update")}
	}
	selectedRune := m.runes[m.cursor]
	originalName := selectedRune.Name

	newName := m.inputs[0].Value()
	newDesc := m.inputs[1].Value()
	newCmdsStr := m.inputs[2].Value()

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
	return runeUpdatedMsg{}
}

// deleteRuneCmd sends the command to delete a rune.
func (m *Model) deleteRuneCmd() tea.Msg {
	if m.cursor < 0 || m.cursor >= len(m.runes) {
		return errMsg{fmt.Errorf("invalid rune selection for delete")}
	}
	runeName := m.runes[m.cursor].Name

	cmd := fmt.Sprintf("delete-rune %q %q", m.pwd, runeName)
	_, err := m.sshClient.Command(cmd)
	if err != nil {
		return errMsg{err}
	}
	return runeDeletedMsg{}
}



// Init is called once when the application starts.
func (m Model) Init() tea.Cmd {
	return tea.Batch(m.checkSpellbookCmd, textinput.Blink)
}
