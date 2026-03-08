# Backlog

## MVP (Must Have)

- [x] Repository initialization (`init` command)
- [x] Device registration with unique ID
- [x] Local SQLite metadata store with schema
- [x] Directory scanning with SHA-256 hashing
- [x] Snapshot system (take, store, retrieve)
- [x] Content-addressed blob storage
- [x] Change detection (create, modify, delete, directory)
- [x] Three-way reconciliation engine
- [x] All 7 sync rules implemented
- [x] Conflict detection and persistence
- [x] USB media export (push)
- [x] USB media import (pull)
- [x] Safe apply with pre-write validation
- [x] CLI commands: init, scan, status, push, pull, sync, resolve, conflicts, doctor
- [x] Dry-run mode
- [x] Unit tests for all core packages
- [x] Integration test (two-device sync)
- [x] CI pipeline (build, test, lint)
- [x] Architecture documentation
- [x] Data model documentation
- [x] Sync rules documentation
- [x] USB format documentation

## Should Have

- [ ] Better CLI output formatting (colors, tables)
- [ ] Progress bars for large sync operations
- [ ] `--verbose` and `--quiet` flags
- [ ] Ignore patterns (`.syncignore` file)
- [ ] Sync session logging to sync_log table
- [ ] File permission tracking

## Nice to Have (Stretch Goals)

- [ ] Rename detection heuristic
- [ ] Partial/selective sync (subdirectory filtering)
- [ ] Pattern-based file filtering
- [ ] Simple text-based conflict resolution UI
- [ ] Encryption of USB media data
- [ ] Network transport extension (experimental)
- [ ] Web-based status dashboard

## Won't Have (Out of Scope)

- Full multi-device mesh synchronization
- Distributed P2P over Internet
- GUI application
- Automatic binary file merging
- Smart sync without explicit rules
