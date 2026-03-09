// Package core ties together the internal packages to implement the sync
// engine.
package core

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Aronwwo/sneakernet-sync/internal/apply"
	"github.com/Aronwwo/sneakernet-sync/internal/archive"
	"github.com/Aronwwo/sneakernet-sync/internal/blobstore"
	"github.com/Aronwwo/sneakernet-sync/internal/conflict"
	"github.com/Aronwwo/sneakernet-sync/internal/reconcile"
	"github.com/Aronwwo/sneakernet-sync/internal/scan"
	"github.com/Aronwwo/sneakernet-sync/internal/store/sqlite"
	"github.com/Aronwwo/sneakernet-sync/internal/transport"
)

const (
	// MetaDir is the hidden directory for sync metadata.
	MetaDir = ".sneakernet"
	// DBFile is the database filename.
	DBFile = "meta.db"
	// BlobDir is the blob storage directory.
	BlobDir = "blobs"
)

// Engine orchestrates push/pull/sync operations.
type Engine struct {
	RootDir    string
	DeviceID   string
	DeviceName string
	Store      *sqlite.Store
	Archive    *archive.Archive
	Blobs      *blobstore.BlobStore
}

// InitResult holds the result of initialization.
type InitResult struct {
	DeviceID string
	RootDir  string
	MetaDir  string
}

// ScanResult holds the result of a scan.
type ScanResult struct {
	Files       int
	Directories int
	SnapshotID  string
}

// StatusResult holds the sync status.
type StatusResult struct {
	DeviceID       string
	DeviceName     string
	RootDir        string
	LastSnapshot   string
	TrackedFiles   int
	HasConflicts   bool
	ConflictCount  int
}

// PushResult holds the result of a push operation.
type PushResult struct {
	Manifest   *transport.Manifest
	BlobCount  int
	SnapshotID string
}

// PullResult holds the result of a pull operation.
type PullResult struct {
	RemoteDevice string
	FileCount    int
	BlobsFound   int
	Actions      int
	Conflicts    int
	Applied      []apply.Result
}

// Init initializes a sync repository in the given directory.
func Init(rootDir, deviceName string) (*InitResult, error) {
	absRoot, err := filepath.Abs(rootDir)
	if err != nil {
		return nil, fmt.Errorf("resolve root dir: %w", err)
	}

	metaDir := filepath.Join(absRoot, MetaDir)
	if _, err := os.Stat(metaDir); err == nil {
		return nil, fmt.Errorf("already initialized: %s", metaDir)
	}

	if err := os.MkdirAll(filepath.Join(metaDir, BlobDir), 0o755); err != nil {
		return nil, fmt.Errorf("create meta dir: %w", err)
	}

	dbPath := filepath.Join(metaDir, DBFile)
	store, err := sqlite.New(dbPath)
	if err != nil {
		return nil, fmt.Errorf("create store: %w", err)
	}
	defer store.Close()

	deviceID, err := generateDeviceID()
	if err != nil {
		return nil, fmt.Errorf("generate device id: %w", err)
	}

	if deviceName == "" {
		hostname, _ := os.Hostname()
		if hostname == "" {
			hostname = "unknown"
		}
		deviceName = hostname
	}

	if err := store.SaveDevice(sqlite.Device{
		DeviceID:  deviceID,
		Name:      deviceName,
		CreatedAt: time.Now().UTC(),
	}); err != nil {
		return nil, fmt.Errorf("save device: %w", err)
	}

	if err := store.SetConfig("device_id", deviceID); err != nil {
		return nil, fmt.Errorf("save device_id config: %w", err)
	}
	if err := store.SetConfig("device_name", deviceName); err != nil {
		return nil, fmt.Errorf("save device_name config: %w", err)
	}
	if err := store.SetConfig("root_dir", absRoot); err != nil {
		return nil, fmt.Errorf("save root_dir config: %w", err)
	}

	return &InitResult{
		DeviceID: deviceID,
		RootDir:  absRoot,
		MetaDir:  metaDir,
	}, nil
}

