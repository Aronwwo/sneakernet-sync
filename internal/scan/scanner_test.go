package scan_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Aronwwo/sneakernet-sync/internal/scan"
	"github.com/stretchr/testify/require"
)

func TestScan_BasicFiles(t *testing.T) {
	tmp := t.TempDir()

	files := map[string]string{
		"a.txt":        "hello",
		"sub/b.txt":    "world",
		".hidden":      "should be skipped",
		".hiddendir/c": "also skipped",
	}

	for rel, content := range files {
		full := filepath.Join(tmp, rel)
		require.NoError(t, os.MkdirAll(filepath.Dir(full), 0o755))
		require.NoError(t, os.WriteFile(full, []byte(content), 0o600))
	}

	s := scan.New(tmp)
	changes, err := s.Scan()
	require.NoError(t, err)

	// Only non-hidden files and directories should be returned.
	paths := make(map[string]bool)
	for _, c := range changes {
		paths[filepath.ToSlash(c.RelPath)] = true
	}

	require.True(t, paths["a.txt"], "expected a.txt")
	require.True(t, paths["sub/b.txt"], "expected sub/b.txt")
	require.True(t, paths["sub"], "expected sub directory")
	require.False(t, paths[".hidden"], "hidden file should be skipped")
	require.False(t, paths[".hiddendir/c"], "hidden dir contents should be skipped")
	require.Equal(t, 3, len(changes)) // a.txt, sub/, sub/b.txt
}

func TestScan_EmptyDirectory(t *testing.T) {
	tmp := t.TempDir()

	s := scan.New(tmp)
	changes, err := s.Scan()
	require.NoError(t, err)
	require.Empty(t, changes)
}

func TestScan_HashConsistency(t *testing.T) {
	tmp := t.TempDir()
	content := []byte("consistent content")
	require.NoError(t, os.WriteFile(filepath.Join(tmp, "f.txt"), content, 0o600))

	s := scan.New(tmp)
	changes, err := s.Scan()
	require.NoError(t, err)
	require.Len(t, changes, 1)
	require.NotEmpty(t, changes[0].ContentHash)
	require.Equal(t, int64(len(content)), changes[0].Size)
	require.Equal(t, scan.StateNew, changes[0].State)
}
