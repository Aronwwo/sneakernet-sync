# USB Media Format

## Overview

sneakernet-sync uses a versioned directory structure on external media (USB
drive, SD card, etc.) to exchange sync data between devices.

## Directory Structure

```
<media_root>/
└── .offsync/
    ├── manifest.json              # Session metadata
    ├── blobs/                     # Content-addressed file storage
    │   ├── ab/                    # First 2 hex chars of hash
    │   │   └── c123def456...      # Remaining hex chars
    │   └── ...
    ├── snapshots/                 # One snapshot file per device
    │   └── <device_id>.json
    └── lock                       # Lock file (present only during export)
```

## manifest.json

```json
{
  "schema_version": 1,
  "device_id": "e37264e5d8c250eac18f34a8dcb3812f",
  "device_name": "DeviceA",
  "snapshot_id": "90bf8ae71c90dc34ef515e17ac633e4c",
  "created_at": "2025-01-15T10:30:00Z",
  "file_count": 42
}
```

Fields:
- `schema_version`: Integer version of the media format. Currently `1`.
- `device_id`: Hex ID of the exporting device.
- `device_name`: Human-readable name.
- `snapshot_id`: ID of the exported snapshot.
- `created_at`: ISO 8601 timestamp of the export.
- `file_count`: Number of files in the snapshot.

## snapshots/<device_id>.json

Array of file entries:

```json
[
  {
    "rel_path": "docs/readme.txt",
    "content_hash": "abc123...",
    "size": 1024,
    "mod_time": "2025-01-15T10:00:00Z",
    "is_dir": false,
    "exists": true
  },
  {
    "rel_path": "src",
    "content_hash": "DIR",
    "size": 0,
    "mod_time": "0001-01-01T00:00:00Z",
    "is_dir": true,
    "exists": true
  }
]
```

## blobs/

Content-addressed storage using SHA-256 hashes. Files are stored in a two-level
directory structure for filesystem efficiency:

- Hash: `abcdef1234567890...` → Path: `blobs/ab/cdef1234567890...`

Only file blobs are stored. Directories have `content_hash = "DIR"` and no blob.

## Lock File

The `lock` file is created at the start of an export and removed upon
completion. If the lock file exists when importing, it indicates the export was
interrupted and the data may be incomplete.

Import will refuse to proceed if a lock file is present. The user can manually
remove the lock file to force import after verifying data integrity.

## Schema Version

The `schema_version` field in `manifest.json` allows future format changes
while maintaining backward compatibility. The current version is `1`.

Import will reject media with a schema version higher than supported.

## Design Properties

1. **Human-readable**: JSON format, inspectable with any text editor.
2. **Crash-safe**: Lock file prevents importing incomplete exports.
3. **Extensible**: Schema version allows future additions.
4. **Deduplication**: Content-addressed blobs avoid redundant copies.
5. **Portable**: No OS-specific features; works on any filesystem.
