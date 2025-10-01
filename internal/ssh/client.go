package ssh

import (
	"bytes"
	"fmt"
	"os/exec"
)

// Client handles SSH connections to the RuneCraft server.
type Client struct {
	Host string
}

// NewClient creates a new SSH client.
func NewClient(host string) *Client {
	return &Client{Host: host}
}

// Command executes a command on the remote server.
// It returns stdout if successful, or an error containing stderr if not.
func (c *Client) Command(command string) (string, error) {
	cmd := exec.Command("ssh", c.Host, command)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		// If stderr has content, return it as the primary error.
		if stderr.Len() > 0 {
			return "", fmt.Errorf("ssh command failed: %s", stderr.String())
		}
		return "", fmt.Errorf("ssh command execution failed: %w", err)
	}

	// RuneCraft API rule: if stderr is not empty, it's an error.
	if stderr.Len() > 0 {
		return "", fmt.Errorf("operation failed: %s", stderr.String())
	}

	return stdout.String(), nil
}
