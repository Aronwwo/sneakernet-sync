// Package archive manages file snapshots used during reconciliation.
package archive

import (
	"fmt"

	"github.com/Aronwwo/sneakernet-sync/internal/store/sqlite"
)

// Snapshot represents the state of the file tree at a point in time.
type Snapshot struct {
	Files map[string]string // rel_path -> content_hash
}

// Archive manages historical snapshots of the file tree.
type Archive struct {
	store *sqlite.Store
}

// New creates a new Archive backed by the given store.
func New(store *sqlite.Store) *Archive {
	return &Archive{store: store}
}

// TakeSnapshot captures the current state of the file tree.
func (a *Archive) TakeSnapshot() (*Snapshot, error) {
	return nil, fmt.Errorf("not implemented yet")
}

// GetLastSnapshot returns the most recent archived snapshot.
func (a *Archive) GetLastSnapshot() (*Snapshot, error) {
	return nil, fmt.Errorf("not implemented yet")
}
