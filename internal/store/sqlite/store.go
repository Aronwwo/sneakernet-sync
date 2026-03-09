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
	IsDir       bool
	Exists      bool
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
	Kind         string // "content", "delete_modify", "create_create"
}

// SnapshotRecord holds metadata for a snapshot.
type SnapshotRecord struct {
	SnapshotID string
	DeviceID   string
	CreatedAt  time.Time
	FileCount  int
}

// SnapshotEntry holds a single file entry within a snapshot.
type SnapshotEntry struct {
	SnapshotID  string
	RelPath     string
	ContentHash string
	Size        int64
	ModTime     time.Time
	IsDir       bool
	Exists      bool
}

// Device holds information about a registered device.
type Device struct {
	DeviceID  string
	Name      string
	CreatedAt time.Time
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

	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("set journal mode: %w", err)
	}
	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("enable foreign keys: %w", err)
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

// --- Device operations ---

// SaveDevice persists a Device record.
func (s *Store) SaveDevice(d Device) error {
	_, err := s.db.Exec(
		`INSERT OR REPLACE INTO devices (device_id, name, created_at) VALUES (?, ?, ?)`,
		d.DeviceID, d.Name, d.CreatedAt.UTC().Format(time.RFC3339),
	)
	return err
}

// GetDevice retrieves a device by ID.
func (s *Store) GetDevice(deviceID string) (*Device, error) {
	row := s.db.QueryRow(`SELECT device_id, name, created_at FROM devices WHERE device_id = ?`, deviceID)
	var d Device
	var createdAt string
	if err := row.Scan(&d.DeviceID, &d.Name, &createdAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	d.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	return &d, nil
}

// --- File record operations ---

// SaveFileRecord persists a FileRecord to the store.
func (s *Store) SaveFileRecord(r FileRecord) error {
	isDirInt := 0
	if r.IsDir {
		isDirInt = 1
	}
	existsInt := 1
	if !r.Exists {
		existsInt = 0
	}
	_, err := s.db.Exec(
		`INSERT OR REPLACE INTO files (rel_path, content_hash, size, mod_time, state, device_id, last_sync_at, is_dir, exists_flag)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		r.RelPath,
		r.ContentHash,
		r.Size,
		r.ModTime.UTC().Format(time.RFC3339Nano),
		r.State,
		r.DeviceID,
		r.LastSyncAt.UTC().Format(time.RFC3339),
		isDirInt,
		existsInt,
	)
	return err
}

// GetFileRecord retrieves a FileRecord by its relative path and device.
func (s *Store) GetFileRecord(relPath, deviceID string) (*FileRecord, error) {
	row := s.db.QueryRow(
		`SELECT rel_path, content_hash, size, mod_time, state, device_id, last_sync_at, is_dir, exists_flag
		 FROM files WHERE rel_path = ? AND device_id = ?`,
		relPath, deviceID,
	)
	return scanFileRecord(row)
}

// ListFiles returns all FileRecords for a given device.
func (s *Store) ListFiles(deviceID string) ([]FileRecord, error) {
	rows, err := s.db.Query(
		`SELECT rel_path, content_hash, size, mod_time, state, device_id, last_sync_at, is_dir, exists_flag
		 FROM files WHERE device_id = ? ORDER BY rel_path`,
		deviceID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []FileRecord
	for rows.Next() {
		r, err := scanFileRecordRows(rows)
		if err != nil {
			return nil, err
		}
		records = append(records, *r)
	}
	return records, rows.Err()
}

// DeleteFileRecord removes a file record.
func (s *Store) DeleteFileRecord(relPath, deviceID string) error {
	_, err := s.db.Exec(`DELETE FROM files WHERE rel_path = ? AND device_id = ?`, relPath, deviceID)
	return err
}

func scanFileRecord(row *sql.Row) (*FileRecord, error) {
	var r FileRecord
	var modTime, lastSync string
	var isDirInt, existsInt int
	if err := row.Scan(&r.RelPath, &r.ContentHash, &r.Size, &modTime, &r.State, &r.DeviceID, &lastSync, &isDirInt, &existsInt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	r.ModTime, _ = time.Parse(time.RFC3339Nano, modTime)
	r.LastSyncAt, _ = time.Parse(time.RFC3339, lastSync)
	r.IsDir = isDirInt == 1
	r.Exists = existsInt == 1
	return &r, nil
}

func scanFileRecordRows(rows *sql.Rows) (*FileRecord, error) {
	var r FileRecord
	var modTime, lastSync string
	var isDirInt, existsInt int
	if err := rows.Scan(&r.RelPath, &r.ContentHash, &r.Size, &modTime, &r.State, &r.DeviceID, &lastSync, &isDirInt, &existsInt); err != nil {
		return nil, err
	}
	r.ModTime, _ = time.Parse(time.RFC3339Nano, modTime)
	r.LastSyncAt, _ = time.Parse(time.RFC3339, lastSync)
	r.IsDir = isDirInt == 1
	r.Exists = existsInt == 1
	return &r, nil
}

// --- Conflict operations ---

// SaveConflict persists a Conflict record to the store.
func (s *Store) SaveConflict(c Conflict) error {
	kind := c.Kind
	if kind == "" {
		kind = "content"
	}
	_, err := s.db.Exec(
		`INSERT INTO conflicts (rel_path, local_hash, remote_hash, local_device, remote_device, detected_at, resolved, resolution, kind)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		c.RelPath, c.LocalHash, c.RemoteHash, c.LocalDevice, c.RemoteDevice,
		c.DetectedAt.UTC().Format(time.RFC3339), boolToInt(c.Resolved), c.Resolution, kind,
	)
	return err
}

// GetUnresolvedConflicts returns all unresolved Conflict records.
func (s *Store) GetUnresolvedConflicts() ([]Conflict, error) {
	rows, err := s.db.Query(
		`SELECT id, rel_path, local_hash, remote_hash, local_device, remote_device, detected_at, resolved, resolution, kind
		 FROM conflicts WHERE resolved = 0 ORDER BY rel_path`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conflicts []Conflict
	for rows.Next() {
		c, err := scanConflictRow(rows)
		if err != nil {
			return nil, err
		}
		conflicts = append(conflicts, *c)
	}
	return conflicts, rows.Err()
}

// ResolveConflict marks a conflict as resolved with the given resolution strategy.
func (s *Store) ResolveConflict(id int64, resolution string) error {
	_, err := s.db.Exec(
		`UPDATE conflicts SET resolved = 1, resolution = ? WHERE id = ?`,
		resolution, id,
	)
	return err
}

func scanConflictRow(rows *sql.Rows) (*Conflict, error) {
	var c Conflict
	var detectedAt string
	var resolvedInt int
	if err := rows.Scan(&c.ID, &c.RelPath, &c.LocalHash, &c.RemoteHash, &c.LocalDevice, &c.RemoteDevice, &detectedAt, &resolvedInt, &c.Resolution, &c.Kind); err != nil {
		return nil, err
	}
	c.DetectedAt, _ = time.Parse(time.RFC3339, detectedAt)
	c.Resolved = resolvedInt == 1
	return &c, nil
}

// --- Snapshot operations ---

// SaveSnapshot persists a snapshot and its entries.
func (s *Store) SaveSnapshot(snap SnapshotRecord, entries []SnapshotEntry) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	_, err = tx.Exec(
		`INSERT INTO snapshots (snapshot_id, device_id, created_at, file_count) VALUES (?, ?, ?, ?)`,
		snap.SnapshotID, snap.DeviceID, snap.CreatedAt.UTC().Format(time.RFC3339), snap.FileCount,
	)
	if err != nil {
		return fmt.Errorf("insert snapshot: %w", err)
	}

	stmt, err := tx.Prepare(
		`INSERT INTO snapshot_entries (snapshot_id, rel_path, content_hash, size, mod_time, is_dir, exists_flag)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
	)
	if err != nil {
		return fmt.Errorf("prepare entry insert: %w", err)
	}
	defer stmt.Close()

	for _, e := range entries {
		isDirInt := 0
		if e.IsDir {
			isDirInt = 1
		}
		existsInt := 1
		if !e.Exists {
			existsInt = 0
		}
		_, err = stmt.Exec(e.SnapshotID, e.RelPath, e.ContentHash, e.Size,
			e.ModTime.UTC().Format(time.RFC3339Nano), isDirInt, existsInt)
		if err != nil {
			return fmt.Errorf("insert entry %q: %w", e.RelPath, err)
		}
	}

	return tx.Commit()
}

// GetLatestSnapshot returns the most recent snapshot for a device.
func (s *Store) GetLatestSnapshot(deviceID string) (*SnapshotRecord, error) {
	row := s.db.QueryRow(
		`SELECT snapshot_id, device_id, created_at, file_count
		 FROM snapshots WHERE device_id = ? ORDER BY created_at DESC LIMIT 1`,
		deviceID,
	)
	var snap SnapshotRecord
	var createdAt string
	if err := row.Scan(&snap.SnapshotID, &snap.DeviceID, &createdAt, &snap.FileCount); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	snap.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	return &snap, nil
}

// GetSnapshotEntries returns all entries for a given snapshot.
func (s *Store) GetSnapshotEntries(snapshotID string) ([]SnapshotEntry, error) {
	rows, err := s.db.Query(
		`SELECT snapshot_id, rel_path, content_hash, size, mod_time, is_dir, exists_flag
		 FROM snapshot_entries WHERE snapshot_id = ? ORDER BY rel_path`,
		snapshotID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []SnapshotEntry
	for rows.Next() {
		var e SnapshotEntry
		var modTime string
		var isDirInt, existsInt int
		if err := rows.Scan(&e.SnapshotID, &e.RelPath, &e.ContentHash, &e.Size, &modTime, &isDirInt, &existsInt); err != nil {
			return nil, err
		}
		e.ModTime, _ = time.Parse(time.RFC3339Nano, modTime)
		e.IsDir = isDirInt == 1
		e.Exists = existsInt == 1
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// --- Config operations ---

// SetConfig sets a configuration key-value pair.
func (s *Store) SetConfig(key, value string) error {
	_, err := s.db.Exec(`INSERT OR REPLACE INTO config (key, value) VALUES (?, ?)`, key, value)
	return err
}

// GetConfig retrieves a config value by key. Returns empty string if not found.
func (s *Store) GetConfig(key string) (string, error) {
	row := s.db.QueryRow(`SELECT value FROM config WHERE key = ?`, key)
	var value string
	if err := row.Scan(&value); err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	return value, nil
}

// --- Tombstone operations ---

// SaveTombstone records a file deletion.
func (s *Store) SaveTombstone(relPath, deletedBy string, deletedAt time.Time) error {
	_, err := s.db.Exec(
		`INSERT OR REPLACE INTO tombstones (rel_path, deleted_by, deleted_at) VALUES (?, ?, ?)`,
		relPath, deletedBy, deletedAt.UTC().Format(time.RFC3339),
	)
	return err
}

// GetTombstones returns all tombstone records for a device.
func (s *Store) GetTombstones(deviceID string) (map[string]time.Time, error) {
	rows, err := s.db.Query(
		`SELECT rel_path, deleted_at FROM tombstones WHERE deleted_by = ?`,
		deviceID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]time.Time)
	for rows.Next() {
		var relPath, deletedAt string
		if err := rows.Scan(&relPath, &deletedAt); err != nil {
			return nil, err
		}
		t, _ := time.Parse(time.RFC3339, deletedAt)
		result[relPath] = t
	}
	return result, rows.Err()
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
