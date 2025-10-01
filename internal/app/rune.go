package app

// Rune represents a single command or script that can be executed.
type Rune struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Commands    []string `json:"commands"`
}
