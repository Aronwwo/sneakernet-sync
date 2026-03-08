// Package conflict provides conflict detection for bidirectional sync.
package conflict

import (
	"time"

	"github.com/Aronwwo/sneakernet-sync/internal/archive"
)

// ConflictInfo describes a detected conflict for a single file.
type ConflictInfo struct {
	RelPath      string
	LocalHash    string
	RemoteHash   string
	LocalDevice  string
	RemoteDevice string
	DetectedAt   time.Time
}

// Detect compares local, remote, and archive snapshots and returns a list of
// conflicts. A conflict occurs when both local and remote have changed a file
// relative to the archive snapshot and their new hashes differ.
func Detect(localState, remoteState, archiveState archive.Snapshot) []ConflictInfo {
	var conflicts []ConflictInfo

	for path, localHash := range localState.Files {
		remoteHash, inRemote := remoteState.Files[path]
		archiveHash, inArchive := archiveState.Files[path]

		if !inRemote {
			continue
		}

		localChanged := !inArchive || localHash != archiveHash
		remoteChanged := !inArchive || remoteHash != archiveHash

		if localChanged && remoteChanged && localHash != remoteHash {
			conflicts = append(conflicts, ConflictInfo{
				RelPath:    path,
				LocalHash:  localHash,
				RemoteHash: remoteHash,
				DetectedAt: time.Now(),
			})
		}
	}

	return conflicts
}