// Open opens an existing sync repository.
func Open(rootDir string) (*Engine, error) {
	absRoot, err := filepath.Abs(rootDir)
	if err != nil {
		return nil, fmt.Errorf("resolve root dir: %w", err)
	}

	metaDir := filepath.Join(absRoot, MetaDir)
	if _, err := os.Stat(metaDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("not initialized: run 'sneakernet-sync init' first")
	}

	dbPath := filepath.Join(metaDir, DBFile)
	store, err := sqlite.New(dbPath)
	if err != nil {
		return nil, fmt.Errorf("open store: %w", err)
	}

	deviceID, err := store.GetConfig("device_id")
	if err != nil {
		store.Close()
		return nil, fmt.Errorf("get device_id: %w", err)
	}

	deviceName, err := store.GetConfig("device_name")
	if err != nil {
		store.Close()
		return nil, fmt.Errorf("get device_name: %w", err)
	}

	blobPath := filepath.Join(metaDir, BlobDir)

	return &Engine{
		RootDir:    absRoot,
		DeviceID:   deviceID,
		DeviceName: deviceName,
		Store:      store,
		Archive:    archive.New(store),
		Blobs:      blobstore.New(blobPath),
	}, nil
}

// Close releases resources.
func (e *Engine) Close() error {
	return e.Store.Close()
}

// Scan scans the root directory and takes a snapshot.
func (e *Engine) Scan() (*ScanResult, error) {
	scanner := scan.New(e.RootDir)
	changes, err := scanner.Scan()
	if err != nil {
		return nil, fmt.Errorf("scan: %w", err)
	}

	// Store blobs for all files.
	for _, c := range changes {
		if c.IsDir || c.ContentHash == "DIR" {
			continue
		}
		absPath := filepath.Join(e.RootDir, filepath.FromSlash(c.RelPath))
		if _, err := e.Blobs.Store(absPath); err != nil {
			return nil, fmt.Errorf("store blob for %q: %w", c.RelPath, err)
		}
	}

	// Take snapshot.
	snap, err := e.Archive.TakeSnapshot(e.DeviceID, changes)
	if err != nil {
		return nil, fmt.Errorf("take snapshot: %w", err)
	}

	dirs := 0
	files := 0
	for _, c := range changes {
		if c.IsDir {
			dirs++
		} else {
			files++
		}
	}

	return &ScanResult{
		Files:       files,
		Directories: dirs,
		SnapshotID:  snap.ID,
	}, nil
}

// Status returns the current sync status.
func (e *Engine) Status() (*StatusResult, error) {
	result := &StatusResult{
		DeviceID:   e.DeviceID,
		DeviceName: e.DeviceName,
		RootDir:    e.RootDir,
	}

	snap, err := e.Archive.GetLastSnapshot(e.DeviceID)
	if err != nil {
		return nil, fmt.Errorf("get last snapshot: %w", err)
	}
	if snap != nil {
		result.LastSnapshot = snap.ID
		result.TrackedFiles = len(snap.Files)
	}

	conflicts, err := e.Store.GetUnresolvedConflicts()
	if err != nil {
		return nil, fmt.Errorf("get conflicts: %w", err)
	}
	result.ConflictCount = len(conflicts)
	result.HasConflicts = len(conflicts) > 0

	return result, nil
}

// Push exports the current snapshot and blobs to external media.
func (e *Engine) Push(mediaRoot string, dryRun bool) (*PushResult, error) {
	snap, err := e.Archive.GetLastSnapshot(e.DeviceID)
	if err != nil {
		return nil, fmt.Errorf("get last snapshot: %w", err)
	}
	if snap == nil {
		return nil, fmt.Errorf("no snapshot found; run 'scan' first")
	}

	manifest, err := transport.Export(mediaRoot, snap, e.DeviceName, e.Blobs, transport.ExportOptions{DryRun: dryRun})
	if err != nil {
		return nil, fmt.Errorf("export: %w", err)
	}

	blobCount := 0
	for _, s := range snap.Files {
		if !s.IsDir && s.ContentHash != "DIR" {
			blobCount++
		}
	}

	return &PushResult{
		Manifest:   manifest,
		BlobCount:  blobCount,
		SnapshotID: snap.ID,
	}, nil
}

