package app

// Spellbook represents the entire content of a spellbook, acting as our in-memory cache.
type Spellbook struct {
	Runes []Rune              `json:"runes"`
	Loegs map[string]string   `json:"loegs"`
	// Warnings can be added here in the future if the API supports it.
}
