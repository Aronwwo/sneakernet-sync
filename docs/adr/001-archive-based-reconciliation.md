# ADR-001: Archive-Based Reconciliation

## Status: Accepted

## Context

We need a synchronization model for offline file sync between two devices.
The naive approach (copy newer file) fails when both devices modify the same
file, or when clocks are unreliable.

## Decision

Use **archive-based three-way reconciliation** inspired by Unison. Each device
maintains snapshots of the file tree. Changes are detected relative to the
last known common state (base/ancestor), not by comparing timestamps.

## Consequences

- Correct conflict detection without relying on synchronized clocks.
- Requires storing snapshot history, increasing storage overhead.
- More complex than timestamp-based approaches, but fundamentally more correct.
- Rename detection requires additional heuristics (deferred to future work).
