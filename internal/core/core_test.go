package core_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Aronwwo/sneakernet-sync/internal/core"
	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	tmp := t.TempDir()
	dir := filepath.Join(tmp, "myproject")
	require.NoError(t, os.MkdirAll(dir, 0o755))

	result, err := core.Init(dir, "TestDevice")
	require.NoError(t, err)
	require.NotEmpty(t, result.DeviceID)
	require.DirExists(t, result.MetaDir)
}

func TestInit_AlreadyInitialized(t *testing.T) {
	tmp := t.TempDir()

	_, err := core.Init(tmp, "Dev1")
	require.NoError(t, err)

	_, err = core.Init(tmp, "Dev2")
	require.Error(t, err)
	require.Contains(t, err.Error(), "already initialized")
}

func TestScanAndStatus(t *testing.T) {
	tmp := t.TempDir()
	dir := filepath.Join(tmp, "project")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "file.txt"), []byte("hello"), 0o600))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "sub"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "sub", "nested.txt"), []byte("world"), 0o600))

	_, err := core.Init(dir, "TestDev")
	require.NoError(t, err)

	engine, err := core.Open(dir)
	require.NoError(t, err)
	defer engine.Close()

	scanResult, err := engine.Scan()
	require.NoError(t, err)
	require.Equal(t, 2, scanResult.Files)
	require.Equal(t, 1, scanResult.Directories)
	require.NotEmpty(t, scanResult.SnapshotID)

	status, err := engine.Status()
	require.NoError(t, err)
	require.NotEmpty(t, status.DeviceID)
	require.Equal(t, 3, status.TrackedFiles) // 2 files + 1 dir
	require.False(t, status.HasConflicts)
}

func TestDoctor(t *testing.T) {
	tmp := t.TempDir()
	dir := filepath.Join(tmp, "project")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "file.txt"), []byte("content"), 0o600))

	_, err := core.Init(dir, "TestDev")
	require.NoError(t, err)

	engine, err := core.Open(dir)
	require.NoError(t, err)
	defer engine.Close()

	// Before scan, should report no snapshots.
	issues, err := engine.Doctor()
	require.NoError(t, err)
	require.Contains(t, issues, "no snapshots found; run 'scan' first")

	// After scan, should pass.
	_, err = engine.Scan()
	require.NoError(t, err)

	issues, err = engine.Doctor()
	require.NoError(t, err)
	require.Contains(t, issues, "all checks passed")
}

// Integration test: full two-device sync via USB media.
func TestTwoDeviceSync(t *testing.T) {
	tmp := t.TempDir()
	dirA := filepath.Join(tmp, "device-A")
	dirB := filepath.Join(tmp, "device-B")
	media := filepath.Join(tmp, "usb-media")
	require.NoError(t, os.MkdirAll(dirA, 0o755))
	require.NoError(t, os.MkdirAll(dirB, 0o755))
	require.NoError(t, os.MkdirAll(media, 0o755))

	// Device A: init and create files.
	require.NoError(t, os.WriteFile(filepath.Join(dirA, "shared.txt"), []byte("shared content"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dirA, "a-only.txt"), []byte("only on A"), 0o600))

	_, err := core.Init(dirA, "DeviceA")
	require.NoError(t, err)

	engineA, err := core.Open(dirA)
	require.NoError(t, err)
	defer engineA.Close()

	_, err = engineA.Scan()
	require.NoError(t, err)

	// Device A: push to media.
	pushResult, err := engineA.Push(media, false)
	require.NoError(t, err)
	require.NotNil(t, pushResult.Manifest)

	// Device B: init and create its own files.
	require.NoError(t, os.WriteFile(filepath.Join(dirB, "b-only.txt"), []byte("only on B"), 0o600))

	_, err = core.Init(dirB, "DeviceB")
	require.NoError(t, err)

	engineB, err := core.Open(dirB)
	require.NoError(t, err)
	defer engineB.Close()

	_, err = engineB.Scan()
	require.NoError(t, err)

	// Device B: pull from media.
	pullResult, err := engineB.Pull(media, false)
	require.NoError(t, err)
	require.Equal(t, 0, pullResult.Conflicts)
	require.Greater(t, pullResult.Actions, 0)

	// Verify files were copied to Device B.
	sharedContent, err := os.ReadFile(filepath.Join(dirB, "shared.txt"))
	require.NoError(t, err)
	require.Equal(t, "shared content", string(sharedContent))

	aOnlyContent, err := os.ReadFile(filepath.Join(dirB, "a-only.txt"))
	require.NoError(t, err)
	require.Equal(t, "only on A", string(aOnlyContent))
}

// Test dry-run doesn't write anything.
func TestPushDryRun(t *testing.T) {
	tmp := t.TempDir()
	dir := filepath.Join(tmp, "project")
	media := filepath.Join(tmp, "usb")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.MkdirAll(media, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "file.txt"), []byte("content"), 0o600))

	_, err := core.Init(dir, "TestDev")
	require.NoError(t, err)

	engine, err := core.Open(dir)
	require.NoError(t, err)
	defer engine.Close()

	_, err = engine.Scan()
	require.NoError(t, err)

	// Push with dry-run.
	result, err := engine.Push(media, true)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify nothing was written to media.
	entries, err := os.ReadDir(media)
	require.NoError(t, err)
	require.Empty(t, entries, "dry-run should not create any files on media")
}
