package types

// Rune represents a single, executable script or command collection.
type Rune struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Commands    []string `json:"commands"`
}

// RuneCommandOutputMsg is sent for each line of output from a command.
type RuneCommandOutputMsg struct {
	Output string
}

// RuneCommandFinished is sent when a command has finished executing.
type RuneCommandFinished struct {
	Err error
}

