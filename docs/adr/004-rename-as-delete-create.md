# ADR-004: Rename Treated as Delete + Create

## Status: Accepted

## Context

Detecting file renames is complex. It requires matching deleted files with
newly created files based on content hash similarity, path similarity, or
other heuristics. This adds significant complexity to the reconciliation
engine.

## Decision

In MVP, **rename is treated as delete + create**. If a file is renamed from
`a.txt` to `b.txt`, the system sees:
1. `a.txt` deleted
2. `b.txt` created

## Rationale

- Simpler implementation, fewer edge cases.
- Correct behavior (no data loss), just suboptimal (re-transfers content).
- Content-addressed blob store means the blob isn't duplicated.
- Rename detection is listed as a stretch goal.

## Consequences

- Suboptimal transfer: blob already exists but system doesn't know it's a rename.
- Delete+create may trigger false conflicts in some scenarios.
- Acceptable trade-off for MVP scope.
