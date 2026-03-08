package sqlite_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/Aronwwo/sneakernet-sync/internal/store/sqlite"
	"github.com/stretchr/testify/require"
)

func TestNewStore_OpenClose(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")

	store, err := sqlite.New(dbPath)
	require.NoError(t, err)
	require.NotNil(t, store)

	require.NoError(t, store.Close())
}

func TestNewStore_SchemaCreated(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "schema.db")

	store, err := sqlite.New(dbPath)
	require.NoError(t, err)
	require.NoError(t, store.Close())

	// Re-open to verify schema persists across connections.
	store2, err := sqlite.New(dbPath)
	require.NoError(t, err)
	require.NoError(t, store2.Close())
}

func newTestStore(t *testing.T) *sqlite.Store {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := sqlite.New(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { store.Close() })
	return store
}

func TestDevice_SaveAndGet(t *testing.T) {
	store := newTestStore(t)

	dev := sqlite.Device{
		DeviceID:  "dev-001",
		Name:      "TestDevice",
		CreatedAt: time.Now().UTC().Truncate(time.Second),
	}
	require.NoError(t, store.SaveDevice(dev))

	got, err := store.GetDevice("dev-001")
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, "dev-001", got.DeviceID)
	require.Equal(t, "TestDevice", got.Name)
}

func TestDevice_GetNotFound(t *testing.T) {
	store := newTestStore(t)
	got, err := store.GetDevice("nonexistent")
	require.NoError(t, err)
	require.Nil(t, got)
}

func TestFileRecord_SaveAndGet(t *testing.T) {
	store := newTestStore(t)

	require.NoError(t, store.SaveDevice(sqlite.Device{
		DeviceID: "dev-001", Name: "Test", CreatedAt: time.Now(),
	}))

	now := time.Now().UTC().Truncate(time.Nanosecond)
	rec := sqlite.FileRecord{
		RelPath:     "docs/readme.txt",
		ContentHash: "abc123",
		Size:        42,
		ModTime:     now,
		State:       0,
		DeviceID:    "dev-001",
		LastSyncAt:  now,
		IsDir:       false,
		Exists:      true,
	}
	require.NoError(t, store.SaveFileRecord(rec))

	got, err := store.GetFileRecord("docs/readme.txt", "dev-001")
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, "docs/readme.txt", got.RelPath)
	require.Equal(t, "abc123", got.ContentHash)
	require.Equal(t, int64(42), got.Size)
	require.True(t, got.Exists)
	require.False(t, got.IsDir)
}

func TestFileRecord_GetNotFound(t *testing.T) {
	store := newTestStore(t)
	got, err := store.GetFileRecord("nonexistent", "dev-001")
	require.NoError(t, err)
	require.Nil(t, got)
}

func TestFileRecord_ListFiles(t *testing.T) {
	store := newTestStore(t)

	require.NoError(t, store.SaveDevice(sqlite.Device{
		DeviceID: "dev-001", Name: "Test", CreatedAt: time.Now(),
	}))

	now := time.Now().UTC()
	for _, name := range []string{"a.txt", "b.txt", "c.txt"} {
		require.NoError(t, store.SaveFileRecord(sqlite.FileRecord{
			RelPath: name, ContentHash: "hash-" + name, Size: 10,
			ModTime: now, DeviceID: "dev-001", LastSyncAt: now, Exists: true,
		}))
	}

	files, err := store.ListFiles("dev-001")
	require.NoError(t, err)
	require.Len(t, files, 3)
	require.Equal(t, "a.txt", files[0].RelPath)
	require.Equal(t, "b.txt", files[1].RelPath)
	require.Equal(t, "c.txt", files[2].RelPath)
}

func TestFileRecord_Delete(t *testing.T) {
	store := newTestStore(t)

	require.NoError(t, store.SaveDevice(sqlite.Device{
		DeviceID: "dev-001", Name: "Test", CreatedAt: time.Now(),
	}))

	now := time.Now().UTC()
	require.NoError(t, store.SaveFileRecord(sqlite.FileRecord{
		RelPath: "a.txt", ContentHash: "hash", Size: 10,
		ModTime: now, DeviceID: "dev-001", LastSyncAt: now, Exists: true,
	}))

	require.NoError(t, store.DeleteFileRecord("a.txt", "dev-001"))

	got, err := store.GetFileRecord("a.txt", "dev-001")
	require.NoError(t, err)
	require.Nil(t, got)
}

