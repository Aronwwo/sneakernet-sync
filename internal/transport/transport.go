// Package transport handles export/import of sync data to/from external media.
//
// The USB media format is:
//
//	<media_root>/.offsync/
//	├── manifest.json          # session metadata and schema version
//	├── blobs/                 # content-addressed file blobs (AA/BBCC...)
//	├── snapshots/             # snapshot entries per device
//	│   └── <device_id>.json
//	└── lock                   # lock file to prevent partial import
package transport

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/Aronwwo/sneakernet-sync/internal/archive"
	"github.com/Aronwwo/sneakernet-sync/internal/blobstore"
	"github.com/Aronwwo/sneakernet-sync/internal/fsops"
)

const (
	// MediaDir is the directory name on external media.
	MediaDir = ".offsync"
	// SchemaVersion is the current schema version of the media format.
	SchemaVersion = 1
	// ManifestFile is the manifest filename.
	ManifestFile = "manifest.json"
	// LockFile is the lock filename.
	LockFile = "lock"
)

// Manifest describes the sync session on external media.
type Manifest struct {
	SchemaVersion int       `json:"schema_version"`
	DeviceID      string    `json:"device_id"`
	DeviceName    string    `json:"device_name"`
	SnapshotID    string    `json:"snapshot_id"`
	CreatedAt     time.Time `json:"created_at"`
	FileCount     int       `json:"file_count"`
}

// SnapshotFile represents a single file entry in a snapshot export.
type SnapshotFile struct {
	RelPath     string    `json:"rel_path"`
	ContentHash string    `json:"content_hash"`
	Size        int64     `json:"size"`
	ModTime     time.Time `json:"mod_time"`
	IsDir       bool      `json:"is_dir"`
	Exists      bool      `json:"exists"`
}

// ExportOptions controls export behavior.
type ExportOptions struct {
	DryRun bool
}

// ImportResult holds the result of an import operation.
type ImportResult struct {
	Manifest   Manifest
	Snapshot   *archive.Snapshot
	FileCount  int
	BlobsFound int
}

// Export writes the current snapshot and blobs to external media.
func Export(mediaRoot string, snap *archive.Snapshot, deviceName string, localBlobs *blobstore.BlobStore, opts ExportOptions) (*Manifest, error) {
	offsyncDir := filepath.Join(mediaRoot, MediaDir)
	blobsDir := filepath.Join(offsyncDir, "blobs")
	snapshotsDir := filepath.Join(offsyncDir, "snapshots")

	if opts.DryRun {
		manifest := &Manifest{
			SchemaVersion: SchemaVersion,
			DeviceID:      snap.DeviceID,
			DeviceName:    deviceName,
			SnapshotID:    snap.ID,
			CreatedAt:     time.Now().UTC(),
			FileCount:     len(snap.Files),
		}
		return manifest, nil
	}

	// Create directory structure.
	for _, dir := range []string{offsyncDir, blobsDir, snapshotsDir} {
		if err := fsops.EnsureDir(dir); err != nil {
			return nil, fmt.Errorf("create media dir %q: %w", dir, err)
		}
	}

	// Write lock file.
	lockPath := filepath.Join(offsyncDir, LockFile)
	if err := fsops.AtomicWrite(lockPath, []byte(fmt.Sprintf("locked by %s at %s", snap.DeviceID, time.Now().UTC().Format(time.RFC3339)))); err != nil {
		return nil, fmt.Errorf("write lock: %w", err)
	}

	// Export blobs.
	mediaBlobStore := blobstore.New(blobsDir)
	for _, state := range snap.Files {
		if state.IsDir || state.ContentHash == "" || state.ContentHash == "DIR" {
			continue
		}
		if mediaBlobStore.Has(state.ContentHash) {
			continue
		}
		srcPath := localBlobs.Path(state.ContentHash)
		if _, err := os.Stat(srcPath); os.IsNotExist(err) {
			continue
		}
		if _, err := mediaBlobStore.Store(srcPath); err != nil {
			return nil, fmt.Errorf("export blob %q: %w", state.ContentHash, err)
		}
	}

	// Export snapshot.
	var files []SnapshotFile
	for relPath, state := range snap.Files {
		files = append(files, SnapshotFile{
			RelPath:     relPath,
			ContentHash: state.ContentHash,
			Size:        state.Size,
			ModTime:     state.ModTime,
			IsDir:       state.IsDir,
			Exists:      state.Exists,
		})
	}

	snapData, err := json.MarshalIndent(files, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal snapshot: %w", err)
	}

	snapPath := filepath.Join(snapshotsDir, snap.DeviceID+".json")
	if err := fsops.AtomicWrite(snapPath, snapData); err != nil {
		return nil, fmt.Errorf("write snapshot: %w", err)
	}

	// Write manifest.
	manifest := &Manifest{
		SchemaVersion: SchemaVersion,
		DeviceID:      snap.DeviceID,
		DeviceName:    deviceName,
		SnapshotID:    snap.ID,
		CreatedAt:     time.Now().UTC(),
		FileCount:     len(snap.Files),
	}

	manifestData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal manifest: %w", err)
	}

	manifestPath := filepath.Join(offsyncDir, ManifestFile)
	if err := fsops.AtomicWrite(manifestPath, manifestData); err != nil {
		return nil, fmt.Errorf("write manifest: %w", err)
	}

	// Remove lock file.
	os.Remove(lockPath)

	return manifest, nil
}

