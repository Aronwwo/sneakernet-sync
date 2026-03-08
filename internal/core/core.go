// Package core ties together the internal packages to implement the sync
// engine.
package core

import (
	"fmt"

	"github.com/Aronwwo/sneakernet-sync/internal/store/sqlite"
)

// Engine orchestrates push/pull/sync operations.
type Engine struct {
	rootDir  string
	mediaDir string
	store    *sqlite.Store
}

// New creates a new Engine.
func New(rootDir, mediaDir string, store *sqlite.Store) *Engine {
	return &Engine{
		rootDir:  rootDir,
		mediaDir: mediaDir,
		store:    store,
	}
}

// Push exports local changes to the external media directory.
func (e *Engine) Push() error {
	return fmt.Errorf("not implemented yet")
}

// Pull imports changes from the external media directory.
func (e *Engine) Pull() error {
	return fmt.Errorf("not implemented yet")
}

// Sync performs a full bidirectional sync cycle.
func (e *Engine) Sync() error {
	return fmt.Errorf("not implemented yet")
}
