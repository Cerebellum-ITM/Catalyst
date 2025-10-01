package app

import (
	"catalyst/internal/config"
	"catalyst/internal/ssh"
	"encoding/json"
	"fmt"
	"os"

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
	errState
)

// Define messages for async operations.
type spellbookCheckMsg struct{ spellbookJSON string }
type spellbookCreateMsg struct{ spellbookJSON string }
type gotRunesMsg struct{ runes []Rune }
type runeCreatedMsg struct{}
type errMsg struct{ err error }

// Model is the main application model.
type Model struct {
	sshClient  *ssh.Client
	state      state
	pwd        string // Current working directory (spellbook path)
	menuItems  []string
	cursor     int
	runes      []Rune
	inputs     []textinput.Model // For the "Create Rune" form
	focusIndex int
	err        error
}

// NewModel creates a new application model.
func NewModel(cfg *config.Config) Model {
	pwd, _ := os.Getwd() // Get PWD once at the start

	m := Model{
		sshClient:  ssh.NewClient(cfg.RuneCraftHost),
		state:      checkingSpellbook,
		pwd:        pwd,
		menuItems:  []string{"Get Runes", "Create Rune"},
		inputs:     make([]textinput.Model, 3), // name, desc, cmds
		focusIndex: 0,
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
	json, err := m.sshClient.Command(cmd)
	if err != nil {
		return errMsg{err: nil} // Signal to create the spellbook
	}
	return spellbookCheckMsg{spellbookJSON: json}
}

// createSpellbookCmd is a command that creates a new spellbook.
func (m *Model) createSpellbookCmd() tea.Msg {
	cmd := fmt.Sprintf("create-spellbook %q", m.pwd)
	json, err := m.sshClient.Command(cmd)
	if err != nil {
		return errMsg{err}
	}
	return spellbookCreateMsg{spellbookJSON: json}
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

// Init is called once when the application starts.
func (m Model) Init() tea.Cmd {
	return tea.Batch(m.checkSpellbookCmd, textinput.Blink)
}
