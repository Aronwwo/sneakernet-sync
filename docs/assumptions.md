# Assumptions

This document records assumptions made during development when requirements were
ambiguous or incomplete.

## A1: Two-Device Sync Only (MVP)

The MVP targets synchronization between exactly two devices. Multi-device
topologies (A↔B, B↔C, A↔C) are out of scope and listed as future work.

## A2: Rename = Delete + Create

File renames are treated as a deletion of the old path and creation of the new
path. Rename detection heuristics are a stretch goal.

## A3: No Binary Merge

The system does not attempt to merge binary files or even text files at the
line level. All conflict resolution is at the whole-file level.

## A4: Content Hash as Source of Truth

File identity is determined by content hash (SHA-256), not by path or
modification time. Two files with identical content at different paths are
independent.

## A5: No Permission/Ownership Tracking

File permissions, ownership, and extended attributes are NOT tracked in MVP.
Files are created with default permissions.

## A6: UTF-8 Paths Only

The system assumes file paths are valid UTF-8 strings. Paths with invalid
encoding may cause undefined behavior.

## A7: No Symlinks

Symbolic links are not followed or tracked. They are silently skipped during
scanning.

## A8: Single USB Session

Each export to USB creates a new session. Previous sessions on the same USB
are overwritten (not versioned). History is maintained in local SQLite only.

## A9: Hidden Files (.dotfiles) Skipped

Files and directories starting with `.` are excluded from scanning. This
includes the `.sneakernet` metadata directory.

## A10: No Encryption on USB

Data on USB media is stored in plaintext. Encryption is a stretch goal.

## A11: File Size Limit

No explicit file size limit, but very large files may cause high memory usage
during blob operations. Streaming is used where possible.

## A12: Filesystem Must Support Rename

Atomic write depends on `os.Rename()` working within the same filesystem. This
is true for all common filesystems but may fail across mount points.

## A13: Single Concurrent Sync

The system does not support multiple concurrent sync operations on the same
directory. The lock file on USB media provides partial protection, but local
operations are not locked.
