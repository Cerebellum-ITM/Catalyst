package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// HistoryEntry represents a single rune execution record.
type HistoryEntry struct {
	ID          int
	RuneID      string
	SpellbookID string
	ExecutedAt  time.Time
}

// Database holds the connection pool.
type Database struct {
	*sql.DB
}

// InitDB initializes the SQLite database and creates the necessary tables.
func InitDB() (*Database, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user config directory: %w", err)
	}
	dbPath := filepath.Join(configDir, "Catalyst", "catalyst.db")

	if err := os.MkdirAll(filepath.Dir(dbPath), 0750); err != nil {
		return nil, fmt.Errorf("failed to create db directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Create the history table if it doesn't exist.
	query := `
	CREATE TABLE IF NOT EXISTS history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		rune_id TEXT NOT NULL,
		spellbook_id TEXT NOT NULL,
		executed_at DATETIME NOT NULL
	);
	`
	if _, err := db.Exec(query); err != nil {
		return nil, fmt.Errorf("failed to create history table: %w", err)
	}

	return &Database{db}, nil
}

// AddHistoryEntry inserts a new record into the history table.
func (db *Database) AddHistoryEntry(runeID, spellbookID string) error {
	query := `INSERT INTO history (rune_id, spellbook_id, executed_at) VALUES (?, ?, ?)`
	_, err := db.Exec(query, runeID, spellbookID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to insert history entry: %w", err)
	}
	return nil
}

// GetHistory retrieves all execution records from the database.
func (db *Database) GetHistory() ([]HistoryEntry, error) {
	query := `SELECT id, rune_id, spellbook_id, executed_at FROM history ORDER BY executed_at DESC`
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query history: %w", err)
	}
	defer rows.Close()

	var entries []HistoryEntry
	for rows.Next() {
		var entry HistoryEntry
		if err := rows.Scan(&entry.ID, &entry.RuneID, &entry.SpellbookID, &entry.ExecutedAt); err != nil {
			return nil, fmt.Errorf("failed to scan history row: %w", err)
		}
		entries = append(entries, entry)
	}

	return entries, nil
}
