package blobstore_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Aronwwo/sneakernet-sync/internal/blobstore"
	"github.com/stretchr/testify/require"
)

func TestBlobStore_StoreAndRetrieve(t *testing.T) {
	tmp := t.TempDir()
	blobDir := filepath.Join(tmp, "blobs")
	bs := blobstore.New(blobDir)

	// Create a source file.
	srcDir := filepath.Join(tmp, "src")
	require.NoError(t, os.MkdirAll(srcDir, 0o755))
	srcFile := filepath.Join(srcDir, "test.txt")
	content := []byte("hello blob store")
	require.NoError(t, os.WriteFile(srcFile, content, 0o600))

	// Store it.
	hash, err := bs.Store(srcFile)
	require.NoError(t, err)
	require.NotEmpty(t, hash)

	// Verify the blob exists.
	require.True(t, bs.Has(hash))

	// Retrieve it.
	destFile := filepath.Join(tmp, "retrieved.txt")
	require.NoError(t, bs.Retrieve(hash, destFile))

	// Verify content matches.
	got, err := os.ReadFile(destFile)
	require.NoError(t, err)
	require.Equal(t, content, got)
}

func TestBlobStore_Deduplication(t *testing.T) {
	tmp := t.TempDir()
	bs := blobstore.New(filepath.Join(tmp, "blobs"))

	content := []byte("duplicate content")
	for i := 0; i < 3; i++ {
		f := filepath.Join(tmp, "src.txt")
		require.NoError(t, os.WriteFile(f, content, 0o600))

		hash, err := bs.Store(f)
		require.NoError(t, err)
		require.NotEmpty(t, hash)
	}

	// Only one blob should exist (deduplication).
	blobCount := 0
	err := filepath.Walk(filepath.Join(tmp, "blobs"), func(_ string, info os.FileInfo, _ error) error {
		if info != nil && !info.IsDir() {
			blobCount++
		}
		return nil
	})
	require.NoError(t, err)
	require.Equal(t, 1, blobCount)
}

func TestBlobStore_HasMissing(t *testing.T) {
	tmp := t.TempDir()
	bs := blobstore.New(filepath.Join(tmp, "blobs"))
	require.False(t, bs.Has("nonexistent-hash"))
}

func TestBlobStore_RetrieveMissing(t *testing.T) {
	tmp := t.TempDir()
	bs := blobstore.New(filepath.Join(tmp, "blobs"))

	err := bs.Retrieve("nonexistent-hash", filepath.Join(tmp, "out.txt"))
	require.Error(t, err)
}
