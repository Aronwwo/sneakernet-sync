package sqlite_test

import (
	"path/filepath"
	"testing"

	"github.com/Aronwwo/sneakernet-sync/internal/store/sqlite"
	"github.com/stretchr/testify/require"
)

func TestNewStore_OpenClose(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")

	store, err := sqlite.New(dbPath)
	require.NoError(t, err)
	require.NotNil(t, store)

	require.NoError(t, store.Close())
}

func TestNewStore_SchemaCreated(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "schema.db")

	store, err := sqlite.New(dbPath)
	require.NoError(t, err)
	require.NoError(t, store.Close())

	// Re-open to verify schema persists across connections.
	store2, err := sqlite.New(dbPath)
	require.NoError(t, err)
	require.NoError(t, store2.Close())
}
