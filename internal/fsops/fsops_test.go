package fsops_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Aronwwo/sneakernet-sync/internal/fsops"
	"github.com/stretchr/testify/require"
)

func TestAtomicWrite(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "test.txt")
	content := []byte("atomic write content")

	require.NoError(t, fsops.AtomicWrite(path, content))

	got, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, content, got)
}

func TestAtomicWrite_Overwrite(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "test.txt")

	require.NoError(t, os.WriteFile(path, []byte("old"), 0o600))
	require.NoError(t, fsops.AtomicWrite(path, []byte("new")))

	got, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, []byte("new"), got)
}

func TestEnsureDir(t *testing.T) {
	tmp := t.TempDir()
	nested := filepath.Join(tmp, "a", "b", "c")

	require.NoError(t, fsops.EnsureDir(nested))

	info, err := os.Stat(nested)
	require.NoError(t, err)
	require.True(t, info.IsDir())
}

func TestEnsureDir_Existing(t *testing.T) {
	tmp := t.TempDir()
	require.NoError(t, fsops.EnsureDir(tmp))
}
