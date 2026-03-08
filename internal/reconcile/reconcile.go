// Package reconcile computes the set of actions required to bring two file
// trees into sync using three-way reconciliation against a common ancestor.
package reconcile

import (
	"github.com/Aronwwo/sneakernet-sync/internal/archive"
)

// Action describes a single sync operation.
type Action int

const (
	// ActionNone means no operation is needed.
	ActionNone Action = iota
	// ActionCopyToLocal copies the remote version to the local tree.
	ActionCopyToLocal
	// ActionCopyToRemote copies the local version to the remote.
	ActionCopyToRemote
	// ActionDeleteLocal deletes the file from the local tree.
	ActionDeleteLocal
	// ActionDeleteRemote deletes the file from the remote tree.
	ActionDeleteRemote
	// ActionConflict marks the file as conflicting.
	ActionConflict
	// ActionCreateDirLocal creates a directory locally.
	ActionCreateDirLocal
	// ActionCreateDirRemote creates a directory on remote.
	ActionCreateDirRemote
)

// String returns a human-readable action name.
func (a Action) String() string {
	switch a {
	case ActionNone:
		return "none"
	case ActionCopyToLocal:
		return "copy-to-local"
	case ActionCopyToRemote:
		return "copy-to-remote"
	case ActionDeleteLocal:
		return "delete-local"
	case ActionDeleteRemote:
		return "delete-remote"
	case ActionConflict:
		return "conflict"
	case ActionCreateDirLocal:
		return "create-dir-local"
	case ActionCreateDirRemote:
		return "create-dir-remote"
	default:
		return "unknown"
	}
}

// Entry holds the reconciliation result for a single path.
type Entry struct {
	Action     Action
	RelPath    string
	LocalHash  string
	RemoteHash string
	BaseHash   string
	IsDir      bool
	Reason     string
}

// Plan is the result of reconciliation.
type Plan struct {
	Entries []Entry
}

// Conflicts returns only conflict entries from the plan.
func (p *Plan) Conflicts() []Entry {
	var result []Entry
	for _, e := range p.Entries {
		if e.Action == ActionConflict {
			result = append(result, e)
		}
	}
	return result
}

// Actions returns only non-conflict, non-none entries.
func (p *Plan) Actions() []Entry {
	var result []Entry
	for _, e := range p.Entries {
		if e.Action != ActionNone && e.Action != ActionConflict {
			result = append(result, e)
		}
	}
	return result
}

// HasConflicts returns true if the plan contains any conflicts.
func (p *Plan) HasConflicts() bool {
	for _, e := range p.Entries {
		if e.Action == ActionConflict {
			return true
		}
	}
	return false
}

// Reconcile computes the sync plan given local, remote, and base (common ancestor)
// snapshots. Implements the following rules:
//
//  1. Changed only on one side → propagate change.
//  2. Changed on both sides differently → conflict.
//  3. Deleted on one side, modified on other → conflict.
//  4. Deleted on both sides → no action needed.
//  5. Created on both sides with different content → conflict.
//  6. Created on both sides with same content → no action needed.
func Reconcile(local, remote, base *archive.Snapshot) *Plan {
	plan := &Plan{}

	// Collect all paths from all three snapshots.
	allPaths := make(map[string]bool)
	if local != nil {
		for p := range local.Files {
			allPaths[p] = true
		}
	}
	if remote != nil {
		for p := range remote.Files {
			allPaths[p] = true
		}
	}
	if base != nil {
		for p := range base.Files {
			allPaths[p] = true
		}
	}

	for path := range allPaths {
		localState, inLocal := getState(local, path)
		remoteState, inRemote := getState(remote, path)
		baseState, inBase := getState(base, path)

		entry := reconcilePath(path, localState, remoteState, baseState, inLocal, inRemote, inBase)
		if entry.Action != ActionNone {
			plan.Entries = append(plan.Entries, entry)
		}
	}

	return plan
}

