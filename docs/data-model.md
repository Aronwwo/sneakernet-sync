# Data Model

## Overview

sneakernet-sync uses SQLite as its local metadata store. The database resides at
`<sync_root>/.sneakernet/meta.db`.

## Entity-Relationship

```
devices 1──N files
devices 1──N snapshots
snapshots 1──N snapshot_entries
devices 1──N sync_log
conflicts (standalone)
tombstones (standalone)
config (key-value store)
```

## Tables

### devices

| Column     | Type | Description                |
|------------|------|----------------------------|
| device_id  | TEXT | Primary key, random hex ID |
| name       | TEXT | Human-readable device name |
| created_at | TEXT | ISO 8601 timestamp         |

### files

| Column       | Type    | Description                        |
|--------------|---------|------------------------------------|
| rel_path     | TEXT    | Relative path from sync root       |
| content_hash | TEXT    | SHA-256 hex hash of file content   |
| size         | INTEGER | File size in bytes                 |
| mod_time     | TEXT    | Last modification time (RFC3339)   |
| state        | INTEGER | Change state (0=new, 1=mod, etc.)  |
| device_id    | TEXT    | FK to devices                      |
| last_sync_at | TEXT    | Last sync timestamp                |
| is_dir       | INTEGER | 1 if directory, 0 if file          |
| exists_flag  | INTEGER | 1 if exists, 0 if tombstone        |

Primary key: `(rel_path, device_id)`

### snapshots

| Column      | Type    | Description            |
|-------------|---------|------------------------|
| snapshot_id | TEXT    | Primary key, random ID |
| device_id   | TEXT    | FK to devices          |
| created_at  | TEXT    | ISO 8601 timestamp     |
| file_count  | INTEGER | Number of entries      |

### snapshot_entries

| Column       | Type    | Description               |
|--------------|---------|---------------------------|
| snapshot_id  | TEXT    | FK to snapshots           |
| rel_path     | TEXT    | Relative file path        |
| content_hash | TEXT    | SHA-256 hash              |
| size         | INTEGER | File size in bytes        |
| mod_time     | TEXT    | Modification time         |
| is_dir       | INTEGER | Directory flag            |
| exists_flag  | INTEGER | Existence flag            |

Primary key: `(snapshot_id, rel_path)`

### conflicts

| Column        | Type    | Description                    |
|---------------|---------|--------------------------------|
| id            | INTEGER | Auto-increment primary key     |
| rel_path      | TEXT    | Conflicting file path          |
| local_hash    | TEXT    | Local content hash             |
| remote_hash   | TEXT    | Remote content hash            |
| local_device  | TEXT    | Local device ID                |
| remote_device | TEXT    | Remote device ID               |
| detected_at   | TEXT    | Detection timestamp            |
| resolved      | INTEGER | 0=unresolved, 1=resolved       |
| resolution    | TEXT    | Resolution strategy if resolved|
| kind          | TEXT    | content/delete_modify/create_create |

### tombstones

| Column     | Type | Description           |
|------------|------|-----------------------|
| rel_path   | TEXT | Deleted file path     |
| deleted_by | TEXT | Device that deleted   |
| deleted_at | TEXT | Deletion timestamp    |
| sync_id    | TEXT | Associated sync       |

Primary key: `(rel_path, deleted_by)`

### config

| Column | Type | Description        |
|--------|------|--------------------|
| key    | TEXT | Configuration key  |
| value  | TEXT | Configuration value|

### sync_log

| Column         | Type    | Description           |
|----------------|---------|-----------------------|
| sync_id        | TEXT    | Primary key           |
| device_id      | TEXT    | FK to devices         |
| created_at     | TEXT    | Timestamp             |
| parent_sync_id | TEXT    | Previous sync session |
| file_count     | INTEGER | Files in session      |
| direction      | TEXT    | push/pull             |

## Indexes

- `idx_files_state` — on `files(state)`
- `idx_files_device` — on `files(device_id)`
- `idx_conflicts_unresolved` — partial index on `conflicts(resolved) WHERE resolved = 0`
- `idx_tombstones_path` — on `tombstones(rel_path)`
- `idx_snapshot_entries_snap` — on `snapshot_entries(snapshot_id)`

## Content Addressing

Files are stored in a blob store at `<sync_root>/.sneakernet/blobs/` using a
two-level directory structure: `<first_2_hex_chars>/<remaining_hex_chars>`.

Example: hash `abc123...` is stored at `blobs/ab/c123...`

## Metadata Scope in MVP

Supported:
- File content (via SHA-256 hash)
- File size
- Modification time
- Directory tracking
- Existence/deletion tracking

Not supported in MVP:
- File permissions/ownership
- Symlinks
- Extended attributes
- File type (binary vs. text) distinction