// Pull imports changes from external media and applies them locally.
func (e *Engine) Pull(mediaRoot string, dryRun bool) (*PullResult, error) {
	importResult, err := transport.Import(mediaRoot, e.Blobs)
	if err != nil {
		return nil, fmt.Errorf("import: %w", err)
	}

	remoteSnap := importResult.Snapshot

	// Get local snapshot.
	localSnap, err := e.Archive.GetLastSnapshot(e.DeviceID)
	if err != nil {
		return nil, fmt.Errorf("get local snapshot: %w", err)
	}

	// Get common ancestor (the last snapshot we had of the remote device).
	baseSnap, err := e.Archive.GetLastSnapshot(importResult.Manifest.DeviceID)
	if err != nil {
		return nil, fmt.Errorf("get base snapshot: %w", err)
	}

	// Reconcile.
	plan := reconcile.Reconcile(localSnap, remoteSnap, baseSnap)

	// Detect and save conflicts.
	conflicts := conflict.FromPlan(plan, e.DeviceID, importResult.Manifest.DeviceID)
	if len(conflicts) > 0 {
		if err := conflict.SaveConflicts(e.Store, conflicts); err != nil {
			return nil, fmt.Errorf("save conflicts: %w", err)
		}
	}

	// Apply actions.
	results := apply.Apply(plan, localSnap, e.Blobs, apply.Options{
		DryRun:  dryRun,
		RootDir: e.RootDir,
	})

	// Save the remote snapshot locally for next sync's base.
	if !dryRun {
		scanner := scan.New(e.RootDir)
		changes, err := scanner.Scan()
		if err != nil {
			return nil, fmt.Errorf("re-scan after apply: %w", err)
		}
		if _, err := e.Archive.TakeSnapshot(e.DeviceID, changes); err != nil {
			return nil, fmt.Errorf("post-apply snapshot: %w", err)
		}
	}

	actionCount := 0
	for _, r := range results {
		if r.OK {
			actionCount++
		}
	}

	return &PullResult{
		RemoteDevice: importResult.Manifest.DeviceID,
		FileCount:    importResult.FileCount,
		BlobsFound:   importResult.BlobsFound,
		Actions:      actionCount,
		Conflicts:    len(conflicts),
		Applied:      results,
	}, nil
}

// GetConflicts returns all unresolved conflicts.
func (e *Engine) GetConflicts() ([]sqlite.Conflict, error) {
	return e.Store.GetUnresolvedConflicts()
}

// ResolveConflict resolves a conflict by ID.
func (e *Engine) ResolveConflict(id int64, resolution string) error {
	return e.Store.ResolveConflict(id, resolution)
}

// Doctor runs integrity checks on the repository.
func (e *Engine) Doctor() ([]string, error) {
	var issues []string

	// Check meta dir exists.
	metaDir := filepath.Join(e.RootDir, MetaDir)
	if _, err := os.Stat(metaDir); os.IsNotExist(err) {
		issues = append(issues, "meta directory missing")
		return issues, nil
	}

	// Check DB is accessible.
	if e.Store == nil {
		issues = append(issues, "database not accessible")
		return issues, nil
	}

	// Check device is registered.
	dev, err := e.Store.GetDevice(e.DeviceID)
	if err != nil {
		issues = append(issues, fmt.Sprintf("error reading device: %v", err))
	} else if dev == nil {
		issues = append(issues, "device not found in database")
	}

	// Check for last snapshot.
	snap, err := e.Archive.GetLastSnapshot(e.DeviceID)
	if err != nil {
		issues = append(issues, fmt.Sprintf("error reading snapshot: %v", err))
	} else if snap == nil {
		issues = append(issues, "no snapshots found; run 'scan' first")
	} else {
		// Verify blob integrity.
		missingBlobs := 0
		for _, state := range snap.Files {
			if state.IsDir || state.ContentHash == "DIR" {
				continue
			}
			if !e.Blobs.Has(state.ContentHash) {
				missingBlobs++
			}
		}
		if missingBlobs > 0 {
			issues = append(issues, fmt.Sprintf("%d blob(s) missing from store", missingBlobs))
		}
	}

	if len(issues) == 0 {
		issues = append(issues, "all checks passed")
	}

	return issues, nil
}

func generateDeviceID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
