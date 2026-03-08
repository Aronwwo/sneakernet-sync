# sneakernet-sync

![CI](https://github.com/Aronwwo/sneakernet-sync/actions/workflows/ci.yml/badge.svg)

**Offline file synchronization tool via external storage media (USB)**

sneakernet-sync lets you keep directories in sync across air-gapped machines using a USB drive as the transport. No cloud, no network — just a drive and a CLI.

This is an engineering thesis project developed by a 2-person team following an iterative-incremental model with Scrum practices.

---

## Features (planned)

- **Change detection** — SHA-256 content-addressed file scanning
- **Metadata tracking** — SQLite database stores file state and sync history
- **Content-addressed storage** — blobs stored by hash, deduplication by default
- **Conflict detection** — three-way merge to identify concurrent edits
- **Bidirectional sync** — push local changes and pull remote changes in a single command

---

## Tech Stack

| Component         | Technology                                      |
|-------------------|-------------------------------------------------|
| Language          | Go 1.22+                                        |
| CLI framework     | [cobra](https://github.com/spf13/cobra)         |
| SQLite driver     | [ncruces/go-sqlite3](https://github.com/ncruces/go-sqlite3) (CGo-free) |
| Hashing           | `crypto/sha256` (stdlib)                        |
| Logging           | `log/slog` (stdlib)                             |
| CI                | GitHub Actions (Linux, Windows, macOS)          |

---

## Quick Start

```bash
# Build
go build ./cmd/synccli

# Run
./sneakernet-sync --help
```

---

## CLI Commands

| Command              | Description                                   |
|----------------------|-----------------------------------------------|
| `init`               | Initialize a sync repository in a directory   |
| `scan`               | Scan directory for changes                    |
| `status`             | Show current sync status                      |
| `push`               | Export changes to external media              |
| `pull`               | Import changes from external media            |
| `sync`               | Full sync cycle (push + pull)                 |
| `resolve`            | Resolve sync conflicts                        |
| `doctor`             | Verify repository integrity                   |

---

## Project Structure

```
sneakernet-sync/
├── cmd/synccli/          # CLI entry point
├── internal/
│   ├── core/             # Sync engine
│   ├── scan/             # Directory scanning & change detection
│   ├── archive/          # Snapshot management
│   ├── reconcile/        # Three-way reconciliation
│   ├── conflict/         # Conflict detection
│   ├── store/sqlite/     # SQLite metadata store
│   ├── blobstore/        # Content-addressed blob storage
│   ├── hash/             # SHA-256 file hashing
│   └── fsops/            # Atomic filesystem helpers
├── docs/                 # Project documentation
├── .github/workflows/    # GitHub Actions CI
└── Makefile
```

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