func getState(snap *archive.Snapshot, path string) (archive.FileState, bool) {
	if snap == nil {
		return archive.FileState{}, false
	}
	s, ok := snap.Files[path]
	return s, ok
}

func reconcilePath(path string, local, remote, base archive.FileState, inLocal, inRemote, inBase bool) Entry {
	entry := Entry{
		RelPath: path,
		Action:  ActionNone,
	}

	if inLocal {
		entry.LocalHash = local.ContentHash
		entry.IsDir = local.IsDir
	}
	if inRemote {
		entry.RemoteHash = remote.ContentHash
		if remote.IsDir {
			entry.IsDir = true
		}
	}
	if inBase {
		entry.BaseHash = base.ContentHash
	}

	localChanged := hasChanged(local, base, inLocal, inBase)
	remoteChanged := hasChanged(remote, base, inRemote, inBase)

	// Case: neither side changed.
	if !localChanged && !remoteChanged {
		return entry
	}

	// Case: only local changed.
	if localChanged && !remoteChanged {
		return handleOneSideChange(entry, local, inLocal, inBase, true)
	}

	// Case: only remote changed.
	if !localChanged && remoteChanged {
		return handleOneSideChange(entry, remote, inRemote, inBase, false)
	}

	// Both sides changed — check for conflicts.
	return handleBothChanged(entry, local, remote, inLocal, inRemote, inBase)
}

func hasChanged(state, base archive.FileState, inState, inBase bool) bool {
	if !inState && !inBase {
		return false
	}
	if inState && !inBase {
		return true // new file
	}
	if !inState && inBase {
		return true // deleted
	}
	return state.ContentHash != base.ContentHash
}

func handleOneSideChange(entry Entry, changed archive.FileState, inChanged, inBase, isLocal bool) Entry {
	if !inChanged && inBase {
		// File was deleted on this side.
		if isLocal {
			entry.Action = ActionDeleteRemote
			entry.Reason = "deleted locally, propagate to remote"
		} else {
			entry.Action = ActionDeleteLocal
			entry.Reason = "deleted remotely, propagate to local"
		}
		return entry
	}

	if inChanged && !inBase {
		// New file created.
		if changed.IsDir {
			if isLocal {
				entry.Action = ActionCreateDirRemote
				entry.Reason = "new directory created locally"
			} else {
				entry.Action = ActionCreateDirLocal
				entry.Reason = "new directory created remotely"
			}
		} else {
			if isLocal {
				entry.Action = ActionCopyToRemote
				entry.Reason = "new file created locally"
			} else {
				entry.Action = ActionCopyToLocal
				entry.Reason = "new file created remotely"
			}
		}
		return entry
	}

	// File was modified on this side.
	if isLocal {
		entry.Action = ActionCopyToRemote
		entry.Reason = "modified locally"
	} else {
		entry.Action = ActionCopyToLocal
		entry.Reason = "modified remotely"
	}
	return entry
}

func handleBothChanged(entry Entry, local, remote archive.FileState, inLocal, inRemote, inBase bool) Entry {
	// Rule 4: Both deleted → no action.
	if !inLocal && !inRemote {
		entry.Action = ActionNone
		entry.Reason = "both sides deleted"
		return entry
	}

	// Rule 3: Deleted on one side, modified on other → conflict.
	if !inLocal && inRemote {
		entry.Action = ActionConflict
		entry.Reason = "deleted locally but modified remotely"
		return entry
	}
	if inLocal && !inRemote {
		entry.Action = ActionConflict
		entry.Reason = "modified locally but deleted remotely"
		return entry
	}

	// Both exist and both changed. If hashes match → convergent change, no action.
	if local.ContentHash == remote.ContentHash {
		entry.Action = ActionNone
		entry.Reason = "both sides changed identically"
		return entry
	}

	// Rule 2 & 5: Both changed differently → conflict.
	if inBase {
		entry.Action = ActionConflict
		entry.Reason = "both sides modified differently"
	} else {
		entry.Action = ActionConflict
		entry.Reason = "both sides created with different content"
	}
	return entry
}
