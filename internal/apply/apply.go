// Package apply safely executes reconciliation actions on the local filesystem.
// Before writing any file, it re-validates the local state to avoid overwriting
// concurrent changes.
package apply

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Aronwwo/sneakernet-sync/internal/archive"
	"github.com/Aronwwo/sneakernet-sync/internal/blobstore"
	"github.com/Aronwwo/sneakernet-sync/internal/fsops"
	"github.com/Aronwwo/sneakernet-sync/internal/hash"
	"github.com/Aronwwo/sneakernet-sync/internal/reconcile"
)

// Result describes the outcome of applying one action.
type Result struct {
	RelPath string
	Action  reconcile.Action
	OK      bool
	Error   string
}

// Options controls the apply behavior.
type Options struct {
	DryRun  bool
	RootDir string
}

// Apply executes a reconciliation plan on the local filesystem.
// For each action that copies to local or deletes locally, it validates
// the current state before acting.
func Apply(plan *reconcile.Plan, localSnap *archive.Snapshot, blobs *blobstore.BlobStore, opts Options) []Result {
	var results []Result

	for _, entry := range plan.Entries {
		switch entry.Action {
		case reconcile.ActionCopyToLocal:
			r := applyCopyToLocal(entry, localSnap, blobs, opts)
			results = append(results, r)

		case reconcile.ActionDeleteLocal:
			r := applyDeleteLocal(entry, localSnap, opts)
			results = append(results, r)

		case reconcile.ActionCreateDirLocal:
			r := applyCreateDirLocal(entry, opts)
			results = append(results, r)

		case reconcile.ActionConflict:
			results = append(results, Result{
				RelPath: entry.RelPath,
				Action:  entry.Action,
				OK:      false,
				Error:   "conflict: " + entry.Reason,
			})

		default:
			// Actions like CopyToRemote, DeleteRemote, CreateDirRemote are
			// handled during export, not apply.
			results = append(results, Result{
				RelPath: entry.RelPath,
				Action:  entry.Action,
				OK:      true,
			})
		}
	}

	return results
}

func applyCopyToLocal(entry reconcile.Entry, localSnap *archive.Snapshot, blobs *blobstore.BlobStore, opts Options) Result {
	r := Result{RelPath: entry.RelPath, Action: entry.Action}
	destPath := filepath.Join(opts.RootDir, filepath.FromSlash(entry.RelPath))

	// Pre-write validation: check that local state hasn't changed since scan.
	if err := validateLocalState(destPath, entry.RelPath, localSnap); err != nil {
		r.Error = fmt.Sprintf("pre-write validation failed: %v", err)
		return r
	}

	if opts.DryRun {
		r.OK = true
		return r
	}

	// Retrieve blob from store.
	if err := blobs.Retrieve(entry.RemoteHash, destPath); err != nil {
		r.Error = fmt.Sprintf("retrieve blob: %v", err)
		return r
	}

	r.OK = true
	return r
}

func applyDeleteLocal(entry reconcile.Entry, localSnap *archive.Snapshot, opts Options) Result {
	r := Result{RelPath: entry.RelPath, Action: entry.Action}
	targetPath := filepath.Join(opts.RootDir, filepath.FromSlash(entry.RelPath))

	// Pre-write validation.
	if err := validateLocalState(targetPath, entry.RelPath, localSnap); err != nil {
		r.Error = fmt.Sprintf("pre-write validation failed: %v", err)
		return r
	}

	if opts.DryRun {
		r.OK = true
		return r
	}

	if err := os.Remove(targetPath); err != nil && !os.IsNotExist(err) {
		r.Error = fmt.Sprintf("delete: %v", err)
		return r
	}

	r.OK = true
	return r
}

func applyCreateDirLocal(entry reconcile.Entry, opts Options) Result {
	r := Result{RelPath: entry.RelPath, Action: entry.Action}

	if opts.DryRun {
		r.OK = true
		return r
	}

	dirPath := filepath.Join(opts.RootDir, filepath.FromSlash(entry.RelPath))
	if err := fsops.EnsureDir(dirPath); err != nil {
		r.Error = fmt.Sprintf("create dir: %v", err)
		return r
	}

	r.OK = true
	return r
}

// validateLocalState checks that the local file hasn't changed since the
// snapshot was taken. This prevents overwriting concurrent user edits.
func validateLocalState(absPath, relPath string, localSnap *archive.Snapshot) error {
	if localSnap == nil {
		return nil
	}

	expected, exists := localSnap.Files[relPath]

	info, err := os.Stat(absPath)
	fileExists := err == nil

	if !exists && !fileExists {
		// Neither expected nor present — OK.
		return nil
	}

	if exists && !expected.Exists {
		// Expected to not exist but was recorded — check actual.
		if fileExists {
			return fmt.Errorf("file %q unexpectedly exists", relPath)
		}
		return nil
	}

	if !exists && fileExists {
		// File appeared since snapshot — someone created it.
		return fmt.Errorf("file %q appeared since last scan", relPath)
	}

	if exists && !fileExists {
		// File disappeared since snapshot.
		return fmt.Errorf("file %q disappeared since last scan", relPath)
	}

	// File exists in both; validate hash.
	if expected.IsDir || info.IsDir() {
		return nil
	}

	currentHash, err := hash.FileHash(absPath)
	if err != nil {
		return fmt.Errorf("hash file %q: %w", relPath, err)
	}

	if currentHash != expected.ContentHash {
		return fmt.Errorf("file %q was modified since last scan (expected %s, got %s)", relPath, expected.ContentHash[:8], currentHash[:8])
	}

	return nil
}
