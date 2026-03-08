# Architecture

## Overview

sneakernet-sync is an offline-first file synchronization tool that uses external
storage media (e.g., USB drives) to transfer changes between computers without
requiring network connectivity.

## Architecture Style

The system follows an **archive-based reconciliation** approach inspired by
Unison. Each device maintains its own identity and snapshot history. Changes are
detected relative to a common ancestor (the last known synchronized state),
enabling three-way reconciliation.

## High-Level Data Flow

```
Device A                    USB Media                   Device B
┌──────────┐              ┌──────────┐              ┌──────────┐
│ Sync Root │──scan───→   │          │              │ Sync Root │
│           │  snapshot   │ .offsync/│              │           │
│ .sneakernet│──push──→   │  blobs/  │──pull──→    │ .sneakernet│
│  meta.db  │             │  snaps/  │  reconcile  │  meta.db  │
│  blobs/   │             │ manifest │  apply      │  blobs/   │
└──────────┘              └──────────┘              └──────────┘
```

## Package Structure

```
cmd/synccli/          CLI entry point (cobra commands)
internal/
├── core/             Engine: orchestrates all operations
├── scan/             Directory walking and change detection
├── archive/          Snapshot management (persistent via SQLite)
├── reconcile/        Three-way reconciliation engine
├── conflict/         Conflict detection and persistence
├── apply/            Safe filesystem operations with pre-write validation
├── blobstore/        Content-addressed file storage (SHA-256)
├── transport/        USB media format: export/import
├── fsops/            Atomic filesystem helpers
├── hash/             SHA-256 file hashing
└── store/sqlite/     SQLite metadata store
```

## Key Design Decisions

1. **Offline-first**: No network dependency. USB drive is the transport.
2. **Content-addressed storage**: Files stored by SHA-256 hash for deduplication.
3. **Three-way reconciliation**: Changes detected against common ancestor, not
   naive timestamp comparison.
4. **Pre-write validation**: Before writing any file, re-check that local state
   hasn't changed since the last scan.
5. **Atomic writes**: Use temp file + rename pattern for crash safety.
6. **SQLite metadata**: Durable, transactional storage for snapshots and state.

## Sync Lifecycle

1. `init` — Create `.sneakernet/` metadata directory, generate device ID.
2. `scan` — Walk directory tree, compute SHA-256 hashes, store blobs, take snapshot.
3. `push` — Export snapshot + blobs to USB media (`<media>/.offsync/`).
4. `pull` — Import remote snapshot from USB, reconcile, detect conflicts, apply.
5. `sync` — Convenience: scan + push + pull in one step.

## Conflict Model

Conflicts are detected when:
- Both sides modified the same file differently since common ancestor.
- One side deleted a file the other side modified.
- Both sides independently created the same path with different content.

Conflicts are never auto-resolved. The user must explicitly choose a resolution.

## Security Considerations

- No encryption in MVP (documented as future work).
- Lock file on USB media prevents importing incomplete exports.
- Pre-write validation prevents overwriting concurrent local edits.
