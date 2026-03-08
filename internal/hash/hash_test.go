package hash_test

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/Aronwwo/sneakernet-sync/internal/hash"
	"github.com/stretchr/testify/require"
)

func TestFileHash(t *testing.T) {
	content := []byte("hello sneakernet-sync")

	tmp := t.TempDir()
	path := filepath.Join(tmp, "test.txt")
	require.NoError(t, os.WriteFile(path, content, 0o600))

	got, err := hash.FileHash(path)
	require.NoError(t, err)

	sum := sha256.Sum256(content)
	want := hex.EncodeToString(sum[:])

	require.Equal(t, want, got)
}

func TestFileHash_NotExist(t *testing.T) {
	_, err := hash.FileHash("/no/such/file")
	require.Error(t, err)
}
