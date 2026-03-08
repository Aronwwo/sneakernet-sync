// Package blobstore manages content-addressed file storage.
// Files are stored by their SHA-256 hash in a two-level directory structure
// (first 2 hex chars / remaining hex chars).
package blobstore

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/Aronwwo/sneakernet-sync/internal/fsops"
	"github.com/Aronwwo/sneakernet-sync/internal/hash"
)

// BlobStore stores and retrieves files by their content hash.
type BlobStore struct {
	basePath string
}

// New creates a new BlobStore rooted at basePath.
func New(basePath string) *BlobStore {
	return &BlobStore{basePath: basePath}
}

// Store copies the file at srcPath into the blob store and returns its
// content hash. If the blob already exists, it is not overwritten.
func (b *BlobStore) Store(srcPath string) (string, error) {
	contentHash, err := hash.FileHash(srcPath)
	if err != nil {
		return "", fmt.Errorf("hash file: %w", err)
	}

	blobPath := b.blobPath(contentHash)

	// Skip if blob already exists (deduplication).
	if _, err := os.Stat(blobPath); err == nil {
		return contentHash, nil
	}

	if err := fsops.EnsureDir(filepath.Dir(blobPath)); err != nil {
		return "", fmt.Errorf("create blob dir: %w", err)
	}

	data, err := os.ReadFile(srcPath)
	if err != nil {
		return "", fmt.Errorf("read source: %w", err)
	}

	if err := fsops.AtomicWrite(blobPath, data); err != nil {
		return "", fmt.Errorf("write blob: %w", err)
	}

	return contentHash, nil
}

// StoreReader stores data from a reader and returns the content hash.
func (b *BlobStore) StoreReader(contentHash string, r io.Reader) error {
	blobPath := b.blobPath(contentHash)

	if _, err := os.Stat(blobPath); err == nil {
		return nil // already exists
	}

	if err := fsops.EnsureDir(filepath.Dir(blobPath)); err != nil {
		return fmt.Errorf("create blob dir: %w", err)
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("read data: %w", err)
	}

	return fsops.AtomicWrite(blobPath, data)
}

// Retrieve copies the blob identified by hash to destPath.
func (b *BlobStore) Retrieve(contentHash, destPath string) error {
	blobPath := b.blobPath(contentHash)

	src, err := os.Open(blobPath)
	if err != nil {
		return fmt.Errorf("open blob %q: %w", contentHash, err)
	}
	defer src.Close()

	if err := fsops.EnsureDir(filepath.Dir(destPath)); err != nil {
		return fmt.Errorf("create dest dir: %w", err)
	}

	data, err := io.ReadAll(src)
	if err != nil {
		return fmt.Errorf("read blob: %w", err)
	}

	return fsops.AtomicWrite(destPath, data)
}

// Has checks whether a blob with the given hash exists.
func (b *BlobStore) Has(contentHash string) bool {
	_, err := os.Stat(b.blobPath(contentHash))
	return err == nil
}

// Path returns the filesystem path for a blob.
func (b *BlobStore) Path(contentHash string) string {
	return b.blobPath(contentHash)
}

func (b *BlobStore) blobPath(contentHash string) string {
	if len(contentHash) < 4 {
		return filepath.Join(b.basePath, contentHash)
	}
	return filepath.Join(b.basePath, contentHash[:2], contentHash[2:])
}
