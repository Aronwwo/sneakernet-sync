// Package blobstore manages content-addressed file storage.
package blobstore

import "fmt"

// BlobStore stores and retrieves files by their content hash.
type BlobStore struct {
	basePath string
}

// New creates a new BlobStore rooted at basePath.
func New(basePath string) *BlobStore {
	return &BlobStore{basePath: basePath}
}

// Store copies the file at srcPath into the blob store and returns its
// content hash.
func (b *BlobStore) Store(_ string) (string, error) {
	return "", fmt.Errorf("not implemented yet")
}

// Retrieve copies the blob identified by hash to destPath.
func (b *BlobStore) Retrieve(_, _ string) error {
	return fmt.Errorf("not implemented yet")
}
