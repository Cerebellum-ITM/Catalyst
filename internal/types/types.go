package types

// Rune represents a single, executable script or command collection.
type Rune struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Commands    []string `json:"commands"`
}
