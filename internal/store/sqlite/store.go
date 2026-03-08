// Package sqlite provides a SQLite-backed metadata store.
package sqlite

import (
	"database/sql"
	_ "embed"
	"fmt"
	"time"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

//go:embed schema.sql
var schema string

// FileRecord holds metadata for a tracked file.
type FileRecord struct {
	RelPath     string
	ContentHash string
	Size        int64
	ModTime     time.Time
	State       int
	DeviceID    string
	LastSyncAt  time.Time
}

// Conflict holds information about a detected sync conflict.
type Conflict struct {
	ID           int64
	RelPath      string
	LocalHash    string
	RemoteHash   string
	LocalDevice  string
	RemoteDevice string
	DetectedAt   time.Time
	Resolved     bool
	Resolution   string
}

// Store is a SQLite-backed metadata store.
type Store struct {
	db *sql.DB
}

// New opens (or creates) the SQLite database at dbPath and runs schema
// migrations.
func New(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite3 %q: %w", dbPath, err)
	}

	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("run schema migration: %w", err)
	}

	return &Store{db: db}, nil
}

// Close releases the database connection.
func (s *Store) Close() error {
	return s.db.Close()
}

// SaveFileRecord persists a FileRecord to the store.
func (s *Store) SaveFileRecord(_ FileRecord) error {
	return fmt.Errorf("not implemented yet")
}

// GetFileRecord retrieves a FileRecord by its relative path.
func (s *Store) GetFileRecord(_ string) (*FileRecord, error) {
	return nil, fmt.Errorf("not implemented yet")
}

// ListFiles returns all FileRecords in the store.
func (s *Store) ListFiles() ([]FileRecord, error) {
	return nil, fmt.Errorf("not implemented yet")
}

// SaveConflict persists a Conflict record to the store.
func (s *Store) SaveConflict(_ Conflict) error {
	return fmt.Errorf("not implemented yet")
}

// GetUnresolvedConflicts returns all unresolved Conflict records.
func (s *Store) GetUnresolvedConflicts() ([]Conflict, error) {
	return nil, fmt.Errorf("not implemented yet")
}
