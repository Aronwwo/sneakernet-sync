package reconcile_test

import (
	"testing"

	"github.com/Aronwwo/sneakernet-sync/internal/archive"
	"github.com/Aronwwo/sneakernet-sync/internal/reconcile"
	"github.com/stretchr/testify/require"
)

func makeSnap(files map[string]string) *archive.Snapshot {
	snap := &archive.Snapshot{
		Files: make(map[string]archive.FileState),
	}
	for path, hash := range files {
		snap.Files[path] = archive.FileState{
			ContentHash: hash,
			Exists:      true,
		}
	}
	return snap
}

// Rule 1: File changed only on one side → propagate.
func TestReconcile_OneSideModified(t *testing.T) {
	base := makeSnap(map[string]string{"file.txt": "hash-v1"})
	local := makeSnap(map[string]string{"file.txt": "hash-v2"})
	remote := makeSnap(map[string]string{"file.txt": "hash-v1"})

	plan := reconcile.Reconcile(local, remote, base)
	require.False(t, plan.HasConflicts())

	actions := plan.Actions()
	require.Len(t, actions, 1)
	require.Equal(t, "file.txt", actions[0].RelPath)
	require.Equal(t, reconcile.ActionCopyToRemote, actions[0].Action)
}

// Rule 1 (reverse): File changed only on remote side.
func TestReconcile_OneSideModifiedRemote(t *testing.T) {
	base := makeSnap(map[string]string{"file.txt": "hash-v1"})
	local := makeSnap(map[string]string{"file.txt": "hash-v1"})
	remote := makeSnap(map[string]string{"file.txt": "hash-v2"})

	plan := reconcile.Reconcile(local, remote, base)
	require.False(t, plan.HasConflicts())

	actions := plan.Actions()
	require.Len(t, actions, 1)
	require.Equal(t, reconcile.ActionCopyToLocal, actions[0].Action)
}

// Rule 2: Both sides modified differently → conflict.
func TestReconcile_BothModifiedDifferently(t *testing.T) {
	base := makeSnap(map[string]string{"file.txt": "hash-v1"})
	local := makeSnap(map[string]string{"file.txt": "hash-v2"})
	remote := makeSnap(map[string]string{"file.txt": "hash-v3"})

	plan := reconcile.Reconcile(local, remote, base)
	require.True(t, plan.HasConflicts())

	conflicts := plan.Conflicts()
	require.Len(t, conflicts, 1)
	require.Equal(t, "file.txt", conflicts[0].RelPath)
	require.Equal(t, "both sides modified differently", conflicts[0].Reason)
}

// Rule 3: Deleted on one side, modified on other → conflict.
func TestReconcile_DeleteModifyConflict(t *testing.T) {
	base := makeSnap(map[string]string{"file.txt": "hash-v1"})
	local := makeSnap(map[string]string{}) // deleted locally
	remote := makeSnap(map[string]string{"file.txt": "hash-v2"})

	plan := reconcile.Reconcile(local, remote, base)
	require.True(t, plan.HasConflicts())

	conflicts := plan.Conflicts()
	require.Len(t, conflicts, 1)
	require.Equal(t, "deleted locally but modified remotely", conflicts[0].Reason)
}

// Rule 3 (reverse): Modified locally, deleted remotely.
func TestReconcile_ModifyDeleteConflict(t *testing.T) {
	base := makeSnap(map[string]string{"file.txt": "hash-v1"})
	local := makeSnap(map[string]string{"file.txt": "hash-v2"})
	remote := makeSnap(map[string]string{}) // deleted remotely

	plan := reconcile.Reconcile(local, remote, base)
	require.True(t, plan.HasConflicts())

	conflicts := plan.Conflicts()
	require.Len(t, conflicts, 1)
	require.Equal(t, "modified locally but deleted remotely", conflicts[0].Reason)
}

// Rule 4: Both sides deleted → no action.
func TestReconcile_BothDeleted(t *testing.T) {
	base := makeSnap(map[string]string{"file.txt": "hash-v1"})
	local := makeSnap(map[string]string{})
	remote := makeSnap(map[string]string{})

	plan := reconcile.Reconcile(local, remote, base)
	require.False(t, plan.HasConflicts())
	require.Empty(t, plan.Actions())
}

// Rule 5: Both created same path with different content → conflict.
func TestReconcile_BothCreatedDifferent(t *testing.T) {
	base := makeSnap(map[string]string{}) // empty base
	local := makeSnap(map[string]string{"new.txt": "local-hash"})
	remote := makeSnap(map[string]string{"new.txt": "remote-hash"})

	plan := reconcile.Reconcile(local, remote, base)
	require.True(t, plan.HasConflicts())

	conflicts := plan.Conflicts()
	require.Len(t, conflicts, 1)
	require.Equal(t, "both sides created with different content", conflicts[0].Reason)
}

