package local

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"sync"

	"catalyst/internal/types"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/creack/pty"
)

var (
	// Terminal query sequences that cause timeouts.
	cursorPosQuery = []byte("\x1b[6n")
	bgColorQuery   = []byte("\x1b]11;?\x1b\\")
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
	cmd.Env = os.Environ()

	ptmx, err := pty.Start(cmd)
	if err != nil {
		msgChan <- types.RuneCommandFinished{Err: err}
		return
	}
	defer func() { _ = ptmx.Close() }()

	var wg sync.WaitGroup
	wg.Add(1)

	// Goroutine to stream stdout and handle terminal queries.
	go func() {
		defer wg.Done()
		buffer := make([]byte, 4096)
		for {
			n, err := ptmx.Read(buffer)
			if n <= 0 {
				if err != nil {
					// Handle read error, though it usually just means the pty closed.
				}
				return
			}

			data := buffer[:n]

			// --- This is the core fix ---
			// Check for and respond to terminal queries to prevent timeouts.
			if bytes.Contains(data, cursorPosQuery) {
				_, _ = ptmx.Write([]byte("\x1b[1;1R")) // Respond: "cursor is at 1;1"
				data = bytes.ReplaceAll(data, cursorPosQuery, []byte(""))
			}
			if bytes.Contains(data, bgColorQuery) {
				// Respond: "background is black"
				_, _ = ptmx.Write([]byte("\x1b]11;rgb:0000/0000/0000\x07"))
				data = bytes.ReplaceAll(data, bgColorQuery, []byte(""))
			}
			// --- End of fix ---

			// Send the cleaned output (without queries) to the UI.
			if len(data) > 0 {
				msgChan <- types.RuneCommandOutputMsg{Output: string(data)}
			}

			if err != nil {
				return
			}
		}
	}()

	// Wait for the command to finish.
	processErr := cmd.Wait()

	// Now that the process is done, close the PTY.
	_ = ptmx.Close()

	// Wait for the reader to finish flushing any remaining output.
	wg.Wait()

	// Send the final message.
	msgChan <- types.RuneCommandFinished{Err: processErr}
}
