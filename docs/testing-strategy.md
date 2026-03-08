# Testing Strategy

## Overview

sneakernet-sync uses a layered testing approach: unit tests for individual
packages, integration tests for the full sync workflow, and manual E2E tests
for real-world scenarios.

## Test Levels

### Unit Tests

Each internal package has focused unit tests:

| Package           | Tests | Focus |
|-------------------|-------|-------|
| `hash`            | 2     | SHA-256 hashing, error handling |
| `scan`            | 3     | Directory walking, hidden file filtering, hash consistency |
| `fsops`           | 4     | Atomic write, directory creation |
| `blobstore`       | 4     | Store/retrieve, deduplication, missing blobs |
| `store/sqlite`    | 14    | All CRUD operations, schema, config, tombstones |
| `reconcile`       | 15    | All 7 sync rules, mixed scenarios, nil base |

### Integration Tests

The `core` package contains integration tests that exercise the full stack:

| Test                    | Description |
|-------------------------|-------------|
| `TestInit`              | Repository initialization |
| `TestInit_AlreadyInit`  | Double-init prevention |
| `TestScanAndStatus`     | Scan + status reporting |
| `TestDoctor`            | Integrity checking |
| `TestTwoDeviceSync`     | Full A→USB→B sync workflow |
| `TestPushDryRun`        | Dry-run doesn't write files |

### Manual E2E Testing

For real-world validation:

```bash
# Device A
mkdir -p /tmp/device-a && echo "hello" > /tmp/device-a/test.txt
sneakernet-sync init /tmp/device-a
sneakernet-sync scan /tmp/device-a
sneakernet-sync push /media/usb /tmp/device-a

# Device B
mkdir -p /tmp/device-b
sneakernet-sync init /tmp/device-b
sneakernet-sync scan /tmp/device-b
sneakernet-sync pull /media/usb /tmp/device-b
cat /tmp/device-b/test.txt  # → "hello"
```

## Running Tests

```bash
# All tests
make test

# Specific package
go test -v ./internal/reconcile/...

# With coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Test Conventions

- Use `t.TempDir()` for all filesystem operations (auto-cleanup).
- Use `require` from testify for assertions (fail fast).
- Tests are deterministic — no time-dependent assertions.
- Each test is independent — no shared state between tests.
- Named with `Test<Package>_<Scenario>` pattern.

## Coverage Goals

- Core sync rules: 100% of defined rules tested.
- Store operations: All CRUD paths tested.
- Error paths: Key error conditions covered (missing files, double init, etc.).
- Edge cases: Empty directories, hidden files, nil snapshots.

## CI Integration

Tests run on every push and PR via GitHub Actions:
- Linux, Windows, macOS
- Go 1.24
- Race detector enabled (`-race`)
- Coverage report generated
