// Package scan provides directory scanning and change detection.
package scan

import (
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"github.com/Aronwwo/sneakernet-sync/internal/hash"
)

// ChangeState describes the sync state of a file.
type ChangeState int

const (
	// StateNew indicates a file that has not been synced before.
	StateNew ChangeState = iota
	// StateModified indicates a file whose content has changed since last sync.
	StateModified
	// StateDeleted indicates a file that was removed since last sync.
	StateDeleted
	// StateUnchanged indicates a file that has not changed since last sync.
	StateUnchanged
)

// FileChange holds metadata about a single file detected during a scan.
type FileChange struct {
	RelPath     string
	ContentHash string
	Size        int64
	ModTime     time.Time
	State       ChangeState
}

// Scanner walks a root directory and reports file changes.
type Scanner struct {
	rootDir string
}

// New creates a new Scanner rooted at rootDir.
func New(rootDir string) *Scanner {
	return &Scanner{rootDir: rootDir}
}

// Scan walks the root directory and returns a FileChange entry for every
// non-hidden file found. Hidden files and directories (names starting with
// ".") are skipped. All returned entries have State set to StateNew; callers
// are responsible for comparing against stored state to determine the actual
// change state.
func (s *Scanner) Scan() ([]FileChange, error) {
	var changes []FileChange

	err := filepath.WalkDir(s.rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		name := d.Name()

		// Skip hidden files and directories.
		if strings.HasPrefix(name, ".") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if d.IsDir() {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(s.rootDir, path)
		if err != nil {
			return err
		}

		contentHash, err := hash.FileHash(path)
		if err != nil {
			return err
		}

		changes = append(changes, FileChange{
			RelPath:     relPath,
			ContentHash: contentHash,
			Size:        info.Size(),
			ModTime:     info.ModTime(),
			State:       StateNew,
		})

		return nil
	})

	return changes, err
}
