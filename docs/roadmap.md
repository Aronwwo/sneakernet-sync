# Roadmap

## Iteration 1: Foundation ✅

**Goal**: Project scaffold, CI, basic structure.

**Scope**:
- Go module setup
- CLI skeleton with cobra
- SQLite store with schema
- SHA-256 hashing
- Directory scanner
- GitHub Actions CI
- Basic documentation

**Definition of Done**: Build passes, CI green, tests run.

## Iteration 2: Core Domain ✅

**Goal**: Implement snapshot system and reconciliation engine.

**Scope**:
- Archive/snapshot management
- Three-way reconciliation with all sync rules
- Conflict detection
- Content-addressed blob storage
- Comprehensive unit tests for reconciliation

**Definition of Done**: All 7 sync rules tested and passing.

## Iteration 3: Offline Transport ✅

**Goal**: Implement USB media export/import.

**Scope**:
- USB media format (manifest, blobs, snapshots)
- Export (push) command
- Import (pull) command
- Lock file for crash safety
- Integration test for two-device sync

**Definition of Done**: Full A→USB→B sync workflow passing.

## Iteration 4: Safe Apply & CLI ✅

**Goal**: Safe filesystem operations and polished CLI.

**Scope**:
- Pre-write validation before apply
- Atomic file writes
- All CLI commands wired and functional
- Dry-run mode
- Doctor command for integrity checks
- Conflict listing and resolution commands

**Definition of Done**: CLI usable for real sync scenarios.

## Iteration 5: Documentation & Stabilization ✅

**Goal**: Complete documentation for thesis defense.

**Scope**:
- Architecture documentation
- Data model documentation
- Sync rules documentation
- USB format documentation
- Testing strategy
- ADRs for key decisions
- Thesis notes

**Definition of Done**: All docs written, all tests passing.

## Future Iterations (Post-MVP)

### Iteration 6: Polish
- Better CLI output (colors, progress)
- Ignore patterns
- Verbose/quiet modes
- Performance optimization for large trees

### Iteration 7: Stretch Goals
- Rename detection
- Encryption
- Partial sync
- Network transport extension
