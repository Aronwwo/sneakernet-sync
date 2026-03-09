# ADR-003: Content-Addressed Blob Storage

## Status: Accepted

## Context

We need to store file content for transfer between devices. Options:
- Store files as-is with original names
- Store files by content hash (content-addressed)

## Decision

Use **content-addressed storage** with SHA-256 hashes. Files are stored at
`<hash_prefix>/<hash_suffix>` paths within the blob store.

## Rationale

- Automatic deduplication: identical files stored only once.
- Integrity verification: hash can be re-computed to verify blob integrity.
- Decoupled from file paths: content can be shared across renames.
- Standard approach used by Git, Docker, and similar systems.

## Consequences

- Requires hash computation on every file during scan (acceptable overhead).
- Two-level directory structure prevents filesystem issues with many files.
- Content is not human-readable in blob store (by design).
