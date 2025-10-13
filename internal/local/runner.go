package local

import (
	"bytes"
	"fmt"
	"os/exec"
)

// Runner executes local shell commands.
type Runner struct{}

// NewRunner creates a new command runner.
func NewRunner() *Runner {
	return &Runner{}
}

// ExecuteCommand runs a single command.
// It returns the combined output of stdout and stderr.
func (r *Runner) ExecuteCommand(command string) (string, error) {
	// Using "sh -c" to properly handle commands with arguments.
	cmd := exec.Command("sh", "-c", command)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out // Combine stdout and stderr

	err := cmd.Run()
	if err != nil {
		// Append error message to the output and return.
		return out.String() + fmt.Sprintf("\nCommand failed: %s", err.Error()), err
	}
	return out.String(), nil
}