// Import reads sync data from external media.
func Import(mediaRoot string, localBlobs *blobstore.BlobStore) (*ImportResult, error) {
	offsyncDir := filepath.Join(mediaRoot, MediaDir)

	// Check for lock file (indicates incomplete export).
	lockPath := filepath.Join(offsyncDir, LockFile)
	if _, err := os.Stat(lockPath); err == nil {
		return nil, fmt.Errorf("media has a lock file — export may be incomplete; remove %s to force", lockPath)
	}

	// Read manifest.
	manifestPath := filepath.Join(offsyncDir, ManifestFile)
	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("read manifest: %w", err)
	}

	var manifest Manifest
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}

	if manifest.SchemaVersion > SchemaVersion {
		return nil, fmt.Errorf("unsupported schema version %d (max supported: %d)", manifest.SchemaVersion, SchemaVersion)
	}

	// Read snapshot.
	snapPath := filepath.Join(offsyncDir, "snapshots", manifest.DeviceID+".json")
	snapData, err := os.ReadFile(snapPath)
	if err != nil {
		return nil, fmt.Errorf("read snapshot: %w", err)
	}

	var files []SnapshotFile
	if err := json.Unmarshal(snapData, &files); err != nil {
		return nil, fmt.Errorf("parse snapshot: %w", err)
	}

	// Build snapshot.
	snap := &archive.Snapshot{
		ID:       manifest.SnapshotID,
		DeviceID: manifest.DeviceID,
		Files:    make(map[string]archive.FileState, len(files)),
	}
	for _, f := range files {
		snap.Files[f.RelPath] = archive.FileState{
			ContentHash: f.ContentHash,
			Size:        f.Size,
			ModTime:     f.ModTime,
			IsDir:       f.IsDir,
			Exists:      f.Exists,
		}
	}

	// Import blobs.
	blobsDir := filepath.Join(offsyncDir, "blobs")
	mediaBlobStore := blobstore.New(blobsDir)
	blobsFound := 0

	for _, state := range snap.Files {
		if state.IsDir || state.ContentHash == "" || state.ContentHash == "DIR" {
			continue
		}
		if localBlobs.Has(state.ContentHash) {
			blobsFound++
			continue
		}
		srcPath := mediaBlobStore.Path(state.ContentHash)
		if _, err := os.Stat(srcPath); os.IsNotExist(err) {
			continue
		}

		src, err := os.Open(srcPath)
		if err != nil {
			return nil, fmt.Errorf("open media blob %q: %w", state.ContentHash, err)
		}
		err = localBlobs.StoreReader(state.ContentHash, src)
		src.Close()
		if err != nil {
			return nil, fmt.Errorf("import blob %q: %w", state.ContentHash, err)
		}
		blobsFound++
	}

	return &ImportResult{
		Manifest:   manifest,
		Snapshot:   snap,
		FileCount:  len(files),
		BlobsFound: blobsFound,
	}, nil
}

// ListRemoteDevices returns device IDs found in snapshots on external media.
func ListRemoteDevices(mediaRoot string) ([]string, error) {
	snapshotsDir := filepath.Join(mediaRoot, MediaDir, "snapshots")
	entries, err := os.ReadDir(snapshotsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var devices []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		ext := filepath.Ext(name)
		if ext == ".json" {
			devices = append(devices, name[:len(name)-len(ext)])
		}
	}
	return devices, nil
}

// ReadRemoteSnapshot reads a specific device's snapshot from media.
func ReadRemoteSnapshot(mediaRoot, deviceID string) (*archive.Snapshot, error) {
	snapPath := filepath.Join(mediaRoot, MediaDir, "snapshots", deviceID+".json")
	data, err := os.ReadFile(snapPath)
	if err != nil {
		return nil, fmt.Errorf("read snapshot for %q: %w", deviceID, err)
	}

	var files []SnapshotFile
	if err := json.Unmarshal(data, &files); err != nil {
		return nil, fmt.Errorf("parse snapshot for %q: %w", deviceID, err)
	}

	snap := &archive.Snapshot{
		DeviceID: deviceID,
		Files:    make(map[string]archive.FileState, len(files)),
	}
	for _, f := range files {
		snap.Files[f.RelPath] = archive.FileState{
			ContentHash: f.ContentHash,
			Size:        f.Size,
			ModTime:     f.ModTime,
			IsDir:       f.IsDir,
			Exists:      f.Exists,
		}
	}

	return snap, nil
}

// CopyBlob copies a single blob from media to local blob store.
func CopyBlob(mediaRoot string, contentHash string, localBlobs *blobstore.BlobStore) error {
	if localBlobs.Has(contentHash) {
		return nil
	}
	mediaBlobStore := blobstore.New(filepath.Join(mediaRoot, MediaDir, "blobs"))
	srcPath := mediaBlobStore.Path(contentHash)
	src, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("open media blob: %w", err)
	}
	defer src.Close()

	data, err := io.ReadAll(src)
	if err != nil {
		return fmt.Errorf("read media blob: %w", err)
	}

	return localBlobs.StoreReader(contentHash, bytesReader(data))
}

type bytesReaderWrapper struct {
	data []byte
	pos  int
}

func (r *bytesReaderWrapper) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n = copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

func bytesReader(data []byte) io.Reader {
	return &bytesReaderWrapper{data: data}
}
