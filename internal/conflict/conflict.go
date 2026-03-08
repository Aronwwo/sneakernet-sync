// Package conflict provides conflict detection and management for bidirectional sync.
package conflict

import (
	"time"

	"github.com/Aronwwo/sneakernet-sync/internal/reconcile"
	"github.com/Aronwwo/sneakernet-sync/internal/store/sqlite"
)

// ConflictInfo describes a detected conflict for a single file.
type ConflictInfo struct {
	RelPath      string
	LocalHash    string
	RemoteHash   string
	BaseHash     string
	LocalDevice  string
	RemoteDevice string
	DetectedAt   time.Time
	Kind         string // "content", "delete_modify", "create_create"
	Reason       string
}

// FromPlan extracts conflicts from a reconciliation plan and enriches them
// with device information.
func FromPlan(plan *reconcile.Plan, localDevice, remoteDevice string) []ConflictInfo {
	var conflicts []ConflictInfo
	now := time.Now().UTC()

	for _, e := range plan.Conflicts() {
		kind := classifyConflict(e)
		conflicts = append(conflicts, ConflictInfo{
			RelPath:      e.RelPath,
			LocalHash:    e.LocalHash,
			RemoteHash:   e.RemoteHash,
			BaseHash:     e.BaseHash,
			LocalDevice:  localDevice,
			RemoteDevice: remoteDevice,
			DetectedAt:   now,
			Kind:         kind,
			Reason:       e.Reason,
		})
	}

	return conflicts
}

// SaveConflicts persists conflict records to the store.
func SaveConflicts(store *sqlite.Store, conflicts []ConflictInfo) error {
	for _, c := range conflicts {
		err := store.SaveConflict(sqlite.Conflict{
			RelPath:      c.RelPath,
			LocalHash:    c.LocalHash,
			RemoteHash:   c.RemoteHash,
			LocalDevice:  c.LocalDevice,
			RemoteDevice: c.RemoteDevice,
			DetectedAt:   c.DetectedAt,
			Kind:         c.Kind,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func classifyConflict(e reconcile.Entry) string {
	if e.LocalHash == "" || e.RemoteHash == "" {
		return "delete_modify"
	}
	if e.BaseHash == "" {
		return "create_create"
	}
	return "content"
}
