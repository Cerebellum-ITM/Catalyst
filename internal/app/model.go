package app

import (
	"catalyst/internal/config"
	"catalyst/internal/ssh"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea/v2"
)

// Define application states.
type state int

const (
	checkingSpellbook state = iota
	creatingSpellbook
	ready
	errState // Renamed to avoid conflict with error interface
)

// Define messages for async operations.
type spellbookCheckMsg struct{ spellbookJSON string }
type spellbookCreateMsg struct{ spellbookJSON string }
type errMsg struct{ err error }

// Model is the main application model.
type Model struct {
	sshClient *ssh.Client
	state     state
	err       error
}

// NewModel creates a new application model.
func NewModel(cfg *config.Config) Model {
	return Model{
		sshClient: ssh.NewClient(cfg.RuneCraftHost),
		state:     checkingSpellbook,
	}
}

// checkSpellbookCmd is a command that checks for the spellbook's existence.
func (m *Model) checkSpellbookCmd() tea.Msg {
	pwd, err := os.Getwd()
	if err != nil {
		return errMsg{err}
	}

	cmd := fmt.Sprintf("get-runes %q", pwd)
	json, err := m.sshClient.Command(cmd)
	if err != nil {
		// We assume an error here means the spellbook doesn't exist.
		// A more robust solution would check the specific error message.
		return errMsg{err: nil} // Signal to create the spellbook
	}
	return spellbookCheckMsg{spellbookJSON: json}
}

// createSpellbookCmd is a command that creates a new spellbook.
func (m *Model) createSpellbookCmd() tea.Msg {
	pwd, err := os.Getwd()
	if err != nil {
		return errMsg{err}
	}

	cmd := fmt.Sprintf("create-spellbook %q", pwd)
	json, err := m.sshClient.Command(cmd)
	if err != nil {
		return errMsg{err}
	}
	return spellbookCreateMsg{spellbookJSON: json}
}

// Init is called once when the application starts.
func (m Model) Init() tea.Cmd {
	return m.checkSpellbookCmd
}
