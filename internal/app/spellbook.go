package app

import "catalyst/internal/types"

// Spellbook represents the entire content of a spellbook, acting as our in-memory cache.
type Spellbook struct {
	Name  string              `json:"name"`
	Runes []types.Rune        `json:"runes"`
	Loegs map[string]string   `json:"loegs"`
	// Warnings can be added here in the future if the API supports it.
}
