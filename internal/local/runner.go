package local

import (
	"bytes"
	"os/exec"
	"strings"
)

// Runner executes local shell commands.
type Runner struct{}

// NewRunner creates a new command runner.
func NewRunner() *Runner {
	return &Runner{}
}

// ExecuteCommands runs a series of commands sequentially.
// It returns the combined output of all commands.
func (r *Runner) ExecuteCommands(commands []string) (string, error) {
	var output strings.Builder
	for _, command := range commands {
		// Using "sh -c" to properly handle commands with arguments.
		cmd := exec.Command("sh", "-c", command)
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &out // Combine stdout and stderr

		err := cmd.Run()
		output.WriteString(out.String())
		if err != nil {
			// Append error message to the output and return.
			output.WriteString("\nCommand failed: " + err.Error())
			return output.String(), err
		}
	}
	return output.String(), nil
}
