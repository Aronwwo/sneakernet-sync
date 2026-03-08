# Thesis Notes

## Useful Points for Project and Implementation Chapters

### Architecture

- The system follows an archive-based reconciliation approach, similar to Unison
  but simplified for the engineering thesis scope.
- Three-way reconciliation detects conflicts correctly without relying on
  synchronized clocks between devices.
- Content-addressed storage provides automatic deduplication.

### Implementation Highlights

- **Go** chosen for its cross-platform support, static binaries, and strong
  standard library for file I/O and hashing.
- **SQLite (CGo-free)** provides ACID transactions without C dependencies,
  simplifying cross-compilation.
- **Atomic file writes** (temp file + rename) provide crash safety.
- **Pre-write validation** re-checks file state before writing to prevent
  overwriting concurrent user edits.

### Testing Strategy

- 42+ automated tests covering unit, integration, and E2E scenarios.
- All 7 synchronization rules have dedicated test cases.
- Integration test simulates full two-device sync through USB media.
- Race detector enabled in CI for concurrency safety.

### Key Design Decisions (ADRs)

1. Archive-based reconciliation (ADR-001)
2. SQLite for metadata (ADR-002)
3. Content-addressed blob storage (ADR-003)
4. Rename as delete + create (ADR-004)

### Diagrams for Thesis

Consider creating:
1. Architecture overview (package diagram)
2. Sync flow sequence diagram (init → scan → push → pull → apply)
3. Three-way reconciliation decision tree
4. USB media format structure
5. Data model ER diagram

### Metrics

- ~2200 lines of Go code (excluding tests)
- ~500 lines of test code
- 42+ test cases
- 11 internal packages
- 9 CLI commands
- 7 documented sync rules

### Comparison with Existing Tools

| Feature | sneakernet-sync | Unison | rsync | Syncthing |
|---------|----------------|--------|-------|-----------|
| Offline-first | ✓ | ✗ | ✗ | ✗ |
| No network required | ✓ | ✗ | ✗ | ✗ |
| Conflict detection | ✓ | ✓ | ✗ | ✓ |
| USB transport | ✓ | ✗ | ✗ | ✗ |
| Content-addressed | ✓ | ✗ | ✗ | ✗ |
| Three-way merge | ✓ | ✓ | ✗ | Partial |

### Limitations and Future Work

- Rename detection (heuristic matching by content hash)
- Multi-device sync (>2 devices)
- Encryption of USB media
- Partial/selective sync
- File permission tracking
- Network transport as extension
- Conflict resolution UI