// Rule 6: Both created same path with same content → no action.
func TestReconcile_BothCreatedSame(t *testing.T) {
	base := makeSnap(map[string]string{})
	local := makeSnap(map[string]string{"new.txt": "same-hash"})
	remote := makeSnap(map[string]string{"new.txt": "same-hash"})

	plan := reconcile.Reconcile(local, remote, base)
	require.False(t, plan.HasConflicts())
	require.Empty(t, plan.Actions())
}

// New file created locally → propagate.
func TestReconcile_NewFileLocal(t *testing.T) {
	base := makeSnap(map[string]string{})
	local := makeSnap(map[string]string{"new.txt": "new-hash"})
	remote := makeSnap(map[string]string{})

	plan := reconcile.Reconcile(local, remote, base)
	require.False(t, plan.HasConflicts())

	actions := plan.Actions()
	require.Len(t, actions, 1)
	require.Equal(t, reconcile.ActionCopyToRemote, actions[0].Action)
}

// New file created remotely → propagate.
func TestReconcile_NewFileRemote(t *testing.T) {
	base := makeSnap(map[string]string{})
	local := makeSnap(map[string]string{})
	remote := makeSnap(map[string]string{"new.txt": "new-hash"})

	plan := reconcile.Reconcile(local, remote, base)
	require.False(t, plan.HasConflicts())

	actions := plan.Actions()
	require.Len(t, actions, 1)
	require.Equal(t, reconcile.ActionCopyToLocal, actions[0].Action)
}

// File deleted on one side, unchanged on other → propagate deletion.
func TestReconcile_DeletePropagate(t *testing.T) {
	base := makeSnap(map[string]string{"file.txt": "hash-v1"})
	local := makeSnap(map[string]string{}) // deleted locally
	remote := makeSnap(map[string]string{"file.txt": "hash-v1"})

	plan := reconcile.Reconcile(local, remote, base)
	require.False(t, plan.HasConflicts())

	actions := plan.Actions()
	require.Len(t, actions, 1)
	require.Equal(t, reconcile.ActionDeleteRemote, actions[0].Action)
}

// No changes on either side → no action.
func TestReconcile_NoChanges(t *testing.T) {
	base := makeSnap(map[string]string{"file.txt": "hash-v1"})
	local := makeSnap(map[string]string{"file.txt": "hash-v1"})
	remote := makeSnap(map[string]string{"file.txt": "hash-v1"})

	plan := reconcile.Reconcile(local, remote, base)
	require.False(t, plan.HasConflicts())
	require.Empty(t, plan.Actions())
}

// Both sides modified identically → convergent change, no action.
func TestReconcile_BothModifiedSame(t *testing.T) {
	base := makeSnap(map[string]string{"file.txt": "hash-v1"})
	local := makeSnap(map[string]string{"file.txt": "hash-v2"})
	remote := makeSnap(map[string]string{"file.txt": "hash-v2"})

	plan := reconcile.Reconcile(local, remote, base)
	require.False(t, plan.HasConflicts())
	require.Empty(t, plan.Actions())
}

// Nil base (first sync) → propagate local files.
func TestReconcile_NilBase(t *testing.T) {
	local := makeSnap(map[string]string{"a.txt": "hash-a"})
	remote := makeSnap(map[string]string{})

	plan := reconcile.Reconcile(local, remote, nil)
	require.False(t, plan.HasConflicts())

	actions := plan.Actions()
	require.Len(t, actions, 1)
	require.Equal(t, reconcile.ActionCopyToRemote, actions[0].Action)
}

// Multiple files with mixed states.
func TestReconcile_MixedChanges(t *testing.T) {
	base := makeSnap(map[string]string{
		"unchanged.txt": "hash-u",
		"modified.txt":  "hash-old",
		"deleted.txt":   "hash-d",
		"conflict.txt":  "hash-c",
	})
	local := makeSnap(map[string]string{
		"unchanged.txt": "hash-u",
		"modified.txt":  "hash-new",
		// deleted.txt is gone
		"conflict.txt": "hash-c-local",
		"new-local.txt": "hash-nl",
	})
	remote := makeSnap(map[string]string{
		"unchanged.txt": "hash-u",
		"modified.txt":  "hash-old",
		"deleted.txt":   "hash-d",
		"conflict.txt":  "hash-c-remote",
	})

	plan := reconcile.Reconcile(local, remote, base)

	// Count actions by type.
	actionMap := make(map[string]reconcile.Action)
	for _, e := range plan.Entries {
		actionMap[e.RelPath] = e.Action
	}

	require.Equal(t, reconcile.ActionCopyToRemote, actionMap["modified.txt"])
	require.Equal(t, reconcile.ActionDeleteRemote, actionMap["deleted.txt"])
	require.Equal(t, reconcile.ActionConflict, actionMap["conflict.txt"])
	require.Equal(t, reconcile.ActionCopyToRemote, actionMap["new-local.txt"])
}
