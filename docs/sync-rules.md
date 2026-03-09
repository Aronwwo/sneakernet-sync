# Synchronization Rules

## Overview

sneakernet-sync uses **three-way reconciliation** to determine how to merge
changes between two devices. The three inputs are:

1. **Local snapshot** — current state of the local device.
2. **Remote snapshot** — state received from the other device via USB media.
3. **Base snapshot** — the last known common ancestor (previous sync state).

## Rules

### Rule 1: One-Side Change

If a file changed only on **one side** relative to the base, the change is
propagated to the other side.

| Local | Remote | Base | Action |
|-------|--------|------|--------|
| v2    | v1     | v1   | Copy local → remote |
| v1    | v2     | v1   | Copy remote → local |

### Rule 2: Both Sides Modified Differently

If both sides modified the same file differently relative to the base, this is
a **conflict**. No automatic resolution is performed.

| Local | Remote | Base | Action |
|-------|--------|------|--------|
| v2    | v3     | v1   | CONFLICT |

### Rule 3: Delete vs. Modify

If one side deleted a file and the other modified it, this is a **conflict**.
User data is never silently lost.

| Local   | Remote | Base | Action |
|---------|--------|------|--------|
| deleted | v2     | v1   | CONFLICT (delete-modify) |
| v2      | deleted| v1   | CONFLICT (modify-delete) |

### Rule 4: Both Deleted

If both sides deleted the same file, no action is needed.

| Local   | Remote  | Base | Action |
|---------|---------|------|--------|
| deleted | deleted | v1   | None   |

### Rule 5: Both Created Different Content

If both sides independently created a file at the same path with different
content (no base exists), this is a **conflict**.

| Local | Remote | Base      | Action |
|-------|--------|-----------|--------|
| v1    | v2     | not exist | CONFLICT (create-create) |

### Rule 6: Both Created Same Content

If both sides independently created the same file with identical content,
no action is needed (convergent creation).

| Local | Remote | Base      | Action |
|-------|--------|-----------|--------|
| v1    | v1     | not exist | None   |

### Rule 7: Both Modified Identically

If both sides modified a file to the same new content, no action is needed
(convergent modification).

| Local | Remote | Base | Action |
|-------|--------|------|--------|
| v2    | v2     | v1   | None   |

## New Files and Directories

- New file on one side → copy to the other side.
- New directory on one side → create on the other side.

## Deletions

- Deletion on one side, unchanged on other → propagate deletion.
- Deletion on one side, modified on other → conflict.

## Rename Detection

In MVP, rename is treated as **delete + create**. This is documented as a
known limitation. Rename detection is a stretch goal.

## Pre-Write Validation

Before applying any write operation locally, the system re-checks the current
state of the file on disk. If the file has changed since the last scan (e.g.,
the user edited it while sync was in progress), the operation is **aborted**
for that file and an error is reported. This prevents overwriting concurrent
user edits.

## Conflict Resolution

Conflicts are persisted in the database and must be resolved explicitly by the
user. Available resolution strategies:

- `keep-local` — keep the local version.
- `keep-remote` — keep the remote version.
- `manual` — user resolves externally and marks as resolved.

## Metadata Handling in MVP

- **File content**: Tracked via SHA-256 hash. Primary basis for change detection.
- **Modification time**: Recorded but NOT used for conflict resolution (only
  content hash matters).
- **File size**: Recorded for informational purposes.
- **Permissions**: NOT tracked in MVP.
- **Symlinks**: NOT supported in MVP.
