package local

import (
	"bufio"
	"context"
	"os/exec"
	"sync"

	"catalyst/internal/types"

	tea "github.com/charmbracelet/bubbletea/v2"
)

// Runner executes local shell commands.
type Runner struct{}

// NewRunner creates a new command runner.
func NewRunner() *Runner {
	return &Runner{}
}

// ExecuteCommand runs a single command and streams its output.
// It sends RuneCommandOutputMsg for output and RuneCommandFinished when done.
func (r *Runner) ExecuteCommand(ctx context.Context, command string, msgChan chan<- tea.Msg) {
	cmd := exec.CommandContext(ctx, "zsh", "-c", command)

	stdoutPipe, _ := cmd.StdoutPipe()
	stderrPipe, _ := cmd.StderrPipe()

	var wg sync.WaitGroup
	wg.Add(2)

	// Goroutine to stream stdout
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdoutPipe)
		for scanner.Scan() {
			msgChan <- types.RuneCommandOutputMsg{Output: scanner.Text() + "\n"}
		}
	}()

	// Goroutine to stream stderr
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			msgChan <- types.RuneCommandOutputMsg{Output: scanner.Text() + "\n"}
		}
	}()

	err := cmd.Start()
	if err != nil {
		msgChan <- types.RuneCommandFinished{Err: err}
		return
	}

	// Wait for streaming to finish, then for the command to exit
	wg.Wait()
	err = cmd.Wait()
	msgChan <- types.RuneCommandFinished{Err: err}
}
