// Package archive manages file snapshots used during reconciliation.
package archive

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/Aronwwo/sneakernet-sync/internal/scan"
	"github.com/Aronwwo/sneakernet-sync/internal/store/sqlite"
)

// Snapshot represents the state of the file tree at a point in time.
// Files maps rel_path -> content_hash. An empty hash means the file was deleted
// (tombstone). Directories are tracked with hash = "DIR".
type Snapshot struct {
	ID       string
	DeviceID string
	Files    map[string]FileState
}

// FileState represents a single file's state within a snapshot.
type FileState struct {
	ContentHash string
	Size        int64
	ModTime     time.Time
	IsDir       bool
	Exists      bool
}

// Archive manages historical snapshots of the file tree.
type Archive struct {
	store *sqlite.Store
}

// New creates a new Archive backed by the given store.
func New(store *sqlite.Store) *Archive {
	return &Archive{store: store}
}

// TakeSnapshot captures the current scanned state of the file tree and
// persists it.
func (a *Archive) TakeSnapshot(deviceID string, changes []scan.FileChange) (*Snapshot, error) {
	snapID, err := generateID()
	if err != nil {
		return nil, fmt.Errorf("generate snapshot id: %w", err)
	}

	now := time.Now().UTC()
	snap := &Snapshot{
		ID:       snapID,
		DeviceID: deviceID,
		Files:    make(map[string]FileState, len(changes)),
	}

	entries := make([]sqlite.SnapshotEntry, 0, len(changes))
	for _, c := range changes {
		fs := FileState{
			ContentHash: c.ContentHash,
			Size:        c.Size,
			ModTime:     c.ModTime,
			IsDir:       c.IsDir,
			Exists:      true,
		}
		snap.Files[c.RelPath] = fs

		entries = append(entries, sqlite.SnapshotEntry{
			SnapshotID:  snapID,
			RelPath:     c.RelPath,
			ContentHash: c.ContentHash,
			Size:        c.Size,
			ModTime:     c.ModTime,
			IsDir:       c.IsDir,
			Exists:      true,
		})
	}

	rec := sqlite.SnapshotRecord{
		SnapshotID: snapID,
		DeviceID:   deviceID,
		CreatedAt:  now,
		FileCount:  len(changes),
	}

	if err := a.store.SaveSnapshot(rec, entries); err != nil {
		return nil, fmt.Errorf("save snapshot: %w", err)
	}

	return snap, nil
}

// GetLastSnapshot returns the most recent archived snapshot for a device.
func (a *Archive) GetLastSnapshot(deviceID string) (*Snapshot, error) {
	rec, err := a.store.GetLatestSnapshot(deviceID)
	if err != nil {
		return nil, fmt.Errorf("get latest snapshot: %w", err)
	}
	if rec == nil {
		return nil, nil
	}

	entries, err := a.store.GetSnapshotEntries(rec.SnapshotID)
	if err != nil {
		return nil, fmt.Errorf("get snapshot entries: %w", err)
	}

	snap := &Snapshot{
		ID:       rec.SnapshotID,
		DeviceID: rec.DeviceID,
		Files:    make(map[string]FileState, len(entries)),
	}
	for _, e := range entries {
		snap.Files[e.RelPath] = FileState{
			ContentHash: e.ContentHash,
			Size:        e.Size,
			ModTime:     e.ModTime,
			IsDir:       e.IsDir,
			Exists:      e.Exists,
		}
	}

	return snap, nil
}

// SnapshotFromEntries builds a Snapshot from a list of snapshot entries.
// Used when importing from external media.
func SnapshotFromEntries(id, deviceID string, entries []sqlite.SnapshotEntry) *Snapshot {
	snap := &Snapshot{
		ID:       id,
		DeviceID: deviceID,
		Files:    make(map[string]FileState, len(entries)),
	}
	for _, e := range entries {
		snap.Files[e.RelPath] = FileState{
			ContentHash: e.ContentHash,
			Size:        e.Size,
			ModTime:     e.ModTime,
			IsDir:       e.IsDir,
			Exists:      e.Exists,
		}
	}
	return snap
}

func generateID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
