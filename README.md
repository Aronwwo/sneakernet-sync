# sneakernet-sync

![CI](https://github.com/Aronwwo/sneakernet-sync/actions/workflows/ci.yml/badge.svg)

**Offline file synchronization tool via external storage media (USB)**

sneakernet-sync lets you keep directories in sync across air-gapped machines using a USB drive as the transport. No cloud, no network — just a drive and a CLI.

This is an engineering thesis project developed by a 2-person team following an iterative-incremental model.

---

## Features

- **Change detection** — SHA-256 content-addressed file scanning
- **Metadata tracking** — SQLite database stores file state and sync history
- **Content-addressed storage** — blobs stored by hash, deduplication by default
- **Three-way reconciliation** — detects conflicts relative to common ancestor
- **Conflict detection** — 7 rules for safe conflict handling (no silent overwrites)
- **Bidirectional sync** — push local changes and pull remote changes
- **Safe apply** — pre-write validation prevents overwriting concurrent edits
- **Dry-run mode** — preview operations without writing anything
- **USB media format** — versioned, human-readable, crash-safe

---

## Tech Stack

| Component         | Technology                                      |
|-------------------|-------------------------------------------------|
| Language          | Go 1.24+                                        |
| CLI framework     | [cobra](https://github.com/spf13/cobra)         |
| SQLite driver     | [ncruces/go-sqlite3](https://github.com/ncruces/go-sqlite3) (CGo-free) |
| Hashing           | `crypto/sha256` (stdlib)                        |
| CI                | GitHub Actions (Linux, Windows, macOS)          |

---

## Quick Start

```bash
# Build
go build -o sneakernet-sync ./cmd/synccli

# Initialize a sync repository
sneakernet-sync init ~/my-project

# Scan for changes
sneakernet-sync scan ~/my-project

# Check status
sneakernet-sync status ~/my-project

# Export to USB drive
sneakernet-sync push /media/usb ~/my-project

# On another machine: import from USB
sneakernet-sync pull /media/usb ~/my-project

# Full sync cycle (scan + push + pull)
sneakernet-sync sync /media/usb ~/my-project

# Check for conflicts
sneakernet-sync conflicts ~/my-project

# Verify integrity
sneakernet-sync doctor ~/my-project
```

---

## CLI Commands

| Command              | Description                                   |
|----------------------|-----------------------------------------------|
| `init [dir]`         | Initialize a sync repository in a directory   |
| `scan [dir]`         | Scan directory for changes and take snapshot   |
| `status [dir]`       | Show current sync status                      |
| `push <media> [dir]` | Export changes to external media              |
| `pull <media> [dir]` | Import changes from external media            |
| `sync <media> [dir]` | Full sync cycle (scan + push + pull)          |
| `conflicts [dir]`    | List unresolved sync conflicts                |
| `resolve <id>`       | Resolve a sync conflict                       |
| `doctor [dir]`       | Verify repository integrity                   |

All write commands support `--dry-run` to preview without changes.

---

## Project Structure

```
sneakernet-sync/
├── cmd/synccli/          # CLI entry point
├── internal/
│   ├── core/             # Sync engine (orchestration)
│   ├── scan/             # Directory scanning & change detection
│   ├── archive/          # Snapshot management
│   ├── reconcile/        # Three-way reconciliation
│   ├── conflict/         # Conflict detection & persistence
│   ├── apply/            # Safe filesystem apply with validation
│   ├── transport/        # USB media export/import
│   ├── store/sqlite/     # SQLite metadata store
│   ├── blobstore/        # Content-addressed blob storage
│   ├── hash/             # SHA-256 file hashing
│   └── fsops/            # Atomic filesystem helpers
├── docs/                 # Project documentation
│   ├── architecture.md
│   ├── data-model.md
│   ├── sync-rules.md
│   ├── usb-format.md
│   ├── testing-strategy.md
│   ├── assumptions.md
│   ├── adr/              # Architecture Decision Records
│   ├── backlog.md
│   ├── roadmap.md
│   └── ...
├── .github/workflows/    # GitHub Actions CI
└── Makefile
```

---

## Documentation

- [Architecture](docs/architecture.md)
- [Data Model](docs/data-model.md)
- [Sync Rules](docs/sync-rules.md)
- [USB Media Format](docs/usb-format.md)
- [Testing Strategy](docs/testing-strategy.md)
- [Assumptions](docs/assumptions.md)
- [Backlog](docs/backlog.md)
- [Roadmap](docs/roadmap.md)

---

## Development

```bash
make test    # Run tests with race detector
make lint    # Run golangci-lint
make build   # Build binary
make all     # lint + test + build
```

---

## Team

- [@Aronwwo](https://github.com/Aronwwo)
- [@PawelMierzwa](https://github.com/PawelMierzwa)

---

## License

[MIT](LICENSE)
