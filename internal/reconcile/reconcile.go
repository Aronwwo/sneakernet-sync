// Package reconcile computes the set of actions required to bring two file
// trees into sync.
package reconcile

import (
	"fmt"

	"github.com/Aronwwo/sneakernet-sync/internal/archive"
)

// Action describes a single sync operation.
type Action int

const (
	// ActionNone means no operation is needed.
	ActionNone Action = iota
	// ActionCopyLocal copies the local version to the remote.
	ActionCopyLocal
	// ActionCopyRemote copies the remote version to the local tree.
	ActionCopyRemote
	// ActionDelete deletes the file from both trees.
	ActionDelete
	// ActionConflict marks the file as conflicting.
	ActionConflict
)

// Plan is the result of reconciliation: a mapping from rel_path to Action.
type Plan struct {
	Actions map[string]Action
}

// Reconcile computes the sync plan given local, remote, and archive snapshots.
func Reconcile(local, remote, arch archive.Snapshot) (*Plan, error) {
	return nil, fmt.Errorf("not implemented yet")
}
