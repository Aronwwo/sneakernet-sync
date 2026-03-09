# ADR-002: SQLite for Local Metadata

## Status: Accepted

## Context

We need persistent, queryable storage for file metadata, snapshots, conflicts,
and device configuration. Options considered:
- Flat JSON/TOML files
- SQLite
- BoltDB / BadgerDB

## Decision

Use **SQLite** via the `ncruces/go-sqlite3` driver (CGo-free, uses Wasm).

## Rationale

- ACID transactions for data integrity.
- Rich querying capabilities.
- Single-file database, easy to backup/inspect.
- CGo-free driver avoids cross-compilation issues.
- Well-understood technology for academic work.

## Consequences

- Slight initialization overhead (~9s for first Wasm compilation, cached after).
- Single-file storage simplifies backup but limits concurrent access.
- The Wasm-based driver is slower than CGo alternatives but sufficient for our
  use case (metadata, not bulk data).