func TestConflict_SaveAndList(t *testing.T) {
	store := newTestStore(t)

	c := sqlite.Conflict{
		RelPath:      "file.txt",
		LocalHash:    "local-hash",
		RemoteHash:   "remote-hash",
		LocalDevice:  "dev-A",
		RemoteDevice: "dev-B",
		DetectedAt:   time.Now().UTC(),
		Kind:         "content",
	}
	require.NoError(t, store.SaveConflict(c))

	conflicts, err := store.GetUnresolvedConflicts()
	require.NoError(t, err)
	require.Len(t, conflicts, 1)
	require.Equal(t, "file.txt", conflicts[0].RelPath)
	require.Equal(t, "content", conflicts[0].Kind)
	require.False(t, conflicts[0].Resolved)
}

func TestConflict_Resolve(t *testing.T) {
	store := newTestStore(t)

	require.NoError(t, store.SaveConflict(sqlite.Conflict{
		RelPath: "file.txt", LocalHash: "a", RemoteHash: "b",
		LocalDevice: "A", RemoteDevice: "B",
		DetectedAt: time.Now().UTC(), Kind: "content",
	}))

	conflicts, err := store.GetUnresolvedConflicts()
	require.NoError(t, err)
	require.Len(t, conflicts, 1)

	require.NoError(t, store.ResolveConflict(conflicts[0].ID, "keep-local"))

	unresolved, err := store.GetUnresolvedConflicts()
	require.NoError(t, err)
	require.Empty(t, unresolved)
}

func TestSnapshot_SaveAndGet(t *testing.T) {
	store := newTestStore(t)

	require.NoError(t, store.SaveDevice(sqlite.Device{
		DeviceID: "dev-001", Name: "Test", CreatedAt: time.Now(),
	}))

	now := time.Now().UTC()
	snap := sqlite.SnapshotRecord{
		SnapshotID: "snap-001",
		DeviceID:   "dev-001",
		CreatedAt:  now,
		FileCount:  2,
	}
	entries := []sqlite.SnapshotEntry{
		{SnapshotID: "snap-001", RelPath: "a.txt", ContentHash: "hash-a", Size: 10, ModTime: now, Exists: true},
		{SnapshotID: "snap-001", RelPath: "b.txt", ContentHash: "hash-b", Size: 20, ModTime: now, Exists: true},
	}

	require.NoError(t, store.SaveSnapshot(snap, entries))

	got, err := store.GetLatestSnapshot("dev-001")
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, "snap-001", got.SnapshotID)
	require.Equal(t, 2, got.FileCount)

	gotEntries, err := store.GetSnapshotEntries("snap-001")
	require.NoError(t, err)
	require.Len(t, gotEntries, 2)
}

func TestConfig_SetAndGet(t *testing.T) {
	store := newTestStore(t)

	require.NoError(t, store.SetConfig("key1", "value1"))

	val, err := store.GetConfig("key1")
	require.NoError(t, err)
	require.Equal(t, "value1", val)

	// Override.
	require.NoError(t, store.SetConfig("key1", "value2"))
	val, err = store.GetConfig("key1")
	require.NoError(t, err)
	require.Equal(t, "value2", val)
}

func TestConfig_GetNotFound(t *testing.T) {
	store := newTestStore(t)

	val, err := store.GetConfig("missing")
	require.NoError(t, err)
	require.Equal(t, "", val)
}

func TestTombstone_SaveAndGet(t *testing.T) {
	store := newTestStore(t)

	now := time.Now().UTC()
	require.NoError(t, store.SaveTombstone("deleted.txt", "dev-001", now))

	tombstones, err := store.GetTombstones("dev-001")
	require.NoError(t, err)
	require.Len(t, tombstones, 1)
	_, ok := tombstones["deleted.txt"]
	require.True(t, ok)
}
