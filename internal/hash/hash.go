// Package hash provides file hashing utilities.
package hash

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
)

// FileHash computes the SHA-256 hash of the file at path and returns it as a
// lowercase hex-encoded string.
func FileHash(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
